package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Hossiy21/taso/internal/envreader"
	"github.com/Hossiy21/taso/internal/scanner"
	"github.com/Hossiy21/taso/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ghostDir      string
	ghostEnvFiles []string
	ghostJSON     bool
	ghostFix      bool
)

var ghostCmd = &cobra.Command{
	Use:   "ghost",
	Short: "Find env vars used in code but missing from your .env files",
	Long: `Scans your source code (Go, JS, Python, Rust) and finds every
env var your code calls — then cross-checks against your .env files.

Examples:
  taso ghost
  taso ghost --dir ./src --env .env.production
  taso ghost --env .env --env .env.local --json`,
	RunE: runGhost,
}

func init() {
	ghostCmd.Flags().StringVar(&ghostDir, "dir", ".", "Directory to scan for source files")
	ghostCmd.Flags().StringArrayVar(&ghostEnvFiles, "env", nil, "Env files to check against (auto-detected if not set)")
	ghostCmd.Flags().BoolVar(&ghostJSON, "json", false, "Output results as JSON")
	ghostCmd.Flags().BoolVar(&ghostFix, "fix", false, "Auto-add missing variables to your .env file")
}

func runGhost(cmd *cobra.Command, args []string) error {
	if len(ghostEnvFiles) == 0 {
		ghostEnvFiles = findAllEnvFiles(ghostDir)
	}

	knownKeys := map[string]bool{}
	loadedFiles := []string{}
	for _, ef := range ghostEnvFiles {
		keys, err := envreader.LoadKeys(ef)
		if err != nil {
			continue
		}
		for _, k := range keys {
			knownKeys[k] = true
		}
		loadedFiles = append(loadedFiles, ef)
	}

	findings, err := scanner.ScanDir(ghostDir, viper.GetStringSlice("ignored_dirs"))
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	ghosts := buildGhosts(findings, knownKeys)

	if ghostFix && len(ghosts) > 0 {
		err := autoFixEnv(ghosts, loadedFiles, ghostDir)
		if err != nil {
			return fmt.Errorf("auto-fix failed: %w", err)
		}
		// Refresh knownKeys after fix so we don't print them as ghosts anymore
		for k := range ghosts {
			if k != "__DYNAMIC_ENV_USAGE__" {
				knownKeys[k] = true
			}
		}
		// Also "fix" aliases by not reporting them as ghosts, just keep them for warning
		if _, ok := ghosts["__ALIAS_DETECTION__"]; ok {
			knownKeys["__ALIAS_DETECTION__"] = true
		}
		ghosts = buildGhosts(findings, knownKeys)
	}

	if ghostJSON {
		return printGhostJSON(ghosts, loadedFiles)
	}
	return printGhostHuman(ghosts, findings, loadedFiles)
}

func autoFixEnv(ghosts map[string][]scanner.Usage, loadedFiles []string, dir string) error {
	targetFile := filepath.Join(dir, ".env")
	if len(loadedFiles) > 0 {
		targetFile = loadedFiles[0]
	}

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write a newline just in case the file didn't end with one
	f.WriteString("\n")

	keys := sortedKeys(ghosts)
	for _, k := range keys {
		if k == "__DYNAMIC_ENV_USAGE__" {
			continue
		}
		f.WriteString(fmt.Sprintf("%s=\n", k))
	}
	return nil
}

