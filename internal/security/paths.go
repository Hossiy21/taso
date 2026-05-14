package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath checks if a path is safe to access without enforcing a base directory.
// It checks for null bytes, cleans the path, and rejects system directories.
func ValidatePath(path string) (string, error) {
	if strings.Contains(path, "\x00") {
		return "", fmt.Errorf("path contains null bytes")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", path, err)
	}

	// Clean the path
	cleaned := filepath.Clean(absPath)

	// Reject absolute paths to system directories
	if isSystemPath(cleaned) {
		return "", fmt.Errorf("path %q points to protected system directory", cleaned)
	}

	return cleaned, nil
}

// ValidateScanPath ensures the requested scan path doesn't use path traversal
// to escape a basePath. If basePath is empty, it just validates the path itself.
func ValidateScanPath(requestedPath, basePath string) error {
	cleanedRequested, err := ValidatePath(requestedPath)
	if err != nil {
		return err
	}

	if basePath == "" || basePath == "." {
		// If no specific base path is enforced (other than current dir),
		// we just ensure it's a valid directory.
		info, err := os.Lstat(cleanedRequested)
		if err != nil {
			return fmt.Errorf("path %q not accessible: %w", cleanedRequested, err)
		}
		if (info.Mode() & os.ModeSymlink) != 0 {
			return fmt.Errorf("symlinks not allowed: %q", cleanedRequested)
		}
		return nil
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("invalid base path %q: %w", basePath, err)
	}

	// Ensure the resolved absolute path doesn't escape the original base
	relPath, err := filepath.Rel(absBase, cleanedRequested)
	if err != nil {
		return fmt.Errorf("path is not relative to base: %q", requestedPath)
	}

	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path %q escapes base directory %q", cleanedRequested, absBase)
	}

	// Ensure path exists and is accessible
	info, err := os.Lstat(cleanedRequested)
	if err != nil {
		return fmt.Errorf("path %q not accessible: %w", cleanedRequested, err)
	}

	if (info.Mode() & os.ModeSymlink) != 0 {
		return fmt.Errorf("symlinks not allowed: %q", cleanedRequested)
	}

	return nil
}

// ValidateEnvFilePath validates an individual .env file path
func ValidateEnvFilePath(filePath string) error {
	// Check for path traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid env file path %q: %w", filePath, err)
	}

	// Reject absolute paths to system directories
	if isSystemPath(absPath) {
		return fmt.Errorf("env file path %q points to protected system directory", absPath)
	}

	// Verify file exists and is readable
	info, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("env file %q not accessible: %w", absPath, err)
	}

	// Reject symlinks
	if (info.Mode() & os.ModeSymlink) != 0 {
		return fmt.Errorf("symlinks not allowed for env files: %q", absPath)
	}

	return nil
}

// isSystemPath checks if a path is in a protected system directory
func isSystemPath(path string) bool {
	systemDirs := []string{
		"/etc",
		"/sys",
		"/proc",
		"/dev",
		"/root",
		"C:\\Windows",
		"C:\\System32",
		"C:\\Program Files",
	}

	pathLower := strings.ToLower(path)
	for _, sysDir := range systemDirs {
		if strings.HasPrefix(pathLower, strings.ToLower(sysDir)) {
			return true
		}
	}
	return false
}

// SanitizePath removes potentially dangerous path components
func SanitizePath(input string) string {
	// Reject null bytes
	if strings.Contains(input, "\x00") {
		return ""
	}

	// Use filepath.Clean to normalize
	cleaned := filepath.Clean(input)

	return cleaned
}
