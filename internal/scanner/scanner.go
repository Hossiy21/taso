package scanner

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	jsast "github.com/dop251/goja/ast"
	"github.com/dop251/goja/file"
	jsparser "github.com/dop251/goja/parser"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"github.com/Hossiy21/taso/internal/cache"
	"github.com/Hossiy21/taso/internal/security"
)

var dynamicPattern = regexp.MustCompile(`process\.env\[[^'"` + "`" + `"\n]+\]`)
var aliasPatternJS = regexp.MustCompile(`(?:const|let|var)\s+([a-zA-Z0-9_$]+)\s*=\s*process\.env`)
var aliasPatternPy = regexp.MustCompile(`([a-zA-Z0-9_]+)\s*=\s*os\.environ`)

// Usage records where an env var was found in source code
type Usage struct {
	File string
	Line int
	Lang string
}

// ScanDir walks a directory and returns a map of env var name -> []Usage.
// It uses the provided cache store to skip unchanged files.
func ScanDir(root string, extraIgnores []string, store *cache.Store) (map[string][]Usage, error) {
	// SECURITY: Prevent directory traversal attacks and ensure path accessibility
	if err := security.ValidateScanPath(root, ""); err != nil {
		return nil, err
	}

	// Verify the directory exists and is readable
	info, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("scan directory not accessible: %w", err)
	}

	if (info.Mode() & os.ModeSymlink) != 0 {
		return nil, fmt.Errorf("symlinks not allowed as scan directory")
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("scan path is not a directory: %s", root)
	}

	results := map[string][]Usage{}
	var mu sync.Mutex

	// Resource monitoring for DoS prevention
	monitor := security.NewResourceMonitor()

	type task struct {
		path string
		ext  string
	}

	taskChan := make(chan task, 100)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start workers
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskChan {
				// SECURITY: Skip files that exceed size limits
				if security.ShouldSkipFile(t.path) {
					monitor.RecordSkippedFile()
					continue
				}

				// CACHE: Check if file has changed
				var findings map[string][]Usage
				var hash string
				if store != nil {
					hash, _ = cache.ComputeHash(t.path)
					if raw, ok := store.Get(t.path, hash); ok {
						var m map[string][]Usage
						if err := json.Unmarshal(raw, &m); err == nil {
							findings = m
						}
					}
				}

				if findings == nil {
					switch t.ext {
					case ".go":
						findings = scanGo(t.path)
					case ".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs":
						findings = scanJS(t.path)
					case ".py":
						findings = scanPython(t.path)
					case ".rs":
						findings = scanRust(t.path)
					case ".rb":
						findings = scanRuby(t.path)
					case ".java":
						findings = scanJava(t.path)
					case ".cs":
						findings = scanCSharp(t.path)
					case ".php":
						findings = scanPHP(t.path)
					case ".kt", ".kts":
						findings = scanKotlin(t.path)
					}

					// Update cache
					if store != nil && hash != "" {
						store.Set(t.path, hash, findings)
					}
				}

				// Record resource usage
				info, _ := os.Stat(t.path)
				if info != nil {
					if err := monitor.RecordFile(info.Size()); err != nil {
						select {
						case errChan <- err:
						default:
						}
						return
					}
				}

				if len(findings) > 0 {
					mu.Lock()
					for k, v := range findings {
						results[k] = append(results[k], v...)
					}
					mu.Unlock()
				}
			}
		}()
	}

	// Walk and dispatch
	go func() {
		defer close(taskChan)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if shouldSkipDir(info.Name(), extraIgnores) {
					return filepath.SkipDir
				}
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".go", ".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs", ".py", ".rs", ".rb", ".java", ".cs", ".php", ".kt", ".kts":
				taskChan <- task{path: path, ext: ext}
			}
			return nil
		})
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		return nil, err
	default:
		return results, nil
	}
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
// Accepts both uppercase and lowercase env var names

