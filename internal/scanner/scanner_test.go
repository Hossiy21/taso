package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanJSDetectsEnvPatterns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.js")
	content := `
const publicUrl = process.env.PUBLIC_URL;
const token = process.env?.["SECRET_TOKEN"];
const viteUrl = import.meta.env.VITE_API_URL;
const opt = import.meta.env?.["ANOTHER_KEY"];
// process.env.IGNORED_IN_COMMENT
const text = "process.env.FAKE_KEY";
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	result := scanJS(path)
	keys := map[string]bool{}
	for k := range result {
		keys[k] = true
	}

	want := []string{"PUBLIC_URL", "SECRET_TOKEN", "VITE_API_URL", "ANOTHER_KEY"}
	for _, key := range want {
		if !keys[key] {
			t.Fatalf("expected key %q to be detected", key)
		}
	}

	if keys["FAKE_KEY"] {
		t.Fatal("unexpected key detected from string literal")
	}

	if keys["IGNORED_IN_COMMENT"] {
		t.Fatal("unexpected key detected from comment")
	}
}

func TestScanPythonDetectsEnvPatterns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.py")
	content := `
import os

api = os.environ["API_KEY"]
secret = os.getenv('SECRET_KEY')
defaulted = os.environ.get("OPTIONAL_KEY", "default")
setdef = os.environ.setdefault('SETME', 'default')
# os.environ["IGNORED"]
text = "os.getenv('FAKE')"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	result := scanPython(path)
	keys := map[string]bool{}
	for k := range result {
		keys[k] = true
	}

	for _, key := range []string{"API_KEY", "SECRET_KEY", "OPTIONAL_KEY", "SETME"} {
		if !keys[key] {
			t.Fatalf("expected key %q to be detected", key)
		}
	}

	if keys["IGNORED"] {
		t.Fatal("unexpected key detected from comment")
	}

	if keys["FAKE"] {
		t.Fatal("unexpected key detected from string literal")
	}
}