func printGhostHuman(ghosts map[string][]scanner.Usage, findings map[string][]scanner.Usage, loadedFiles []string) error {
	r := ui.NewRenderer()

	r.Println(ui.Dim("  Scanning: " + ghostDir))

	if len(loadedFiles) == 0 {
		r.Println(ui.Warn2("  No .env files found — showing all env vars detected in code"))
	} else {
		r.Println(ui.Dim("  Env files: " + strings.Join(loadedFiles, ", ")))
	}
	r.Println(ui.Dim(fmt.Sprintf("  %d env var(s) detected in source", len(findings))))
	r.Println("")

	if len(findings) == 0 {
		r.Println(ui.Dim("  No env var usage found in source code."))
		r.Println(ui.Dim("  Tip: taso ghost --dir C:\\Projects\\myapp"))
		fmt.Print(r.String())
		return nil
	}

	// No env files — just show everything found in code
	if len(loadedFiles) == 0 && !ghostFix {
		names := sortedKeys(findings)
		r.Println(ui.Bold(fmt.Sprintf("  %d env var(s) used in your code:", len(names))))
		r.Println("")
		for _, name := range names {
			if name == "__DYNAMIC_ENV_USAGE__" {
				continue
			}
			r.Println(ui.Bold("  " + name))
			for _, u := range findings[name] {
				r.Println(ui.Dim(fmt.Sprintf("    %s:%d", u.File, u.Line)))
			}
			r.Println("")
		}
		r.Println(ui.Warn2("  Add a .env file so taso can check which ones are missing."))
		fmt.Print(r.String())
		return nil
	}

	if len(ghosts) == 0 {
		r.Println(ui.Success("  ✔  No ghost variables found — your env is clean!"))
		fmt.Print(r.String())
		return nil
	}

	names := sortedKeys(ghosts)
	r.Println(ui.Danger(fmt.Sprintf("  👻  %d ghost variable(s) found", len(ghosts))))
	r.Println("")

	for _, name := range names {
		if name == "__DYNAMIC_ENV_USAGE__" {
			r.Println(ui.Warn2("  ⚠️  Dynamic Environment Variables Detected"))
			for _, u := range ghosts[name] {
				r.Println(ui.Dim(fmt.Sprintf("    used in:  %s:%d (Cannot auto-resolve)", u.File, u.Line)))
			}
			r.Println("")
			continue
		}

		if name == "__ALIAS_DETECTION__" {
			r.Println(ui.Warn2("  ⚠️  Environment Alias Detected"))
			for _, u := range ghosts[name] {
				r.Println(ui.Dim(fmt.Sprintf("    used in:  %s:%d (Tracking limited)", u.File, u.Line)))
			}
			r.Println("")
			continue
		}

		usages := ghosts[name]
		r.Println(ui.Bold("  " + name))
		for _, u := range usages {
			r.Println(ui.Dim(fmt.Sprintf("    used in:  %s:%d", u.File, u.Line)))
		}
		r.Println(ui.Warn2(fmt.Sprintf("    not in:   %s", strings.Join(loadedFiles, ", "))))
		r.Println("")
	}

	r.Println(ui.Dim("  Run 'taso score' to see your full env health score."))
	fmt.Print(r.String())
	return nil
}

func printGhostJSON(ghosts map[string][]scanner.Usage, loadedFiles []string) error {
	fmt.Println("{")
	fmt.Printf("  \"ghost_count\": %d,\n", len(ghosts))
	fmt.Printf("  \"checked_files\": %s,\n", jsonStringArray(loadedFiles))
	fmt.Println("  \"ghosts\": [")
	names := sortedKeys(ghosts)
	for i, name := range names {
		comma := ","
		if i == len(names)-1 {
			comma = ""
		}
		usages := ghosts[name]
		fmt.Printf("    {\"var\": %q, \"usages\": [", name)
		for j, u := range usages {
			uc := ","
			if j == len(usages)-1 {
				uc = ""
			}
			fmt.Printf("{\"file\": %q, \"line\": %d}%s", u.File, u.Line, uc)
		}
		fmt.Printf("]}%s\n", comma)
	}
	fmt.Println("  ]")
	fmt.Println("}")
	return nil
}

func buildGhosts(findings map[string][]scanner.Usage, knownKeys map[string]bool) map[string][]scanner.Usage {
	ghosts := map[string][]scanner.Usage{}
	for varName, usages := range findings {
		if !knownKeys[varName] {
			ghosts[varName] = usages
		}
	}
	return ghosts
}

func findAllEnvFiles(root string) []string {
	found := []string{}
	candidates := []string{
		".env", ".env.local", ".env.development",
		".env.production", ".env.staging", ".env.example",
		".env.test", ".env.prod", ".env.dev",
	}
	for _, c := range candidates {
		p := filepath.Join(root, c)
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipEnvDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".env") {
			for _, f := range found {
				if f == path {
					return nil
				}
			}
			found = append(found, path)
		}
		return nil
	})
	return found
}

func shouldSkipEnvDir(name string) bool {
	skip := map[string]bool{
		"node_modules": true, ".git": true, "dist": true,
		"build": true, ".next": true, "target": true,
		"__pycache__": true, ".venv": true, "venv": true,
		"vendor": true, ".cache": true, "coverage": true,
		"bin": true, "obj": true,
	}
	return skip[name]
}

func autoDetectEnvFiles(dir string) []string {
	return findAllEnvFiles(dir)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func jsonStringArray(ss []string) string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
