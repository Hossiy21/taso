package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Entry represents a cached scan result for a single file
type Entry struct {
	Hash     string          `json:"hash"`
	Findings json.RawMessage `json:"findings"`
}

// Store manages the persistent cache
type Store struct {
	mu        sync.RWMutex
	filePath  string
	Entries   map[string]Entry `json:"entries"`
	IsDirty   bool             `json:"-"`
}

// NewStore creates a new cache store
func NewStore(dir string) (*Store, error) {
	cachePath := filepath.Join(dir, "cache.json")
	store := &Store{
		filePath: cachePath,
		Entries:  make(map[string]Entry),
	}

	// Try to load existing cache
	data, err := os.ReadFile(cachePath)
	if err == nil {
		_ = json.Unmarshal(data, store)
	}

	return store, nil
}

// Get returns the cached findings if the hash matches
func (s *Store) Get(path string, currentHash string) (json.RawMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.Entries[path]
	if ok && entry.Hash == currentHash {
		return entry.Findings, true
	}
	return nil, false
}

// Set updates the cache for a file
func (s *Store) Set(path string, hash string, findings interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(findings)
	if err != nil {
		return
	}

	s.Entries[path] = Entry{
		Hash:     hash,
		Findings: data,
	}
	s.IsDirty = true
}

// Save writes the cache to disk
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsDirty {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

// ComputeHash calculates the SHA-256 hash of a file
func ComputeHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