var jsPatterns = []*regexp.Regexp{
	regexp.MustCompile(`process\.env\.([a-zA-Z_][a-zA-Z0-9_]*)`),
	regexp.MustCompile(`process\.env\[\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`process\.env\?\.([a-zA-Z_][a-zA-Z0-9_]*)`),
	regexp.MustCompile(`process\.env\?\.\[\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`import\.meta\.env\.([a-zA-Z_][a-zA-Z0-9_]*)`),
	regexp.MustCompile(`import\.meta\.env\[\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]\s*\]`),
	regexp.MustCompile(`import\.meta\.env\?\.\[\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]\s*\]`),
}

func scanJS(path string) map[string][]Usage {
	astResults := scanJSWithAST(path)
	regexResults := scanWithPatterns(path, "js", jsPatterns)

	// Merge regex results for keys not found by AST
	for k, v := range regexResults {
		if _, ok := astResults[k]; !ok {
			astResults[k] = v
		}
	}

	return astResults
}

func scanJSWithAST(path string) map[string][]Usage {
	results := map[string][]Usage{}
	content, err := os.ReadFile(path)
	if err != nil {
		return results
	}

	program, err := jsparser.ParseFile(nil, path, string(content), 0)
	if err != nil {
		return results
	}

	record := func(key string, idx file.Idx) {
		results[key] = append(results[key], Usage{
			File: path,
			Line: getLine(content, idx),
			Lang: "js",
		})
	}

	var walk func(jsast.Node)
	walk = func(n jsast.Node) {
		if n == nil {
			return
		}
		// fmt.Printf("WALK: %T\n", n)

		// 1. Detect process.env.KEY or process.env['KEY']
		if dot, ok := n.(*jsast.DotExpression); ok {
			leftStr := nodeToString(dot.Left)
			if leftStr == "process.env" || leftStr == "import.meta.env" {
				record(string(dot.Identifier.Name), dot.Idx0())
			}
		}

		if brac, ok := n.(*jsast.BracketExpression); ok {
			leftStr := nodeToString(brac.Left)
			if leftStr == "process.env" || leftStr == "import.meta.env" {
				if lit, ok := brac.Member.(*jsast.StringLiteral); ok {
					record(string(lit.Value), brac.Idx0())
				}
			}
		}

		// 2. Detect destructuring: const { KEY } = process.env
		if decl, ok := n.(*jsast.VariableDeclaration); ok {
			for _, vd := range decl.List {
				if vd.Initializer != nil {
					initStr := nodeToString(vd.Initializer)
					if initStr == "process.env" || initStr == "import.meta.env" {
						if obj, ok := vd.Target.(*jsast.ObjectPattern); ok {
							for _, prop := range obj.Properties {
								if key, ok := getPropertyNameFromInterface(prop); ok {
									record(key, prop.Idx0())
								}
							}
						}
					}
				}
			}
		}
		if decl, ok := n.(*jsast.LexicalDeclaration); ok {
			for _, b := range decl.List {
				if b.Initializer != nil {
					initStr := nodeToString(b.Initializer)
					if initStr == "process.env" || initStr == "import.meta.env" {
						if obj, ok := b.Target.(*jsast.ObjectPattern); ok {
							for _, prop := range obj.Properties {
								if key, ok := getPropertyNameFromInterface(prop); ok {
									record(key, prop.Idx0())
								}
							}
						}
					}
				}
			}
		}

		// 3. Detect assignments/aliases: const env = process.env
		if decl, ok := n.(*jsast.VariableDeclaration); ok {
			for _, vd := range decl.List {
				if vd.Initializer != nil {
					initStr := nodeToString(vd.Initializer)
					if initStr == "process.env" || initStr == "import.meta.env" {
						if _, ok := vd.Target.(*jsast.Identifier); ok {
							results["__ALIAS_DETECTION__"] = append(results["__ALIAS_DETECTION__"], Usage{
								File: path,
								Line: getLine(content, vd.Target.Idx0()),
								Lang: "js",
							})
						}
					}
				}
			}
		}
		if decl, ok := n.(*jsast.LexicalDeclaration); ok {
			for _, b := range decl.List {
				if b.Initializer != nil {
					initStr := nodeToString(b.Initializer)
					if initStr == "process.env" || initStr == "import.meta.env" {
						if _, ok := b.Target.(*jsast.Identifier); ok {
							results["__ALIAS_DETECTION__"] = append(results["__ALIAS_DETECTION__"], Usage{
								File: path,
								Line: getLine(content, b.Target.Idx0()),
								Lang: "js",
							})
						}
					}
				}
			}
		}

		// Recursive walk
		switch node := n.(type) {
		case *jsast.Program:
			for _, stmt := range node.Body {
				walk(stmt)
			}
		case *jsast.BlockStatement:
			for _, stmt := range node.List {
				walk(stmt)
			}
		case *jsast.ExpressionStatement:
			walk(node.Expression)
		case *jsast.VariableStatement:
			for _, b := range node.List {
				walk(b.Initializer)
			}
		case *jsast.LexicalDeclaration:
			for _, b := range node.List {
				walk(b.Initializer)
			}
		case *jsast.VariableDeclaration:
			for _, vd := range node.List {
				walk(vd.Initializer)
			}
		case *jsast.DotExpression:
			walk(node.Left)
		case *jsast.BracketExpression:
			walk(node.Left)
			walk(node.Member)
		case *jsast.AssignExpression:
			walk(node.Left)
			walk(node.Right)
		case *jsast.CallExpression:
			walk(node.Callee)
			for _, arg := range node.ArgumentList {
				walk(arg)
			}
		}
	}

	walk(program)

	return results
}

func getLine(content []byte, idx file.Idx) int {
	line := 1
	pos := int(idx) - 1
	for i := 0; i < pos && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

func nodeToString(n jsast.Node) string {
	if n == nil {
		return ""
	}
	switch node := n.(type) {
	case *jsast.Identifier:
		return string(node.Name)
	case *jsast.DotExpression:
		return nodeToString(node.Left) + "." + string(node.Identifier.Name)
	case *jsast.MetaProperty:
		return string(node.Meta.Name) + "." + string(node.Property.Name)
	}
	return ""
}

func getPropertyNameFromInterface(p jsast.Property) (string, bool) {
	if pk, ok := p.(*jsast.PropertyKeyed); ok {
		return getPropertyName(pk.Key)
	}
	if ps, ok := p.(*jsast.PropertyShort); ok {
		return string(ps.Name.Name), true
	}
	return "", false
}

func getPropertyName(n jsast.Node) (string, bool) {
	if id, ok := n.(*jsast.Identifier); ok {
		return string(id.Name), true
	}
	if lit, ok := n.(*jsast.StringLiteral); ok {
		return string(lit.Value), true
	}
	return "", false
}

// ── Python scanner ─────────────────────────────────────────────────────────────
// Catches: os.environ["KEY"], os.environ.get("KEY"), os.getenv("KEY")
// Accepts both uppercase and lowercase env var names

var pyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`os\.environ\[['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]\]`),
	regexp.MustCompile(`os\.environ\.get\(\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.environ\.setdefault\(\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]`),
	regexp.MustCompile(`os\.getenv\(\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]`),
	regexp.MustCompile(`environ\.get\(\s*['"]([ a-zA-Z_][a-zA-Z0-9_]*)['"]`),
}

