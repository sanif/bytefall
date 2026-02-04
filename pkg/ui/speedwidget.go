package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SpeedStyle defines the visual style for the speed widget
type SpeedStyle int

const (
	SpeedStyleMinimal SpeedStyle = iota // Just numbers
	SpeedStyleBoxed                     // With title and structure
	SpeedStyleRetro                     // Terminal green RX/TX
	SpeedStyleNeon                      // Cyberpunk arrows
	SpeedStyleCompact                   // Horizontal layout
)

var speedStyle SpeedStyle = SpeedStyleMinimal

func SetSpeedStyle(style SpeedStyle) {
	speedStyle = style
}

func CycleSpeedStyle() {
	speedStyle = (speedStyle + 1) % 5
}

func GetSpeedStyleName() string {
	names := []string{"minimal", "boxed", "retro", "neon", "compact"}
	if int(speedStyle) < len(names) {
		return names[speedStyle]
	}
	return "minimal"
}

func RenderSpeedWidget(width, height int, downloadBytesPerSec, uploadBytesPerSec, totalPackets int64, theme Theme) string {
	switch speedStyle {
	case SpeedStyleMinimal:
		return renderSpeedMinimal(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	case SpeedStyleBoxed:
		return renderSpeedBoxed(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	case SpeedStyleRetro:
		return renderSpeedRetro(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	case SpeedStyleNeon:
		return renderSpeedNeon(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	case SpeedStyleCompact:
		return renderSpeedCompact(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	default:
		return renderSpeedMinimal(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	}
}

// ============================================================================
// MINIMAL - Just the numbers, nothing else
// ============================================================================
func renderSpeedMinimal(width, height int, download, upload int64, theme Theme) string {
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)

	var lines []string

	// Just the speeds - big and centered
	lines = append(lines, "")
	lines = append(lines, centerText(downloadStyle.Render("▼ "+formatSpeedLarge(download)), width))
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, centerText(uploadStyle.Render("▲ "+formatSpeedLarge(upload)), width))
	lines = append(lines, "")

	return verticalCenter(lines, width, height)
}

// ============================================================================
// BOXED - Title + structured layout with bars
// ============================================================================
func renderSpeedBoxed(width, height int, download, upload int64, theme Theme) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Text)

	var lines []string

	// Title - only this style has it
	lines = append(lines, "")
	lines = append(lines, centerText(titleStyle.Render("NETSPEED"), width))
	lines = append(lines, "")

	// Download
	lines = append(lines, centerText(labelStyle.Render("Download"), width))
	lines = append(lines, centerText(downloadStyle.Render(formatSpeedLarge(download)), width))

	barWidth := min(width-10, 40)
	downloadBar := renderBar(download, 100*1024*1024, barWidth, theme.Accent, theme.Dim)
	lines = append(lines, centerText(downloadBar, width))
	lines = append(lines, "")

	// Separator
	lines = append(lines, centerText(dimStyle.Render("· · ·"), width))
	lines = append(lines, "")

	// Upload
	lines = append(lines, centerText(labelStyle.Render("Upload"), width))
	lines = append(lines, centerText(uploadStyle.Render(formatSpeedLarge(upload)), width))

	uploadBar := renderBar(upload, 100*1024*1024, barWidth, theme.Secondary, theme.Dim)
	lines = append(lines, centerText(uploadBar, width))

	return verticalCenter(lines, width, height)
}

// ============================================================================
// RETRO - Green terminal with RX/TX
// ============================================================================
func renderSpeedRetro(width, height int, download, upload int64, theme Theme) string {
	green := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#33ff33"))
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("#117711"))

	var lines []string

	lines = append(lines, "")

	// RX with bracket bar
	rxSpeed := formatSpeedLarge(download)
	barWidth := min(width-20, 30)
	rxPct := minFloat(float64(download)/float64(100*1024*1024), 1.0)
	rxFilled := int(float64(barWidth) * rxPct)
	rxBar := "[" + strings.Repeat("#", rxFilled) + strings.Repeat("-", barWidth-rxFilled) + "]"

	lines = append(lines, centerText(green.Render("RX ")+green.Render(fmt.Sprintf("%12s", rxSpeed))+" "+dimGreen.Render(rxBar), width))
	lines = append(lines, "")

	// TX with bracket bar
	txSpeed := formatSpeedLarge(upload)
	txPct := minFloat(float64(upload)/float64(100*1024*1024), 1.0)
	txFilled := int(float64(barWidth) * txPct)
	txBar := "[" + strings.Repeat("#", txFilled) + strings.Repeat("-", barWidth-txFilled) + "]"

	lines = append(lines, centerText(green.Render("TX ")+green.Render(fmt.Sprintf("%12s", txSpeed))+" "+dimGreen.Render(txBar), width))
	lines = append(lines, "")

	return verticalCenter(lines, width, height)
}

// ============================================================================
// NEON - Cyberpunk with big arrows
// ============================================================================
func renderSpeedNeon(width, height int, download, upload int64, theme Theme) string {
	cyan := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ffff"))
	pink := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff00ff"))

	var lines []string

	lines = append(lines, "")

	// Big down arrow with speed
	lines = append(lines, centerText(cyan.Render("▼▼▼"), width))
	lines = append(lines, centerText(cyan.Render(formatSpeedLarge(download)), width))
	lines = append(lines, "")
	lines = append(lines, "")

	// Big up arrow with speed
	lines = append(lines, centerText(pink.Render("▲▲▲"), width))
	lines = append(lines, centerText(pink.Render(formatSpeedLarge(upload)), width))
	lines = append(lines, "")

	return verticalCenter(lines, width, height)
}

// ============================================================================
// COMPACT - Horizontal single-line style
// ============================================================================
func renderSpeedCompact(width, height int, download, upload int64, theme Theme) string {
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	var lines []string

	lines = append(lines, "")

	// Single line with both speeds
	line := downloadStyle.Render("▼ "+formatSpeedLarge(download)) +
		dimStyle.Render("  ·  ") +
		uploadStyle.Render("▲ "+formatSpeedLarge(upload))

	lines = append(lines, centerText(line, width))
	lines = append(lines, "")

	// Mini bars side by side
	barWidth := min((width-10)/2, 25)
	downBar := renderMiniBar(download, 100*1024*1024, barWidth, theme.Accent)
	upBar := renderMiniBar(upload, 100*1024*1024, barWidth, theme.Secondary)

	lines = append(lines, centerText(downBar+"  "+upBar, width))
	lines = append(lines, "")

	return verticalCenter(lines, width, height)
}

// ============================================================================
// HELPERS
// ============================================================================

func renderBar(value, maxValue int64, width int, fillColor, emptyColor lipgloss.Color) string {
	if maxValue == 0 {
		maxValue = 1
	}
	if width < 5 {
		width = 5
	}

	pct := float64(value) / float64(maxValue)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))

	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	emptyStyle := lipgloss.NewStyle().Foreground(emptyColor)

	return fillStyle.Render(strings.Repeat("━", filled)) + emptyStyle.Render(strings.Repeat("─", width-filled))
}

func renderMiniBar(value, maxValue int64, width int, color lipgloss.Color) string {
	if maxValue == 0 {
		maxValue = 1
	}
	pct := float64(value) / float64(maxValue)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))

	style := lipgloss.NewStyle().Foreground(color)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	return style.Render(strings.Repeat("━", filled)) + dimStyle.Render(strings.Repeat("─", width-filled))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func formatRateLarge(bytesPerSec int64) string {
	return formatSpeedLarge(bytesPerSec)
}

func formatSpeedLarge(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", float64(bytesPerSec)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB/s", float64(bytesPerSec)/(1024*1024*1024))
}

func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func verticalCenter(lines []string, width, height int) string {
	contentHeight := len(lines)
	topPadding := (height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var result []string
	for i := 0; i < topPadding; i++ {
		result = append(result, "")
	}
	result = append(result, lines...)

	for len(result) < height {
		result = append(result, "")
	}

	if len(result) > height {
		result = result[:height]
	}

	return strings.Join(result, "\n")
}
