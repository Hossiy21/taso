package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// autoDetectEnvFiles looks for common .env variants in a directory
func autoDetectEnvFiles(dir string) []string {
	variants := []string{".env", ".env.local", ".env.development", ".env.test", ".env.production"}
	found := []string{}
	for _, v := range variants {
		path := filepath.Join(dir, v)
		if _, err := os.Stat(path); err == nil {
			found = append(found, v)
		}
	}
	return found
}

// jsonStringArray converts a slice of strings to a JSON array string
func jsonStringArray(items []string) string {
	if items == nil {
		return "[]"
	}
	data, _ := json.Marshal(items)
	return string(data)
}