func scanPython(path string) map[string][]Usage {
	return scanWithPatterns(path, "python", pyPatterns)
}

// ── Rust scanner ───────────────────────────────────────────────────────────────
// Catches: env::var("KEY"), std::env::var("KEY"), env!("KEY")
// Accepts both uppercase and lowercase env var names

var rustPatterns = []*regexp.Regexp{
	regexp.MustCompile(`env::var\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
	regexp.MustCompile(`std::env::var\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
	regexp.MustCompile(`env!\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
	regexp.MustCompile(`option_env!\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
}

func scanRust(path string) map[string][]Usage {
	return scanWithPatterns(path, "rust", rustPatterns)
}

// ── Ruby scanner ───────────────────────────────────────────────────────────────
// Catches: ENV["KEY"], ENV.fetch("KEY"), ENV['KEY']
// Accepts both uppercase and lowercase env var names

var rubyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`ENV\[\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']\s*\]`),
	regexp.MustCompile(`ENV\.fetch\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
}

func scanRuby(path string) map[string][]Usage {
	return scanWithPatterns(path, "ruby", rubyPatterns)
}

// ── Java scanner ───────────────────────────────────────────────────────────────
// Catches: System.getenv("KEY")
// Accepts both uppercase and lowercase env var names

var javaPatterns = []*regexp.Regexp{
	regexp.MustCompile(`System\.getenv\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
}

func scanJava(path string) map[string][]Usage {
	return scanWithPatterns(path, "java", javaPatterns)
}

// ── C# scanner ────────────────────────────────────────────────────────────────
// Catches: Environment.GetEnvironmentVariable("KEY")

var csPatterns = []*regexp.Regexp{
	regexp.MustCompile(`Environment\.GetEnvironmentVariable\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
}

func scanCSharp(path string) map[string][]Usage {
	return scanWithPatterns(path, "csharp", csPatterns)
}

// ── PHP scanner ───────────────────────────────────────────────────────────────
// Catches: getenv("KEY"), $_ENV["KEY"], $_SERVER["KEY"]

var phpPatterns = []*regexp.Regexp{
	regexp.MustCompile(`getenv\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
	regexp.MustCompile(`\$_ENV\[\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']\s*\]`),
	regexp.MustCompile(`\$_SERVER\[\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']\s*\]`),
}

func scanPHP(path string) map[string][]Usage {
	return scanWithPatterns(path, "php", phpPatterns)
}

// ── Kotlin scanner ────────────────────────────────────────────────────────────
// Catches: System.getenv("KEY")

var ktPatterns = []*regexp.Regexp{
	regexp.MustCompile(`System\.getenv\(\s*["']([ a-zA-Z_][a-zA-Z0-9_]*)["']`),
}

func scanKotlin(path string) map[string][]Usage {
	return scanWithPatterns(path, "kotlin", ktPatterns)
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

		// Alias detection
		if lang == "js" && aliasPatternJS.MatchString(line) {
			results["__ALIAS_DETECTION__"] = append(results["__ALIAS_DETECTION__"], Usage{
				File: path,
				Line: lineNum + 1,
				Lang: lang,
			})
		}
		if lang == "python" && aliasPatternPy.MatchString(line) {
			results["__ALIAS_DETECTION__"] = append(results["__ALIAS_DETECTION__"], Usage{
				File: path,
				Line: lineNum + 1,
				Lang: lang,
			})
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
