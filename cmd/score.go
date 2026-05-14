package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Hossiy21/taso/internal/audit"
	"github.com/Hossiy21/taso/internal/envreader"
	"github.com/Hossiy21/taso/internal/cache"
	"github.com/Hossiy21/taso/internal/scanner"
	"github.com/Hossiy21/taso/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scoreJSON bool
	scoreFail bool
	scoreDir  string
)

// scoreCmd represents the score command
var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Show your project's env health score (0–100)",
	Long: `Runs all taso checks and produces a single health score.
Factors in: ghost vars, stale/empty keys, missing .env.example,
no snapshot taken, and overall coverage.

Examples:
  taso score
  taso score --json`,
	RunE: runScore,
}

func init() {
	scoreCmd.Flags().BoolVar(&scoreJSON, "json", false, "Output as JSON")
	scoreCmd.Flags().BoolVar(&scoreFail, "fail", false, "Exit with error code 1 if score < 80 (for CI/CD)")
	scoreCmd.Flags().StringVar(&scoreDir, "dir", ".", "Directory to score")
}

type scoreReport struct {
	Score       int
	Grade       string
	Ghosts      int
	EmptyKeys   int
	HasExample  bool
	HasSnapshot bool
	TotalKeys   int
	Issues      []string
}

func runScore(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	report := &scoreReport{}
	score := 100
	issues := []string{}

	envFiles := autoDetectEnvFiles(scoreDir)

	// Factor 1: Ghost vars (-8 per ghost, max -40)
	if len(envFiles) > 0 {
		knownKeys := map[string]bool{}
		for _, ef := range envFiles {
			keys, _ := envreader.LoadKeys(ef)
			for _, k := range keys {
				knownKeys[k] = true
			}
		}
		report.TotalKeys = len(knownKeys)

		// CACHE: Load cache store
		cacheStore, _ := cache.NewStore(".taso")

		findings, err := scanner.ScanDir(".", viper.GetStringSlice("ignored_dirs"), cacheStore)
		
		// CACHE: Save cache
		if cacheStore != nil {
			_ = cacheStore.Save()
		}
		if err == nil {
			ghosts := 0
			for varName := range findings {
				if !knownKeys[varName] {
					ghosts++
				}
			}
			report.Ghosts = ghosts
			penalty := ghosts * 8
			if penalty > 40 {
				penalty = 40
			}
			score -= penalty
			if ghosts > 0 {
				issues = append(issues, fmt.Sprintf("%d ghost variable(s) found — run 'taso ghost' to see them", ghosts))
			}
		}
	}

	// Factor 2: Empty/placeholder keys (-3 each, max -20)
	if len(envFiles) > 0 {
		emptyCount := 0
		for _, ef := range envFiles {
			m, err := envreader.LoadMap(ef)
			if err != nil {
				continue
			}
			for _, v := range m {
				if isPlaceholder(v) {
					emptyCount++
				}
			}
		}
		report.EmptyKeys = emptyCount
		penalty := emptyCount * 3
		if penalty > 20 {
			penalty = 20
		}
		score -= penalty
		if emptyCount > 0 {
			issues = append(issues, fmt.Sprintf("%d empty or placeholder value(s) in your env files", emptyCount))
		}
	}

	// Factor 3: No .env.example (-10)
	_, err := envreader.LoadKeys(".env.example")
	report.HasExample = err == nil
	if !report.HasExample {
		score -= 10
		issues = append(issues, "no .env.example found — teammates can't onboard without it")
	}

	// Factor 4: No snapshot taken (-5)
	_, snapErr := loadSnapshot()
	report.HasSnapshot = snapErr == nil
	if !report.HasSnapshot {
		score -= 5
		issues = append(issues, "no snapshot taken — run 'taso snap' to track drift over time")
	}

	if score < 0 {
		score = 0
	}
	report.Score = score
	report.Grade = scoreGrade(score)
	report.Issues = issues

	// Log successful audit entry
	logger, _ := audit.NewLogger(".taso/audit")
	if logger != nil {
		logger.Log(audit.BuildEntry("score", ".", envFiles,
			report.Score, 0, 0, time.Since(startTime), "success"))
	}

	if scoreJSON {
		err := printScoreJSON(report)
		if err != nil {
			return err
		}
	} else {
		err := printScoreHuman(report)
		if err != nil {
			return err
		}
	}

	if scoreFail && report.Score < 80 {
		return fmt.Errorf("CI check failed: Environment health score is too low (%d/100)", report.Score)
	}

	return nil
}

func printScoreHuman(r *scoreReport) error {
	rr := ui.NewRenderer()
	rr.Println("")
	rr.Println(ui.Bold("  Env Health Score"))
	rr.Println("")

	bar := ui.ScoreBar(r.Score, 30)
	rr.Println(fmt.Sprintf("  %s  %s  %d/100", bar, gradeStyle(r.Grade), r.Score))
	rr.Println("")

	if len(r.Issues) == 0 {
		rr.Println(ui.Success("  ✔  All checks passed. Your env is clean."))
	} else {
		rr.Println(ui.Dim(fmt.Sprintf("  %d issue(s) found:", len(r.Issues))))
		rr.Println("")
		for _, issue := range r.Issues {
			rr.Println(ui.Warn2("  ⚠  " + issue))
		}
	}
	rr.Println("")
	fmt.Print(rr.String())
	return nil
}

func printScoreJSON(r *scoreReport) error {
	fmt.Printf("{\n")
	fmt.Printf("  \"score\": %d,\n", r.Score)
	fmt.Printf("  \"grade\": %q,\n", r.Grade)
	fmt.Printf("  \"ghosts\": %d,\n", r.Ghosts)
	fmt.Printf("  \"empty_keys\": %d,\n", r.EmptyKeys)
	fmt.Printf("  \"has_example\": %v,\n", r.HasExample)
	fmt.Printf("  \"has_snapshot\": %v,\n", r.HasSnapshot)
	fmt.Printf("  \"issues\": %s\n", jsonStringArray(r.Issues))
	fmt.Printf("}\n")
	return nil
}

func scoreGrade(s int) string {
	switch {
	case s >= 90:
		return "A"
	case s >= 75:
		return "B"
	case s >= 60:
		return "C"
	case s >= 40:
		return "D"
	default:
		return "F"
	}
}

func gradeStyle(g string) string {
	switch g {
	case "A":
		return ui.Success(g)
	case "B":
		return ui.Info(g)
	case "C", "D":
		return ui.Warn2(g)
	default:
		return ui.Danger(g)
	}
}

func isPlaceholder(v string) bool {
	if v == "" {
		return true
	}
	low := strings.ToLower(v)
	placeholders := []string{"todo", "changeme", "your_", "replace", "example", "xxxxx", "dummy", "test123", "password", "secret"}
	for _, p := range placeholders {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}
