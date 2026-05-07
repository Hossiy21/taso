package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hossiy21/taso/internal/scanner"
)

func TestAutoFixEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	
	// Start with an empty .env
	if err := os.WriteFile(envPath, []byte("EXISTING_VAR=val"), 0o600); err != nil {
		t.Fatal(err)
	}

	ghosts := map[string][]scanner.Usage{
		"MISSING_VAR": {{File: "app.js", Line: 10}},
	}

	err := autoFixEnv(ghosts, []string{envPath}, dir)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "MISSING_VAR=") {
		t.Errorf("expected MISSING_VAR= to be appended to .env, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "EXISTING_VAR=val") {
		t.Error("existing content was lost")
	}
}
