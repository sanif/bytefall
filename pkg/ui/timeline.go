package ui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// TimelinePanel shows 60-second activity sparklines
type TimelinePanel struct {
	width   int
	height  int
	stats   []*data.DomainStats
	focused bool
}

// Sparkline characters from least to most filled
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// NewTimelinePanel creates a new timeline panel
func NewTimelinePanel(width, height int) *TimelinePanel {
	return &TimelinePanel{
		width:  width,
		height: height,
	}
}

// UpdateStats sets the domain statistics
func (t *TimelinePanel) UpdateStats(stats []*data.DomainStats) {
	t.stats = stats
}

// SetFocused sets the focus state
func (t *TimelinePanel) SetFocused(focused bool) {
	t.focused = focused
}

// Resize updates dimensions
func (t *TimelinePanel) Resize(width, height int) {
	t.width = width
	t.height = height
}

// View renders the timeline
func (t *TimelinePanel) View() string {
	theme := CurrentTheme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	// Sort by bytes descending
	sorted := make([]*data.DomainStats, len(t.stats))
	copy(sorted, t.stats)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Bytes > sorted[j].Bytes
	})

	// Build lines
	lines := make([]string, t.height)

	// Title with time indicator
	lines[0] = titleStyle.Render("  Activity Timeline") + dimStyle.Render("  (60s)")
	lines[1] = dimStyle.Render(strings.Repeat("─", t.width))

	// Calculate sparkline width
	labelWidth := 22
	sparkWidth := t.width - labelWidth - 6
	if sparkWidth < 20 {
		sparkWidth = 20
	}
	if sparkWidth > 60 {
		sparkWidth = 60
	}

	// Rows
	maxRows := t.height - 2
	for i := 0; i < maxRows; i++ {
		lineIdx := i + 2
		if lineIdx >= t.height {
			break
		}

		if i < len(sorted) {
			s := sorted[i]
			domain := s.Domain
			if len(domain) > labelWidth-2 {
				domain = domain[:labelWidth-4] + ".."
			}

			spark := t.renderColoredSparkline(s.History, sparkWidth, theme)
			lines[lineIdx] = "  " + labelStyle.Render(padRightSimple(domain, labelWidth-2)) + " " + spark
		} else {
			lines[lineIdx] = strings.Repeat(" ", t.width)
		}
	}

	// Fill remaining
	for i := range lines {
		if lines[i] == "" {
			lines[i] = strings.Repeat(" ", t.width)
		}
	}

	return strings.Join(lines, "\n")
}

func (t *TimelinePanel) renderColoredSparkline(history []int64, width int, theme Theme) string {
	if len(history) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
		return dimStyle.Render(strings.Repeat(string(sparkChars[0]), width))
	}

	// Find max value for scaling
	var maxVal int64
	for _, v := range history {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
		return dimStyle.Render(strings.Repeat(string(sparkChars[0]), width))
	}

	// Color gradient based on theme
	colors := []lipgloss.Color{
		theme.Rain1,
		theme.Rain2,
		theme.Rain3,
		theme.Rain4,
		theme.Rain5,
	}

	// Build sparkline with colors
	var spark strings.Builder
	for i := 0; i < width; i++ {
		// Map position to history index
		idx := i * len(history) / width
		if idx >= len(history) {
			idx = len(history) - 1
		}

		// Scale value to sparkline character
		val := history[idx]
		level := int(val * int64(len(sparkChars)-1) / maxVal)
		if level < 0 {
			level = 0
		}
		if level >= len(sparkChars) {
			level = len(sparkChars) - 1
		}

		// Color based on level
		colorIdx := level * len(colors) / len(sparkChars)
		if colorIdx >= len(colors) {
			colorIdx = len(colors) - 1
		}

		style := lipgloss.NewStyle().Foreground(colors[colorIdx])
		spark.WriteString(style.Render(string(sparkChars[level])))
	}

	return spark.String()
}

func padRightSimple(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}
