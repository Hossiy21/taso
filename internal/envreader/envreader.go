package envreader

import (
	"github.com/joho/godotenv"
)

// LoadKeys reads a .env file and returns all key names (no values)
func LoadKeys(path string) ([]string, error) {
	m, err := LoadMap(path)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys, nil
}

// LoadMap reads a .env file and returns a map of key -> value
func LoadMap(path string) (map[string]string, error) {
	return godotenv.Read(path)
}
