package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/Hossiy21/taso/internal/audit"
	"github.com/Hossiy21/taso/internal/envreader"
	"github.com/Hossiy21/taso/internal/ui"
	"github.com/spf13/cobra"
)

var driftJSON bool

var driftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Show what changed in your env since the last snapshot",
	Long: `Compares your current env files against the saved snapshot.
Shows added, removed, and changed variables.

Run 'taso snap' first to create a baseline.

Examples:
  taso drift
  taso drift --json`,
	RunE: runDrift,
}

func init() {
	driftCmd.Flags().BoolVar(&driftJSON, "json", false, "Output as JSON")
}

func runDrift(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	snap, err := loadSnapshot()
	if err != nil {
		return fmt.Errorf("no snapshot found — run 'taso snap' first")
	}

	current := map[string]string{}
	for _, ef := range snap.EnvFiles {
		m, err := envreader.LoadMap(ef)
		if err != nil {
			continue
		}
		for k, v := range m {
			current[k] = maskValue(v)
		}
	}

	added := []string{}
	removed := []string{}
	changed := []string{}

	for k := range current {
		if _, ok := snap.Keys[k]; !ok {
			added = append(added, k)
		} else if snap.Keys[k] != current[k] {
			changed = append(changed, k)
		}
	}
	for k := range snap.Keys {
		if _, ok := current[k]; !ok {
			removed = append(removed, k)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	// Log successful audit entry
	logger, _ := audit.NewLogger(".taso/audit")
	if logger != nil {
		logger.Log(audit.BuildEntry("drift", ".", snap.EnvFiles,
			len(added)+len(removed)+len(changed), 0, 0, time.Since(startTime), "success"))
	}

	if driftJSON {
		return printDriftJSON(snap, added, removed, changed)
	}
	return printDriftHuman(snap, added, removed, changed)
}

func printDriftHuman(snap *Snapshot, added, removed, changed []string) error {
	r := ui.NewRenderer()
	age := time.Since(snap.CreatedAt).Round(time.Minute)
	r.Println(ui.Dim(fmt.Sprintf("Snapshot taken: %s ago (%s)", age, snap.CreatedAt.Format("2006-01-02 15:04"))))
	r.Println("")

	if len(added)+len(removed)+len(changed) == 0 {
		r.Println(ui.Success("✔  No drift detected — env is unchanged since snapshot."))
		fmt.Print(r.String())
		return nil
	}

	total := len(added) + len(removed) + len(changed)
	r.Println(ui.Danger(fmt.Sprintf("⚡  %d change(s) detected", total)))
	r.Println("")

	for _, k := range added {
		r.Println(ui.Success(fmt.Sprintf("  + %-35s  added", k)))
	}
	for _, k := range removed {
		r.Println(ui.Danger(fmt.Sprintf("  - %-35s  removed", k)))
	}
	for _, k := range changed {
		r.Println(ui.Warn2(fmt.Sprintf("  ~ %-35s  value changed", k)))
	}

	r.Println("")
	r.Println(ui.Dim("Run 'taso snap' to update your baseline."))
	fmt.Print(r.String())
	return nil
}

func printDriftJSON(snap *Snapshot, added, removed, changed []string) error {
	fmt.Println("{")
	fmt.Printf("  \"snapshot_age_seconds\": %.0f,\n", time.Since(snap.CreatedAt).Seconds())
	fmt.Printf("  \"added\": %s,\n", jsonStringArray(added))
	fmt.Printf("  \"removed\": %s,\n", jsonStringArray(removed))
	fmt.Printf("  \"changed\": %s\n", jsonStringArray(changed))
	fmt.Println("}")
	return nil
}
