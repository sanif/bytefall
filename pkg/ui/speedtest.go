package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// SpeedTestPopup displays a speed test overlay
type SpeedTestPopup struct {
	visible bool
	result  data.SpeedTestResult
	test    *data.SpeedTest
}

// NewSpeedTestPopup creates a new speed test popup
func NewSpeedTestPopup() *SpeedTestPopup {
	return &SpeedTestPopup{}
}

// Show displays the popup and starts the test
func (s *SpeedTestPopup) Show() {
	s.visible = true
	s.result = data.SpeedTestResult{Status: "Starting...", Phase: "init"}
	if s.test == nil {
		s.test = data.NewSpeedTest(func(r data.SpeedTestResult) {
			s.result = r
		})
	}
	s.test.Start()
}

// Hide closes the popup
func (s *SpeedTestPopup) Hide() {
	s.visible = false
	if s.test != nil {
		s.test.Stop()
	}
}

// Restart restarts the speed test
func (s *SpeedTestPopup) Restart() {
	if s.test != nil {
		s.test.Stop()
	}
	s.result = data.SpeedTestResult{Status: "Starting...", Phase: "init"}
	s.test = data.NewSpeedTest(func(r data.SpeedTestResult) {
		s.result = r
	})
	s.test.Start()
}

// Toggle toggles visibility
func (s *SpeedTestPopup) Toggle() {
	if s.visible {
		s.Hide()
	} else {
		s.Show()
	}
}

// IsVisible returns visibility state
func (s *SpeedTestPopup) IsVisible() bool {
	return s.visible
}

// UpdateResult sets the current result
func (s *SpeedTestPopup) UpdateResult(r data.SpeedTestResult) {
	s.result = r
}

// Resize updates dimensions (no-op, we calculate dynamically)
func (s *SpeedTestPopup) Resize(termWidth, termHeight int) {}

// View renders the popup as a full-screen modal
func (s *SpeedTestPopup) View(termWidth, termHeight int) string {
	if !s.visible {
		return ""
	}

	theme := CurrentTheme

	// Colors from theme
	green := theme.Primary
	dimGreen := theme.Dim
	darkGreen := theme.Dim
	gray := theme.TextDim
	bg := theme.Background

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(green).
		Background(darkGreen).
		Padding(0, 2)

	labelStyle := lipgloss.NewStyle().Foreground(gray)
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(green)
	statusStyle := lipgloss.NewStyle().Foreground(dimGreen)

	// Build content
	var lines []string

	// Title
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("⚡ SPEED TEST"))
	lines = append(lines, "")

	// Server info
	server := s.result.Server
	if server == "" {
		server = "Cloudflare"
	}
	lines = append(lines, labelStyle.Render("  Server: ")+valueStyle.Render(server))
	lines = append(lines, "")

	// Gauges
	lines = append(lines, s.gauge("  PING    ", s.result.Latency.Milliseconds(), 200, "ms", true))
	lines = append(lines, s.gauge("  DOWN    ", int64(s.result.DownloadSpeed), 500, "Mbps", false))
	lines = append(lines, s.gauge("  UP      ", int64(s.result.UploadSpeed), 500, "Mbps", false))
	lines = append(lines, "")

	// Big stats
	downStr := fmt.Sprintf("%.1f", s.result.DownloadSpeed)
	upStr := fmt.Sprintf("%.1f", s.result.UploadSpeed)
	pingStr := fmt.Sprintf("%d", s.result.Latency.Milliseconds())

	statsLine := fmt.Sprintf("  ↓ %s Mbps    ↑ %s Mbps    ◉ %s ms",
		valueStyle.Render(downStr),
		valueStyle.Render(upStr),
		valueStyle.Render(pingStr))
	lines = append(lines, statsLine)
	lines = append(lines, "")

	// Progress
	lines = append(lines, s.progress(s.result.Progress))
	lines = append(lines, "")

	// Status
	status := s.result.Status
	if status == "" {
		status = "Ready"
	}
	lines = append(lines, "  "+statusStyle.Render(status))
	lines = append(lines, "")

	// Help
	lines = append(lines, labelStyle.Render("  [s] restart  [ESC] close"))
	lines = append(lines, "")

	// Calculate box dimensions
	contentWidth := 50
	_ = len(lines) // contentHeight for reference

	// Create bordered box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(green).
		Background(bg).
		Width(contentWidth).
		Padding(0, 1)

	content := strings.Join(lines, "\n")
	box := boxStyle.Render(content)

	// Center in terminal
	boxLines := strings.Split(box, "\n")
	boxHeight := len(boxLines)
	boxWidth := lipgloss.Width(box)

	padTop := (termHeight - boxHeight) / 2
	padLeft := (termWidth - boxWidth) / 2

	if padTop < 0 {
		padTop = 0
	}
	if padLeft < 0 {
		padLeft = 0
	}

	// Build final output with dark background
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color("#050505"))

	var output strings.Builder

	// Top padding (dark)
	for i := 0; i < padTop; i++ {
		output.WriteString(bgStyle.Render(strings.Repeat(" ", termWidth)))
		output.WriteString("\n")
	}

	// Box lines centered
	for i, line := range boxLines {
		leftPad := bgStyle.Render(strings.Repeat(" ", padLeft))
		rightPadWidth := termWidth - padLeft - lipgloss.Width(line)
		if rightPadWidth < 0 {
			rightPadWidth = 0
		}
		rightPad := bgStyle.Render(strings.Repeat(" ", rightPadWidth))
		output.WriteString(leftPad + line + rightPad)
		if i < len(boxLines)-1 {
			output.WriteString("\n")
		}
	}

	// Bottom padding
	remainingHeight := termHeight - padTop - boxHeight
	for i := 0; i < remainingHeight; i++ {
		output.WriteString("\n")
		output.WriteString(bgStyle.Render(strings.Repeat(" ", termWidth)))
	}

	return output.String()
}

func (s *SpeedTestPopup) gauge(label string, value int64, maxVal int64, unit string, isMs bool) string {
	theme := CurrentTheme
	green := theme.Primary
	dimGreen := theme.Dim
	gray := theme.TextDim

	labelStyle := lipgloss.NewStyle().Foreground(gray)
	filledStyle := lipgloss.NewStyle().Foreground(green)
	emptyStyle := lipgloss.NewStyle().Foreground(dimGreen)
	valStyle := lipgloss.NewStyle().Foreground(green)

	barWidth := 25
	filled := int(float64(value) / float64(maxVal) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", barWidth-filled))

	valStr := fmt.Sprintf("%d %s", value, unit)

	return labelStyle.Render(label) + bar + " " + valStyle.Render(valStr)
}

func (s *SpeedTestPopup) progress(pct float64) string {
	theme := CurrentTheme
	green := theme.Primary
	dimGreen := theme.Dim

	width := 40
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	filledStyle := lipgloss.NewStyle().Foreground(green)
	emptyStyle := lipgloss.NewStyle().Foreground(dimGreen)

	bar := filledStyle.Render(strings.Repeat("━", filled)) +
		emptyStyle.Render(strings.Repeat("─", width-filled))

	return fmt.Sprintf("  [%s] %.0f%%", bar, pct)
}
