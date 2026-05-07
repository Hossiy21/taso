package scanner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Usage records where an env var was found in source code
type Usage struct {
	File string
	Line int
	Lang string
}

// ScanDir walks a directory and returns a map of env var name -> []Usage
func ScanDir(root string) (map[string][]Usage, error) {
	results := map[string][]Usage{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		var findings map[string][]Usage

		switch ext {
		case ".go":
			findings = scanGo(path)
		case ".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs":
			findings = scanJS(path)
		case ".py":
			findings = scanPython(path)
		case ".rs":
			findings = scanRust(path)
		case ".rb":
			findings = scanRuby(path)
		case ".java":
			findings = scanJava(path)
		default:
			return nil
		}

		for k, v := range findings {
			results[k] = append(results[k], v...)
		}
		return nil
	})

	return results, err
}

// shouldSkipDir returns true for dirs that should never be scanned
func shouldSkipDir(name string) bool {
	skip := map[string]bool{
		"vendor": true, "node_modules": true, ".git": true,
		"dist": true, "build": true, ".next": true, "target": true,
		"__pycache__": true, ".venv": true, "venv": true,
		"env": true, ".cache": true, "coverage": true,
	}
	return skip[name]
}

// ── Go AST scanner ─────────────────────────────────────────────────────────────
// Uses the stdlib AST parser for 100% accuracy — no false positives.
// Catches: os.Getenv("KEY"), os.LookupEnv("KEY")

func scanGo(path string) map[string][]Usage {
	results := map[string][]Usage{}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return results
	}

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		fnName := pkg.Name + "." + sel.Sel.Name

		// os.Getenv("VAR") and os.LookupEnv("VAR")
		if fnName == "os.Getenv" || fnName == "os.LookupEnv" {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok {
					if lit.Kind == token.STRING {
						key := strings.Trim(lit.Value, `"` + "`")
						pos := fset.Position(call.Pos())
						results[key] = append(results[key], Usage{
							File: path,
							Line: pos.Line,
							Lang: "go",
						})
					}
				}
			}
		}

		return true
	})

	return results
}

// ── JS / TS scanner ────────────────────────────────────────────────────────────
// Catches: process.env.KEY, process.env["KEY"], process.env?.KEY

var jsPatterns = []*regexp.Regexp{
	regexp.MustCompile(`process\.env\.([A-Z][A-Z0-9_]*)`),
	regexp.MustCompile(`process\.env\[['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`process\.env\?\.([A-Z][A-Z0-9_]*)`),
	regexp.MustCompile(`import\.meta\.env\.([A-Z][A-Z0-9_]*)`), // Vite
}

func scanJS(path string) map[string][]Usage {
	return scanWithPatterns(path, "js", jsPatterns)
}

// ── Python scanner ─────────────────────────────────────────────────────────────
// Catches: os.environ["KEY"], os.environ.get("KEY"), os.getenv("KEY")

var pyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.environ\[['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.environ\.get\(['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.getenv\(['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`environ\.get\(['"]([A-Z][A-Z0-9_]*)['"]`),
}

func scanPython(path string) map[string][]Usage {
	return scanWithPatterns(path, "python", pyPatterns)
}

// ── Rust scanner ───────────────────────────────────────────────────────────────
// Catches: env::var("KEY"), std::env::var("KEY"), env!("KEY")

var rustPatterns = []*regexp.Regexp{
	regexp.MustCompile(`env::var\(["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`std::env::var\(["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`env!\(["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`option_env!\(["']([A-Z][A-Z0-9_]*)["']`),
}

func scanRust(path string) map[string][]Usage {
	return scanWithPatterns(path, "rust", rustPatterns)
}

// ── Ruby scanner ───────────────────────────────────────────────────────────────
// Catches: ENV["KEY"], ENV.fetch("KEY"), ENV['KEY']

var rubyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`ENV\[["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`ENV\.fetch\(["']([A-Z][A-Z0-9_]*)["']`),
}

func scanRuby(path string) map[string][]Usage {
	return scanWithPatterns(path, "ruby", rubyPatterns)
}

// ── Java scanner ───────────────────────────────────────────────────────────────
// Catches: System.getenv("KEY")

var javaPatterns = []*regexp.Regexp{
	regexp.MustCompile(`System\.getenv\(["']([A-Z][A-Z0-9_]*)["']`),
}

func scanJava(path string) map[string][]Usage {
	return scanWithPatterns(path, "java", javaPatterns)
}

// ── Generic regex scanner ──────────────────────────────────────────────────────

func scanWithPatterns(path string, lang string, patterns []*regexp.Regexp) map[string][]Usage {
	results := map[string][]Usage{}

	content, err := os.ReadFile(path)
	if err != nil {
		return results
	}

	lines := strings.Split(string(content), "\n")
	for lineNum, line := range lines {
		// Skip comment lines
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		for _, pat := range patterns {
			matches := pat.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				if len(m) > 1 {
					key := m[1]
					results[key] = append(results[key], Usage{
						File: path,
						Line: lineNum + 1,
						Lang: lang,
					})
				}
			}
		}
	}

	return results
}
