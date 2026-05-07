package envreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMapSupportsExportQuotedAndComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := `export API_KEY="foo bar"
DB_HOST=localhost # local db
PASSWORD='secret'
EMPTY=
MULTI="line1\
line2"
# comment line
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	m, err := LoadMap(path)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"API_KEY":  "foo bar",
		"DB_HOST":  "localhost",
		"PASSWORD": "secret",
		"EMPTY":    "",
		"MULTI":    "line1line2",
	}

	for k, v := range want {
		if got := m[k]; got != v {
			t.Fatalf("expected %s=%q, got %q", k, v, got)
		}
	}

	if len(m) != len(want) {
		t.Fatalf("expected %d keys, got %d", len(want), len(m))
	}
}

func TestLoadKeysReadsAllKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := `ONE=1
TWO=2
EXPORT=exported
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	keys, err := LoadKeys(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	expected := map[string]bool{"ONE": true, "TWO": true, "EXPORT": true}
	for _, k := range keys {
		if !expected[k] {
			t.Fatalf("unexpected key %q", k)
		}
	}
}
