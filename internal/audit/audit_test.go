package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAuditLogCreation verifies audit logs are created properly
func TestAuditLogCreation(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "logs")

	logger, err := NewLogger(logDir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("logger should not be nil")
	}

	// Verify log directory was created
	if _, err := os.Stat(logDir); err != nil {
		t.Fatalf("log directory not created: %v", err)
	}
}

// TestAuditLogEntry verifies entries are logged correctly
func TestAuditLogEntry(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	entry := Entry{
		Timestamp:    time.Now(),
		User:         "testuser",
		Command:      "ghost",
		ScanDir:      "/test/dir",
		EnvFiles:     []string{".env"},
		VarsFound:    5,
		FilesScanned: 100,
		Duration:     "1.5s",
		Status:       "success",
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("failed to log entry: %v", err)
	}

	// Verify log file was created
	logFiles, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read log directory: %v", err)
	}

	if len(logFiles) == 0 {
		t.Fatal("no log files created")
	}

	// Read log file and verify content
	logFile := filepath.Join(dir, logFiles[0].Name())
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ghost") {
		t.Fatal("log content does not contain command name")
	}
	if !strings.Contains(logContent, "testuser") {
		t.Fatal("log content does not contain user")
	}
	if !strings.Contains(logContent, "success") {
		t.Fatal("log content does not contain status")
	}
}

// TestErrorAuditEntry verifies error entries are logged correctly
func TestErrorAuditEntry(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	entry := BuildErrorEntry("ghost", "/test/dir", []string{".env"}, "test error")

	if err := logger.Log(entry); err != nil {
		t.Fatalf("failed to log error entry: %v", err)
	}

	// Read log file and verify error
	logFiles, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read log directory: %v", err)
	}

	if len(logFiles) == 0 {
		t.Fatal("no log files created")
	}

	logFile := filepath.Join(dir, logFiles[0].Name())
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "error") {
		t.Fatal("log content does not contain error status")
	}
	if !strings.Contains(logContent, "test error") {
		t.Fatal("log content does not contain error message")
	}
}

// TestGetUser verifies user detection works
func TestGetUser(t *testing.T) {
	user := GetUser()
	if user == "" {
		t.Fatal("GetUser returned empty string")
	}
	if user == "unknown" {
		t.Log("GetUser returned 'unknown' - environment variables not set")
	}
}

// TestBuildEntry verifies BuildEntry creates correct entries
func TestBuildEntry(t *testing.T) {
	entry := BuildEntry("test_cmd", "/test/dir", []string{".env"}, 10, 100, 5, 2*time.Second, "success")

	if entry.Command != "test_cmd" {
		t.Fatal("command not set correctly")
	}
	if entry.VarsFound != 10 {
		t.Fatal("vars_found not set correctly")
	}
	if entry.FilesScanned != 100 {
		t.Fatal("files_scanned not set correctly")
	}
	if entry.FilesSkipped != 5 {
		t.Fatal("files_skipped not set correctly")
	}
	if entry.Status != "success" {
		t.Fatal("status not set correctly")
	}
}
