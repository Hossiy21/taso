package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

func Success(s string) string { return successStyle.Render(s) }
func Danger(s string) string  { return dangerStyle.Render(s) }
func Warn2(s string) string   { return warnStyle.Render(s) }
func Warn(s string) string    { return warnStyle.Render("⚠  " + s) }
func Dim(s string) string     { return dimStyle.Render(s) }
func Bold(s string) string    { return boldStyle.Render(s) }
func Info(s string) string    { return infoStyle.Render(s) }

// Renderer collects lines before printing (so we can test output easily)
type Renderer struct {
	buf strings.Builder
}

func NewRenderer() *Renderer { return &Renderer{} }

func (r *Renderer) Println(s string) {
	r.buf.WriteString(s + "\n")
}

func (r *Renderer) String() string {
	return r.buf.String()
}

// ScoreBar renders a 0-100 bar in the terminal
func ScoreBar(score int, width int) string {
	filled := (score * width) / 100
	if filled > width {
		filled = width
	}

	// Use safer characters for better Windows compatibility
	// Some older CMD terminals don't handle '█' and '░' well
	charFilled := "="
	charEmpty := "-"
	
	bar := strings.Repeat(charFilled, filled) + strings.Repeat(charEmpty, width-filled)
	bar = "[" + bar + "]"

	var color lipgloss.Color
	switch {
	case score >= 80:
		color = lipgloss.Color("10") // green
	case score >= 50:
		color = lipgloss.Color("11") // yellow
	default:
		color = lipgloss.Color("9") // red
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar)
}
