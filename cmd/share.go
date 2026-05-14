package cmd

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Hossiy21/taso/internal/audit"
	"github.com/Hossiy21/taso/internal/envreader"
	"github.com/Hossiy21/taso/internal/ui"
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Generate a shareable team fingerprint (keys only, no values)",
	Long: `Generates a safe digest of your env structure — key names and types,
no actual values. Share it with teammates so they can validate
their own setup matches yours.

Examples:
  taso share
  taso share --env .env.production`,
	RunE: runShare,
}

var (
	shareEnvFiles []string
	shareDir      string
)

func init() {
	shareCmd.Flags().StringArrayVar(&shareEnvFiles, "env", nil, "Env files to fingerprint")
	shareCmd.Flags().StringVar(&shareDir, "dir", ".", "Directory to fingerprint")
}

func runShare(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// SECURITY: Validate env file paths
	if len(shareEnvFiles) > 0 {
		for _, ef := range shareEnvFiles {
			if strings.Contains(ef, "..") {
				return fmt.Errorf("directory traversal not allowed in env file path: %q", ef)
			}
		}
	}

	if len(shareEnvFiles) == 0 {
		shareEnvFiles = autoDetectEnvFiles(shareDir)
		if len(shareEnvFiles) == 0 {
			return fmt.Errorf("no .env files found in %s", shareDir)
		}
	}

	allKeys := map[string]string{}
	for _, ef := range shareEnvFiles {
		m, err := envreader.LoadMap(ef)
		if err != nil {
			continue
		}
		for k, v := range m {
			allKeys[k] = classifyValue(v)
		}
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build a fingerprint hash from sorted keys
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k + ":" + allKeys[k] + "\n"))
	}
	fingerprint := fmt.Sprintf("%x", h.Sum(nil))[:12]

	r := ui.NewRenderer()
	r.Println(ui.Bold("  Team Fingerprint"))
	r.Println("")
	r.Println(ui.Dim(fmt.Sprintf("  %d variables across %d file(s)", len(keys), len(shareEnvFiles))))
	r.Println("")

	for _, k := range keys {
		vtype := allKeys[k]
		r.Println(fmt.Sprintf("  %-40s %s", k, ui.Dim(vtype)))
	}

	r.Println("")
	r.Println(ui.Bold("  Fingerprint: ") + ui.Info(fingerprint))
	r.Println(ui.Dim("  Share this fingerprint with teammates to verify env parity."))
	r.Println(ui.Dim("  No values are included — only key names and types."))
	fmt.Print(r.String())

	// Log successful audit entry
	logger, _ := audit.NewLogger(".taso/audit")
	if logger != nil {
		logger.Log(audit.BuildEntry("share", ".", shareEnvFiles,
			len(keys), 0, 0, time.Since(startTime), "success"))
	}

	return nil
}

func classifyValue(v string) string {
	if v == "" {
		return "empty"
	}
	low := fmt.Sprintf("%s", v)
	switch {
	case len(v) > 30 && isHex(v):
		return "secret"
	case len(v) > 20:
		return "long-string"
	case isURL(low):
		return "url"
	case isNumber(low):
		return "number"
	case low == "true" || low == "false":
		return "bool"
	default:
		return "string"
	}
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isURL(s string) bool {
	return len(s) > 4 && (s[:4] == "http" || s[:5] == "mysql" || s[:8] == "postgres")
}

func isNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
