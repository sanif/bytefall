package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// LeaderboardPanel displays top domains by traffic
type LeaderboardPanel struct {
	width       int
	height      int
	stats       []*data.DomainStats
	focused     bool
	selectedIdx int
}

// NewLeaderboardPanel creates a new leaderboard panel
func NewLeaderboardPanel(width, height int) *LeaderboardPanel {
	return &LeaderboardPanel{
		width:  width,
		height: height,
	}
}

// UpdateStats sets the domain statistics
func (l *LeaderboardPanel) UpdateStats(stats []*data.DomainStats) {
	l.stats = stats
}

// SetFocused sets the focus state
func (l *LeaderboardPanel) SetFocused(focused bool) {
	l.focused = focused
}

// Resize updates dimensions
func (l *LeaderboardPanel) Resize(width, height int) {
	l.width = width
	l.height = height
}

// SetSelected sets the selected index
func (l *LeaderboardPanel) SetSelected(idx int) {
	l.selectedIdx = idx
}

// View renders the leaderboard
func (l *LeaderboardPanel) View() string {
	theme := CurrentTheme

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	headerStyle := lipgloss.NewStyle().
		Foreground(theme.TextDim)

	rowStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary)

	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	accentStyle := lipgloss.NewStyle().
		Foreground(theme.Accent)

	// Sort by bytes descending
	sorted := make([]*data.DomainStats, len(l.stats))
	copy(sorted, l.stats)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Bytes > sorted[j].Bytes
	})

	// Find max bytes for bar scaling
	var maxBytes int64
	for _, s := range sorted {
		if s.Bytes > maxBytes {
			maxBytes = s.Bytes
		}
	}

	// Build lines
	lines := make([]string, l.height)

	// Title
	lines[0] = titleStyle.Render("  Top Domains")
	lines[1] = dimStyle.Render(strings.Repeat("─", l.width))

	// Header
	lines[2] = headerStyle.Render(fmt.Sprintf("  %-18s %10s  %s", "DOMAIN", "TRAFFIC", ""))

	// Rows
	maxRows := l.height - 3
	barWidth := l.width - 35
	if barWidth < 5 {
		barWidth = 5
	}
	if barWidth > 20 {
		barWidth = 20
	}

	for i := 0; i < maxRows; i++ {
		lineIdx := i + 3
		if lineIdx >= l.height {
			break
		}

		if i < len(sorted) {
			s := sorted[i]
			domain := s.Domain
			maxDomainLen := 16
			if len(domain) > maxDomainLen {
				domain = domain[:maxDomainLen-2] + ".."
			}

			// Traffic bar
			bar := l.renderBar(s.Bytes, maxBytes, barWidth, theme)

			// Rank indicator
			rank := fmt.Sprintf("%d", i+1)
			if i == 0 {
				rank = accentStyle.Render("●")
			} else if i < 3 {
				rank = dimStyle.Render("○")
			} else {
				rank = dimStyle.Render(fmt.Sprintf("%d", i+1))
			}

			row := fmt.Sprintf(" %s %-18s %8s  %s",
				rank,
				domain,
				formatBytes(s.Bytes),
				bar,
			)

			style := rowStyle
			if i == l.selectedIdx && l.focused {
				style = selectedStyle
				row = fmt.Sprintf(" %s %-18s %8s  %s",
					fmt.Sprintf("%d", i+1),
					domain,
					formatBytes(s.Bytes),
					strings.Repeat("█", barWidth),
				)
			}
			lines[lineIdx] = style.Render(padRight(row, l.width))
		} else {
			lines[lineIdx] = strings.Repeat(" ", l.width)
		}
	}

	// Fill remaining
	for i := range lines {
		if lines[i] == "" {
			lines[i] = strings.Repeat(" ", l.width)
		}
	}

	return strings.Join(lines, "\n")
}

func (l *LeaderboardPanel) renderBar(value, maxVal int64, width int, theme Theme) string {
	if maxVal == 0 {
		return strings.Repeat("░", width)
	}

	filled := int(float64(value) / float64(maxVal) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	// Color based on fill level
	var barColor lipgloss.Color
	pct := float64(filled) / float64(width)
	if pct > 0.8 {
		barColor = theme.Rain5
	} else if pct > 0.6 {
		barColor = theme.Rain4
	} else if pct > 0.4 {
		barColor = theme.Rain3
	} else if pct > 0.2 {
		barColor = theme.Rain2
	} else {
		barColor = theme.Rain1
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	return barStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func padRight(s string, width int) string {
	// Account for ANSI escape codes
	visLen := lipgloss.Width(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}
