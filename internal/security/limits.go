package security

import (
	"fmt"
	"os"
	"time"
)

const (
	// MaxFileSizeBytes is the maximum size of a single file to scan (50MB)
	MaxFileSizeBytes = 50 * 1024 * 1024

	// MaxScanTimeSeconds is the maximum time allowed for a single scan (5 minutes)
	MaxScanTimeSeconds = 5 * 60

	// MaxTotalMemoryMB is the approximate maximum memory to use (500MB)
	MaxTotalMemoryMB = 500
)

// ResourceMonitor tracks resource usage during scans to prevent DoS
type ResourceMonitor struct {
	startTime     time.Time
	maxDuration   time.Duration
	maxFileSize   int64
	maxTotalBytes int64
	filesScanned  int
	bytesScanned  int64
	skippedFiles  int
	lastCheckTime time.Time
	checkInterval time.Duration
}

// NewResourceMonitor creates a new resource monitor with default limits
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		startTime:     time.Now(),
		maxDuration:   time.Duration(MaxScanTimeSeconds) * time.Second,
		maxFileSize:   int64(MaxFileSizeBytes),
		maxTotalBytes: 1024 * 1024 * 1024, // 1GB total
		checkInterval: 100 * time.Millisecond,
		lastCheckTime: time.Now(),
	}
}

// CheckFileSize checks if a file is within size limits
func (r *ResourceMonitor) CheckFileSize(size int64) error {
	if size > r.maxFileSize {
		r.skippedFiles++
		return fmt.Errorf("file size %d bytes exceeds limit of %d bytes", size, r.maxFileSize)
	}
	return nil
}

// CheckLimits performs periodic checks on resource usage
func (r *ResourceMonitor) CheckLimits() error {
	now := time.Now()

	// Only check periodically to avoid overhead
	if now.Sub(r.lastCheckTime) < r.checkInterval {
		return nil
	}
	r.lastCheckTime = now

	// Check timeout
	elapsed := now.Sub(r.startTime)
	if elapsed > r.maxDuration {
		return fmt.Errorf("scan timeout: exceeded %v limit", r.maxDuration)
	}

	// Check total bytes read
	if r.bytesScanned > r.maxTotalBytes {
		return fmt.Errorf("total data scanned (%d bytes) exceeds limit (%d bytes)",
			r.bytesScanned, r.maxTotalBytes)
	}

	return nil
}

// RecordFile records that a file was scanned
func (r *ResourceMonitor) RecordFile(size int64) error {
	if err := r.CheckFileSize(size); err != nil {
		return err
	}

	r.filesScanned++
	r.bytesScanned += size

	// Check limits every 10 files or every 10MB to be more responsive
	if r.filesScanned%10 == 0 || r.bytesScanned% (10*1024*1024) < size {
		return r.CheckLimits()
	}

	return nil
}

// RecordSkippedFile records a skipped file
func (r *ResourceMonitor) RecordSkippedFile() {
	r.skippedFiles++
}

// Summary returns a summary of resource usage
func (r *ResourceMonitor) Summary() string {
	elapsed := time.Since(r.startTime)
	return fmt.Sprintf("scanned %d files (%d skipped) in %v, %d MB total",
		r.filesScanned, r.skippedFiles, elapsed, r.bytesScanned/1024/1024)
}

// GetRemainingTime returns how much time is left before timeout
func (r *ResourceMonitor) GetRemainingTime() time.Duration {
	elapsed := time.Since(r.startTime)
	remaining := r.maxDuration - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// ShouldSkipFile determines if a file should be skipped based on size
func ShouldSkipFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	// Skip very large files (likely not source code)
	if info.Size() > MaxFileSizeBytes {
		return true
	}

	// Skip if it looks like a binary file
	if isBinaryFile(filePath) {
		return true
	}

	return false
}

// isBinaryFile does a simple heuristic check for binary files
func isBinaryFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return true
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return true
	}

	// Look for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}
