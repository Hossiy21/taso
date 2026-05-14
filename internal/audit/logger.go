package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single audit log entry
type Entry struct {
	Timestamp    time.Time `json:"timestamp"`
	User         string    `json:"user"`
	Command      string    `json:"command"`
	ScanDir      string    `json:"scan_dir"`
	EnvFiles     []string  `json:"env_files"`
	VarsFound    int       `json:"vars_found"`
	FilesScanned int       `json:"files_scanned"`
	FilesSkipped int       `json:"files_skipped"`
	Duration     string    `json:"duration"`
	Status       string    `json:"status"` // success, error, timeout
	ErrorMsg     string    `json:"error_msg,omitempty"`
}

// Logger manages audit log output
type Logger struct {
	filePath string
	entries  []Entry
}

// NewLogger creates a new audit logger
func NewLogger(logDir string) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Use date-based log files (one per day)
	now := time.Now()
	logFile := filepath.Join(logDir, fmt.Sprintf("taso-%04d%02d%02d.log",
		now.Year(), now.Month(), now.Day()))

	return &Logger{
		filePath: logFile,
		entries:  make([]Entry, 0),
	}, nil
}

// Log records an audit entry
func (l *Logger) Log(entry Entry) error {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	l.entries = append(l.entries, entry)

	// Write to file immediately (unbuffered for safety)
	return l.writeEntry(entry)
}

// writeEntry writes a single entry to the log file
func (l *Logger) writeEntry(entry Entry) error {
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	// Write as JSON with newline
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// GetUser returns the current user for audit logging
func GetUser() string {
	// Try environment variables first
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// BuildEntry is a helper to create audit entries
func BuildEntry(command string, scanDir string, envFiles []string,
	varsFound, filesScanned, filesSkipped int, duration time.Duration, status string) Entry {

	return Entry{
		Timestamp:    time.Now(),
		User:         GetUser(),
		Command:      command,
		ScanDir:      scanDir,
		EnvFiles:     envFiles,
		VarsFound:    varsFound,
		FilesScanned: filesScanned,
		FilesSkipped: filesSkipped,
		Duration:     duration.String(),
		Status:       status,
	}
}

// BuildErrorEntry creates an error audit entry
func BuildErrorEntry(command string, scanDir string, envFiles []string, errMsg string) Entry {
	return Entry{
		Timestamp: time.Now(),
		User:      GetUser(),
		Command:   command,
		ScanDir:   scanDir,
		EnvFiles:  envFiles,
		Status:    "error",
		ErrorMsg:  errMsg,
	}
}
