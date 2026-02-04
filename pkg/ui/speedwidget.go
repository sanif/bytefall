package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderSpeedWidget creates a large centered speed display with gauges
func RenderSpeedWidget(width, height int, downloadBytesPerSec, uploadBytesPerSec, totalPackets int64, theme Theme) string {
	// Styles
	downloadStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent)

	uploadStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Secondary)

	packetStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	// Calculate gauge parameters
	gaugeWidth := width * 60 / 100
	if gaugeWidth < 30 {
		gaugeWidth = 30
	}
	if gaugeWidth > 80 {
		gaugeWidth = 80
	}

	// Calculate peak reference for gauge scaling (10 MB/s as a reasonable max)
	maxSpeed := int64(10 * 1024 * 1024) // 10 MB/s

	// Build content lines
	var lines []string
	lines = append(lines, "")

	// Download speed with arrow: ↓ 12.5 MB/s
	downloadRate := formatRateLarge(downloadBytesPerSec)
	downloadLabel := downloadStyle.Render("↓ " + downloadRate)
	lines = append(lines, centerText(downloadLabel, width))

	// Download gauge
	downloadGauge := renderSpeedGauge(downloadBytesPerSec, maxSpeed, gaugeWidth, theme, true)
	lines = append(lines, centerText(downloadGauge, width))
	lines = append(lines, "")

	// Upload speed with arrow: ↑ 3.2 MB/s
	uploadRate := formatRateLarge(uploadBytesPerSec)
	uploadLabel := uploadStyle.Render("↑ " + uploadRate)
	lines = append(lines, centerText(uploadLabel, width))

	// Upload gauge
	uploadGauge := renderSpeedGauge(uploadBytesPerSec, maxSpeed, gaugeWidth, theme, false)
	lines = append(lines, centerText(uploadGauge, width))
	lines = append(lines, "")

	// Packet counter
	packetText := packetStyle.Render(fmt.Sprintf("◈ %s packets/s", formatNumber(totalPackets)))
	lines = append(lines, centerText(packetText, width))

	// Calculate vertical centering
	contentHeight := len(lines)
	topPadding := (height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Build final output with vertical centering
	var result []string
	for i := 0; i < topPadding; i++ {
		result = append(result, strings.Repeat(" ", width))
	}
	result = append(result, lines...)

	// Fill remaining height
	for len(result) < height {
		result = append(result, strings.Repeat(" ", width))
	}

	// Truncate if too tall
	if len(result) > height {
		result = result[:height]
	}

	return strings.Join(result, "\n")
}

// renderSpeedGauge creates an animated gauge bar
func renderSpeedGauge(value, maxValue int64, width int, theme Theme, isDownload bool) string {
	if maxValue == 0 {
		maxValue = 1
	}

	// Calculate fill percentage
	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}

	filled := int(fillPct * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	// Color based on fill level and type
	var barColor lipgloss.Color
	if isDownload {
		if fillPct > 0.8 {
			barColor = theme.Rain5
		} else if fillPct > 0.6 {
			barColor = theme.Rain4
		} else if fillPct > 0.4 {
			barColor = theme.Rain3
		} else if fillPct > 0.2 {
			barColor = theme.Rain2
		} else {
			barColor = theme.Rain1
		}
	} else {
		// Upload uses secondary color tones
		if fillPct > 0.8 {
			barColor = theme.Secondary
		} else if fillPct > 0.6 {
			barColor = theme.Secondary
		} else if fillPct > 0.4 {
			barColor = theme.Secondary
		} else {
			barColor = theme.Secondary
		}
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	filledBar := strings.Repeat("", filled)
	emptyBar := strings.Repeat("", width-filled)

	return barStyle.Render(filledBar) + emptyStyle.Render(emptyBar)
}

// formatRateLarge formats bytes per second as a large display string
func formatRateLarge(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", float64(bytesPerSec)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB/s", float64(bytesPerSec)/(1024*1024*1024))
	}
}

// formatNumber formats a large number with commas
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	var result []string
	for i := len(str); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		result = append([]string{str[start:i]}, result...)
	}
	return strings.Join(result, ",")
}

// centerText centers a string within the given width
func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-padding-textWidth)
}
