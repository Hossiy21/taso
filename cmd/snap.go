package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hossiy21/taso/internal/audit"
	"github.com/Hossiy21/taso/internal/envreader"
	"github.com/Hossiy21/taso/internal/ui"
	"github.com/spf13/cobra"
)

const snapshotDir = ".taso"
const snapshotFile = ".taso/snapshot.json"

type Snapshot struct {
	CreatedAt time.Time         `json:"created_at"`
	EnvFiles  []string          `json:"env_files"`
	Keys      map[string]string `json:"keys"` // key -> masked value hash
}

var snapEnvFiles []string

var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Save a snapshot of your current env state",
	Long: `Saves the current state of your env files to disk.
Use 'taso drift' later to see what changed.

Examples:
  taso snap
  taso snap --env .env --env .env.local`,
	RunE: runSnap,
}

func init() {
	snapCmd.Flags().StringArrayVar(&snapEnvFiles, "env", nil, "Env files to snapshot (auto-detected if not set)")
}

func runSnap(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// SECURITY: Validate env file paths
	if len(snapEnvFiles) > 0 {
		for _, ef := range snapEnvFiles {
			if strings.Contains(ef, "..") {
				return fmt.Errorf("directory traversal not allowed in env file path: %q", ef)
			}
		}
	}

	if len(snapEnvFiles) == 0 {
		snapEnvFiles = autoDetectEnvFiles(".")
		if len(snapEnvFiles) == 0 {
			return fmt.Errorf("no .env files found. Use --env to specify one")
		}
	}

	combined := map[string]string{}
	for _, ef := range snapEnvFiles {
		m, err := envreader.LoadMap(ef)
		if err != nil {
			ui.Warn(fmt.Sprintf("Could not read %s: %v", ef, err))
			continue
		}
		for k, v := range m {
			combined[k] = maskValue(v)
		}
	}

	snap := Snapshot{
		CreatedAt: time.Now(),
		EnvFiles:  snapEnvFiles,
		Keys:      combined,
	}

	if err := os.MkdirAll(snapshotDir, 0700); err != nil {
		return fmt.Errorf("could not create .taso dir: %w", err)
	}

	f, err := os.Create(snapshotFile)
	if err != nil {
		return fmt.Errorf("could not write snapshot: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return err
	}

	absPath, _ := filepath.Abs(snapshotFile)
	r := ui.NewRenderer()
	r.Println(ui.Success(fmt.Sprintf("✔  Snapshot saved (%d keys)", len(combined))))
	r.Println(ui.Dim("   " + absPath))
	r.Println(ui.Dim("   Run 'taso drift' anytime to see what changed."))
	fmt.Print(r.String())

	// Log successful audit entry
	logger, _ := audit.NewLogger(".taso/audit")
	if logger != nil {
		logger.Log(audit.BuildEntry("snap", ".", snapEnvFiles,
			len(combined), 0, 0, time.Since(startTime), "success"))
	}

	return nil
}

// maskValue stores a short hash of the value so drift can detect changes
// without storing actual secrets on disk
func maskValue(v string) string {
	if v == "" {
		return ""
	}
	h := 0
	for _, c := range v {
		h = h*31 + int(c)
	}
	return fmt.Sprintf("hash:%x", h&0xFFFFFF)
}

func loadSnapshot() (*Snapshot, error) {
	f, err := os.Open(snapshotFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
