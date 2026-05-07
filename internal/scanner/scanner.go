package scanner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var dynamicPattern = regexp.MustCompile(`process\.env\[[^'"`+"`"+`"\n]+\]`)

// Usage records where an env var was found in source code
type Usage struct {
	File string
	Line int
	Lang string
}

// ScanDir walks a directory and returns a map of env var name -> []Usage
func ScanDir(root string, extraIgnores []string) (map[string][]Usage, error) {
	results := map[string][]Usage{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			if shouldSkipDir(info.Name(), extraIgnores) {
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
func shouldSkipDir(name string, extraIgnores []string) bool {
	skip := map[string]bool{
		"vendor": true, "node_modules": true, ".git": true,
		"dist": true, "build": true, ".next": true, "target": true,
		"__pycache__": true, ".venv": true, "venv": true,
		"env": true, ".cache": true, "coverage": true,
	}
	if skip[name] {
		return true
	}
	for _, ign := range extraIgnores {
		if ign == name {
			return true
		}
	}
	return false
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
						key := strings.Trim(lit.Value, `"`+"`")
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
	regexp.MustCompile(`process\.env\[\s*['"]([A-Z][A-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`process\.env\?\.([A-Z][A-Z0-9_]*)`),
	regexp.MustCompile(`process\.env\?\.\[\s*['"]([A-Z][A-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`import\.meta\.env\.([A-Z][A-Z0-9_]*)`),
	regexp.MustCompile(`import\.meta\.env\[\s*['"]([A-Z][A-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`import\.meta\.env\?\.\[\s*['"]([A-Z][A-Z0-9_]*)['"]\s*\]`),
}

func scanJS(path string) map[string][]Usage {
	return scanWithPatterns(path, "js", jsPatterns)
}

// ── Python scanner ─────────────────────────────────────────────────────────────
// Catches: os.environ["KEY"], os.environ.get("KEY"), os.getenv("KEY")

var pyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.environ\[['"]([A-Z][A-Z0-9_]*)['"]\]`),
	regexp.MustCompile(`os\.environ\.get\(\s*['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.environ\.setdefault\(\s*['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.getenv\(\s*['"]([A-Z][A-Z0-9_]*)['"]`),
	regexp.MustCompile(`environ\.get\(\s*['"]([A-Z][A-Z0-9_]*)['"]`),
}

func scanPython(path string) map[string][]Usage {
	return scanWithPatterns(path, "python", pyPatterns)
}

// ── Rust scanner ───────────────────────────────────────────────────────────────
// Catches: env::var("KEY"), std::env::var("KEY"), env!("KEY")

var rustPatterns = []*regexp.Regexp{
	regexp.MustCompile(`env::var\(\s*["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`std::env::var\(\s*["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`env!\(\s*["']([A-Z][A-Z0-9_]*)["']`),
	regexp.MustCompile(`option_env!\(\s*["']([A-Z][A-Z0-9_]*)["']`),
}

func scanRust(path string) map[string][]Usage {
	return scanWithPatterns(path, "rust", rustPatterns)
}

// ── Ruby scanner ───────────────────────────────────────────────────────────────
// Catches: ENV["KEY"], ENV.fetch("KEY"), ENV['KEY']

var rubyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`ENV\[\s*["']([A-Z][A-Z0-9_]*)["']\s*\]`),
	regexp.MustCompile(`ENV\.fetch\(\s*["']([A-Z][A-Z0-9_]*)["']`),
}

func scanRuby(path string) map[string][]Usage {
	return scanWithPatterns(path, "ruby", rubyPatterns)
}

// ── Java scanner ───────────────────────────────────────────────────────────────
// Catches: System.getenv("KEY")

var javaPatterns = []*regexp.Regexp{
	regexp.MustCompile(`System\.getenv\(\s*["']([A-Z][A-Z0-9_]*)["']`),
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
		// Skip minified lines (O(N*M) vulnerability)
		if len(line) > 1000 {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Look for obvious dynamic accesses as a heuristic
		if strings.Contains(line, "process.env[") {
			// This regex tries to find process.env[something] where something doesn't look like a literal string
			// It's a heuristic and not perfect, but it helps.
			// Handled outside regex array for now as a special warning?
			// Actually, let's just use a special key for dynamic usage.
			if dynamicPattern.MatchString(line) {
				results["__DYNAMIC_ENV_USAGE__"] = append(results["__DYNAMIC_ENV_USAGE__"], Usage{
					File: path,
					Line: lineNum + 1,
					Lang: lang,
				})
			}
		}

		line = stripInlineComments(line)
		line = stripQuotedStrings(line)

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

func stripInlineComments(line string) string {
	line = stripBlockComments(line)

	var out strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for i, ch := range line {
		if escaped {
			out.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			out.WriteRune(ch)
			continue
		}

		if ch == '"' && !inSingle {
			inDouble = !inDouble
			out.WriteRune(ch)
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			out.WriteRune(ch)
			continue
		}

		if !inSingle && !inDouble {
			if ch == '/' && i+1 < len(line) && line[i+1] == '/' {
				if i == 0 || unicode.IsSpace(rune(line[i-1])) {
					break
				}
			}
			if ch == '#' {
				if i == 0 || unicode.IsSpace(rune(line[i-1])) {
					break
				}
			}
		}

		out.WriteRune(ch)
	}

	return out.String()
}

func stripQuotedStrings(line string) string {
	var out strings.Builder
	inSingle := false
	inDouble := false
	inBacktick := false
	preserveQuoted := false
	escaped := false

	for i, ch := range line {
		if escaped {
			if inSingle || inDouble || inBacktick {
				if preserveQuoted {
					out.WriteRune(ch)
				} else {
					out.WriteRune(' ')
				}
			} else {
				out.WriteRune(ch)
			}
			escaped = false
			continue
		}

		if ch == '\\' && (inSingle || inDouble || inBacktick) {
			escaped = true
			if preserveQuoted {
				out.WriteRune(' ')
			} else {
				out.WriteRune(' ')
			}
			continue
		}

		if ch == '"' && !inSingle && !inBacktick {
			if inDouble {
				if preserveQuoted {
					out.WriteRune(ch)
				} else {
					out.WriteRune(' ')
				}
				inDouble = false
				preserveQuoted = false
				continue
			}

			preserveQuoted = isEnvStringLiteral(line, i)
			inDouble = true
			if preserveQuoted {
				out.WriteRune(ch)
			} else {
				out.WriteRune(' ')
			}
			continue
		}

		if ch == '\'' && !inDouble && !inBacktick {
			if inSingle {
				if preserveQuoted {
					out.WriteRune(ch)
				} else {
					out.WriteRune(' ')
				}
				inSingle = false
				preserveQuoted = false
				continue
			}

			preserveQuoted = isEnvStringLiteral(line, i)
			inSingle = true
			if preserveQuoted {
				out.WriteRune(ch)
			} else {
				out.WriteRune(' ')
			}
			continue
		}

		if inSingle || inDouble || inBacktick {
			if preserveQuoted {
				out.WriteRune(ch)
			} else {
				out.WriteRune(' ')
			}
			if (ch == '"' && inDouble) || (ch == '\'' && inSingle) {
				preserveQuoted = false
			}
			continue
		}

		out.WriteRune(ch)
	}

	return out.String()
}

func isEnvStringLiteral(line string, index int) bool {
	if isQuoteAfterOpenBracket(line, index) {
		return true
	}

	prefix := strings.TrimRightFunc(line[:index], unicode.IsSpace)
	envCallers := []string{
		"os.getenv(",
		"os.environ.get(",
		"os.environ.setdefault(",
		"environ.get(",
		"System.getenv(",
		"env::var(",
		"std::env::var(",
		"env!(",
		"option_env!(",
		"ENV.fetch(",
	}
	for _, c := range envCallers {
		if strings.HasSuffix(prefix, c) {
			return true
		}
	}
	return false
}

func isQuoteAfterOpenBracket(line string, index int) bool {
	for j := index - 1; j >= 0; j-- {
		if unicode.IsSpace(rune(line[j])) {
			continue
		}
		return line[j] == '['
	}
	return false
}

func stripBlockComments(line string) string {
	for {
		start := strings.Index(line, "/*")
		if start < 0 {
			return line
		}
		end := strings.Index(line[start+2:], "*/")
		if end < 0 {
			return line[:start]
		}
		line = line[:start] + line[start+2+end+2:]
	}
}
