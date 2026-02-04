package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SpeedStyle defines the visual style for the speed widget
type SpeedStyle int

const (
	SpeedStyleMinimal  SpeedStyle = iota // Clean, simple arrows and numbers
	SpeedStyleBoxed                      // Numbers in decorative boxes
	SpeedStyleRetro                      // Retro terminal aesthetic
	SpeedStyleNeon                       // Cyberpunk neon style
	SpeedStyleCompact                    // Compact single-line per metric
)

var speedStyle SpeedStyle = SpeedStyleBoxed

// SetSpeedStyle sets the visual style for the speed widget
func SetSpeedStyle(style SpeedStyle) {
	speedStyle = style
}

// CycleSpeedStyle cycles through available speed styles
func CycleSpeedStyle() {
	speedStyle = (speedStyle + 1) % 5 // 5 styles total
}

// GetSpeedStyleName returns the name of the current speed style
func GetSpeedStyleName() string {
	switch speedStyle {
	case SpeedStyleMinimal:
		return "minimal"
	case SpeedStyleBoxed:
		return "boxed"
	case SpeedStyleRetro:
		return "retro"
	case SpeedStyleNeon:
		return "neon"
	case SpeedStyleCompact:
		return "compact"
	default:
		return "boxed"
	}
}

// Refined ASCII art digits - LARGE (5 lines tall) with elegant hollow design
var bigDigitsLarge = map[rune][]string{
	'0': {"┌───┐", "│   │", "│   │", "│   │", "└───┘"},
	'1': {"    ╷", "    │", "    │", "    │", "    ╵"},
	'2': {"╶───┐", "    │", "┌───┘", "│    ", "└───╴"},
	'3': {"╶───┐", "    │", " ───┤", "    │", "╶───┘"},
	'4': {"╷   ╷", "│   │", "└───┤", "    │", "    ╵"},
	'5': {"┌───╴", "│    ", "└───┐", "    │", "╶───┘"},
	'6': {"┌───╴", "│    ", "├───┐", "│   │", "└───┘"},
	'7': {"╶───┐", "    │", "    │", "    │", "    ╵"},
	'8': {"┌───┐", "│   │", "├───┤", "│   │", "└───┘"},
	'9': {"┌───┐", "│   │", "└───┤", "    │", "╶───┘"},
	'.': {"     ", "     ", "     ", "  ●  ", "     "},
	' ': {"   ", "   ", "   ", "   ", "   "},
}

// Medium ASCII art digits (3 lines tall) with refined look
var bigDigitsMedium = map[rune][]string{
	'0': {"┌─┐", "│ │", "└─┘"},
	'1': {" ╷ ", " │ ", " ╵ "},
	'2': {"╶─┐", "┌─┘", "└─╴"},
	'3': {"╶─┐", " ─┤", "╶─┘"},
	'4': {"╷ ╷", "└─┤", "  ╵"},
	'5': {"┌─╴", "└─┐", "╶─┘"},
	'6': {"┌─╴", "├─┐", "└─┘"},
	'7': {"╶─┐", "  │", "  ╵"},
	'8': {"┌─┐", "├─┤", "└─┘"},
	'9': {"┌─┐", "└─┤", "╶─┘"},
	'.': {"   ", " ● ", "   "},
	' ': {"  ", "  ", "  "},
}

// Small digits (1 line - styled text)
var bigDigitsSmall = map[rune][]string{
	'0': {"0"}, '1': {"1"}, '2': {"2"}, '3': {"3"}, '4': {"4"},
	'5': {"5"}, '6': {"6"}, '7': {"7"}, '8': {"8"}, '9': {"9"},
	'.': {"."}, ' ': {" "},
}

// RenderSpeedWidget creates a large centered speed display with gauges
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
		return renderSpeedBoxed(width, height, downloadBytesPerSec, uploadBytesPerSec, theme)
	}
}

// getDigitSize returns the appropriate digit map based on available space
func getDigitSize(width, height int) (map[rune][]string, int) {
	if height >= 35 && width >= 60 {
		return bigDigitsLarge, 5
	} else if height >= 20 && width >= 40 {
		return bigDigitsMedium, 3
	}
	return bigDigitsSmall, 1
}

// ============================================================================
// STYLE: MINIMAL - Clean and elegant
// ============================================================================
func renderSpeedMinimal(width, height int, download, upload int64, theme Theme) string {
	digits, digitHeight := getDigitSize(width, height)

	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Text)

	var lines []string

	// Elegant header
	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("─────────────────────────────"), width))
	lines = append(lines, centerText(dimStyle.Render("System-wide"), width))
	lines = append(lines, "")

	// Download section
	lines = append(lines, centerText(labelStyle.Render("DOWNLOAD"), width))
	lines = append(lines, "")

	downloadRate, downloadUnit := formatRateComponents(download)
	bigDownload := renderBigNumber(downloadRate, digits, digitHeight)
	for _, line := range bigDownload {
		lines = append(lines, centerText(downloadStyle.Render(line)+"   "+dimStyle.Render(downloadUnit), width))
	}

	// Subtle progress indicator
	lines = append(lines, "")
	downloadProgress := renderMinimalProgress(download, int64(100*1024*1024), 30, theme.Accent)
	lines = append(lines, centerText(downloadProgress, width))

	// Elegant separator
	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("·  ·  ·"), width))
	lines = append(lines, "")

	// Upload section
	lines = append(lines, centerText(labelStyle.Render("UPLOAD"), width))
	lines = append(lines, "")

	uploadRate, uploadUnit := formatRateComponents(upload)
	bigUpload := renderBigNumber(uploadRate, digits, digitHeight)
	for _, line := range bigUpload {
		lines = append(lines, centerText(uploadStyle.Render(line)+"   "+dimStyle.Render(uploadUnit), width))
	}

	// Subtle progress indicator
	lines = append(lines, "")
	uploadProgress := renderMinimalProgress(upload, int64(100*1024*1024), 30, theme.Secondary)
	lines = append(lines, centerText(uploadProgress, width))

	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("─────────────────────────────"), width))

	return verticalCenter(lines, width, height)
}

func renderMinimalProgress(value, maxValue int64, width int, color lipgloss.Color) string {
	if maxValue == 0 {
		maxValue = 1
	}
	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}
	filled := int(fillPct * float64(width))

	barStyle := lipgloss.NewStyle().Foreground(color)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	return barStyle.Render(strings.Repeat("━", filled)) + dimStyle.Render(strings.Repeat("╌", width-filled))
}

// ============================================================================
// STYLE: BOXED - Professional with elegant frames
// ============================================================================
func renderSpeedBoxed(width, height int, download, upload int64, theme Theme) string {
	digits, digitHeight := getDigitSize(width, height)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Text)

	gaugeWidth := min(width*45/100, 45)
	if gaugeWidth < 20 {
		gaugeWidth = 20
	}
	maxSpeed := int64(100 * 1024 * 1024) // 100 MB/s scale

	var lines []string

	// Elegant title with decorative corners
	lines = append(lines, "")
	titleFrame := dimStyle.Render("╭─────") + "  " + titleStyle.Render("NETWORK SPEED") + "  " + dimStyle.Render("─────╮")
	lines = append(lines, centerText(titleFrame, width))
	lines = append(lines, centerText(dimStyle.Render("System-wide"), width))
	lines = append(lines, "")

	// Download section
	lines = append(lines, centerText(downloadStyle.Render("▾")+"  "+labelStyle.Render("DOWNLOAD"), width))
	lines = append(lines, "")

	downloadRate, downloadUnit := formatRateComponents(download)
	bigDownload := renderBigNumber(downloadRate, digits, digitHeight)
	for _, line := range bigDownload {
		lines = append(lines, centerText(downloadStyle.Render(line)+"   "+dimStyle.Render(downloadUnit), width))
	}
	lines = append(lines, "")

	// Elegant gauge
	downloadGauge := renderElegantGauge(download, maxSpeed, gaugeWidth, theme, true)
	lines = append(lines, centerText(downloadGauge, width))

	// Divider
	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("╶─────────────────────────────╴"), width))
	lines = append(lines, "")

	// Upload section
	lines = append(lines, centerText(uploadStyle.Render("▴")+"  "+labelStyle.Render("UPLOAD"), width))
	lines = append(lines, "")

	uploadRate, uploadUnit := formatRateComponents(upload)
	bigUpload := renderBigNumber(uploadRate, digits, digitHeight)
	for _, line := range bigUpload {
		lines = append(lines, centerText(uploadStyle.Render(line)+"   "+dimStyle.Render(uploadUnit), width))
	}
	lines = append(lines, "")

	// Elegant gauge
	uploadGauge := renderElegantGauge(upload, maxSpeed, gaugeWidth, theme, false)
	lines = append(lines, centerText(uploadGauge, width))

	// Bottom frame
	lines = append(lines, "")
	bottomFrame := dimStyle.Render("╰─────────────────────────────────────╯")
	lines = append(lines, centerText(bottomFrame, width))

	return verticalCenter(lines, width, height)
}

func renderElegantGauge(value, maxValue int64, width int, theme Theme, isDownload bool) string {
	if maxValue == 0 {
		maxValue = 1
	}

	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}

	innerWidth := width - 4
	filled := int(fillPct * float64(innerWidth))
	if filled < 0 {
		filled = 0
	}
	if filled > innerWidth {
		filled = innerWidth
	}

	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	var barContent strings.Builder
	for i := 0; i < innerWidth; i++ {
		if i < filled {
			var barColor lipgloss.Color
			pct := float64(i) / float64(innerWidth)
			if isDownload {
				// Gradient effect
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
			} else {
				barColor = theme.Secondary
			}
			barStyle := lipgloss.NewStyle().Foreground(barColor)
			barContent.WriteString(barStyle.Render("▓"))
		} else {
			barContent.WriteString(dimStyle.Render("░"))
		}
	}

	pctStr := fmt.Sprintf("%3.0f%%", fillPct*100)
	pctStyle := lipgloss.NewStyle().Foreground(theme.Text)

	return dimStyle.Render("⟨") + barContent.String() + dimStyle.Render("⟩") + " " + pctStyle.Render(pctStr)
}

// ============================================================================
// STYLE: RETRO - Classic terminal aesthetic
// ============================================================================
func renderSpeedRetro(width, height int, download, upload int64, theme Theme) string {
	digits, digitHeight := getDigitSize(width, height)

	green := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#33ff33"))
	dimGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("#116611"))
	brightGreen := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88ff88"))

	var lines []string

	// Retro header with scan line effect
	lines = append(lines, "")
	lines = append(lines, centerText(dimGreen.Render("╔════════════════════════════════════╗"), width))
	lines = append(lines, centerText(dimGreen.Render("║")+"     "+brightGreen.Render("NETWORK TRAFFIC MONITOR")+"     "+dimGreen.Render("║"), width))
	lines = append(lines, centerText(dimGreen.Render("║")+"          "+dimGreen.Render("[ ACTIVE ]")+"            "+dimGreen.Render("║"), width))
	lines = append(lines, centerText(dimGreen.Render("╠════════════════════════════════════╣"), width))
	lines = append(lines, "")

	// Download
	lines = append(lines, centerText(green.Render("RX <<<"), width))
	downloadRate, downloadUnit := formatRateComponents(download)
	bigDownload := renderBigNumber(downloadRate, digits, digitHeight)
	for _, line := range bigDownload {
		lines = append(lines, centerText(green.Render(line)+"  "+dimGreen.Render(downloadUnit), width))
	}
	lines = append(lines, "")

	// Retro bar with brackets
	barWidth := min(width*35/100, 35)
	downloadBar := renderRetroBar(download, int64(100*1024*1024), barWidth)
	lines = append(lines, centerText(green.Render(downloadBar), width))
	lines = append(lines, "")

	// Divider
	lines = append(lines, centerText(dimGreen.Render("├────────────────────────────────────┤"), width))
	lines = append(lines, "")

	// Upload
	lines = append(lines, centerText(green.Render("TX >>>"), width))
	uploadRate, uploadUnit := formatRateComponents(upload)
	bigUpload := renderBigNumber(uploadRate, digits, digitHeight)
	for _, line := range bigUpload {
		lines = append(lines, centerText(green.Render(line)+"  "+dimGreen.Render(uploadUnit), width))
	}
	lines = append(lines, "")

	uploadBar := renderRetroBar(upload, int64(100*1024*1024), barWidth)
	lines = append(lines, centerText(green.Render(uploadBar), width))

	lines = append(lines, "")
	lines = append(lines, centerText(dimGreen.Render("╚════════════════════════════════════╝"), width))

	return verticalCenter(lines, width, height)
}

func renderRetroBar(value, maxValue int64, width int) string {
	if maxValue == 0 {
		maxValue = 1
	}
	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}
	filled := int(fillPct * float64(width-2))
	return "[" + strings.Repeat("█", filled) + strings.Repeat("·", width-2-filled) + "]"
}

// ============================================================================
// STYLE: NEON - Cyberpunk aesthetic
// ============================================================================
func renderSpeedNeon(width, height int, download, upload int64, theme Theme) string {
	digits, digitHeight := getDigitSize(width, height)

	pink := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff00ff"))
	cyan := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ffff"))
	dimPink := lipgloss.NewStyle().Foreground(lipgloss.Color("#660066"))
	dimCyan := lipgloss.NewStyle().Foreground(lipgloss.Color("#006666"))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	var lines []string

	// Neon frame header
	lines = append(lines, "")
	lines = append(lines, centerText(pink.Render("╔")+dimPink.Render("══════════════════════════════")+pink.Render("╗"), width))
	lines = append(lines, centerText(pink.Render("║")+"    "+cyan.Render("◈ NETWORK SPEED ◈")+"      "+pink.Render("║"), width))
	lines = append(lines, centerText(pink.Render("╚")+dimPink.Render("══════════════════════════════")+pink.Render("╝"), width))
	lines = append(lines, "")

	// Download with neon glow effect
	lines = append(lines, centerText(cyan.Render("▼▼▼")+"  "+white.Render("DOWNLOAD")+"  "+cyan.Render("▼▼▼"), width))
	lines = append(lines, "")

	downloadRate, downloadUnit := formatRateComponents(download)
	bigDownload := renderBigNumber(downloadRate, digits, digitHeight)
	for _, line := range bigDownload {
		lines = append(lines, centerText(cyan.Render(line)+"   "+dimCyan.Render(downloadUnit), width))
	}
	lines = append(lines, "")

	// Neon gauge
	gaugeWidth := min(width*45/100, 40)
	downloadGauge := renderNeonGauge(download, int64(100*1024*1024), gaugeWidth, true)
	lines = append(lines, centerText(downloadGauge, width))

	// Neon separator
	lines = append(lines, "")
	lines = append(lines, centerText(dimPink.Render("═")+pink.Render("◆")+dimPink.Render("═════════════════════")+pink.Render("◆")+dimPink.Render("═"), width))
	lines = append(lines, "")

	// Upload with neon glow
	lines = append(lines, centerText(pink.Render("▲▲▲")+"  "+white.Render("UPLOAD")+"  "+pink.Render("▲▲▲"), width))
	lines = append(lines, "")

	uploadRate, uploadUnit := formatRateComponents(upload)
	bigUpload := renderBigNumber(uploadRate, digits, digitHeight)
	for _, line := range bigUpload {
		lines = append(lines, centerText(pink.Render(line)+"   "+dimPink.Render(uploadUnit), width))
	}
	lines = append(lines, "")

	uploadGauge := renderNeonGauge(upload, int64(100*1024*1024), gaugeWidth, false)
	lines = append(lines, centerText(uploadGauge, width))

	return verticalCenter(lines, width, height)
}

func renderNeonGauge(value, maxValue int64, width int, isDownload bool) string {
	if maxValue == 0 {
		maxValue = 1
	}
	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}
	filled := int(fillPct * float64(width-2))

	var fillColor, emptyColor lipgloss.Color
	if isDownload {
		fillColor = lipgloss.Color("#00ffff")
		emptyColor = lipgloss.Color("#003333")
	} else {
		fillColor = lipgloss.Color("#ff00ff")
		emptyColor = lipgloss.Color("#330033")
	}

	barStyle := lipgloss.NewStyle().Foreground(fillColor)
	dimStyle := lipgloss.NewStyle().Foreground(emptyColor)
	bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	pctStr := fmt.Sprintf(" %3.0f%%", fillPct*100)
	pctStyle := lipgloss.NewStyle().Bold(true).Foreground(fillColor)

	return bracketStyle.Render("⟦") + barStyle.Render(strings.Repeat("▰", filled)) + dimStyle.Render(strings.Repeat("▱", width-2-filled)) + bracketStyle.Render("⟧") + pctStyle.Render(pctStr)
}

// ============================================================================
// STYLE: COMPACT - Dense but elegant
// ============================================================================
func renderSpeedCompact(width, height int, download, upload int64, theme Theme) string {
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Text)

	var lines []string

	// Compact title
	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("╶───")+labelStyle.Render(" NETWORK ")+dimStyle.Render("───╴"), width))
	lines = append(lines, "")

	// Download with inline gauge
	downloadRate := formatRateLarge(download)
	downloadLabel := downloadStyle.Render("▾ DOWN ")
	downloadValue := downloadStyle.Render(fmt.Sprintf("%12s", downloadRate))
	lines = append(lines, centerText(downloadLabel+downloadValue, width))

	gaugeWidth := min(width*55/100, 45)
	downloadGauge := renderCompactGauge(download, int64(100*1024*1024), gaugeWidth, theme, true)
	lines = append(lines, centerText(downloadGauge, width))
	lines = append(lines, "")

	// Upload with inline gauge
	uploadRate := formatRateLarge(upload)
	uploadLabel := uploadStyle.Render("▴ UP   ")
	uploadValue := uploadStyle.Render(fmt.Sprintf("%12s", uploadRate))
	lines = append(lines, centerText(uploadLabel+uploadValue, width))

	uploadGauge := renderCompactGauge(upload, int64(100*1024*1024), gaugeWidth, theme, false)
	lines = append(lines, centerText(uploadGauge, width))

	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render("╶─────────────────────╴"), width))

	return verticalCenter(lines, width, height)
}

func renderCompactGauge(value, maxValue int64, width int, theme Theme, isDownload bool) string {
	if maxValue == 0 {
		maxValue = 1
	}
	fillPct := float64(value) / float64(maxValue)
	if fillPct > 1.0 {
		fillPct = 1.0
	}
	filled := int(fillPct * float64(width))

	var color lipgloss.Color
	if isDownload {
		color = theme.Accent
	} else {
		color = theme.Secondary
	}

	barStyle := lipgloss.NewStyle().Foreground(color)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	return barStyle.Render(strings.Repeat("━", filled)) + dimStyle.Render(strings.Repeat("─", width-filled))
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func renderBigNumber(numStr string, digits map[rune][]string, height int) []string {
	lines := make([]string, height)

	for _, char := range numStr {
		digit, ok := digits[char]
		if !ok {
			continue
		}
		for i := 0; i < height; i++ {
			if i < len(digit) {
				lines[i] += digit[i] + " "
			}
		}
	}

	return lines
}

func formatRateComponents(bytesPerSec int64) (string, string) {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d", bytesPerSec), "B/s"
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f", float64(bytesPerSec)/1024), "KB/s"
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.1f", float64(bytesPerSec)/(1024*1024)), "MB/s"
	} else {
		return fmt.Sprintf("%.2f", float64(bytesPerSec)/(1024*1024*1024)), "GB/s"
	}
}

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

func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-padding-textWidth)
}

func verticalCenter(lines []string, width, height int) string {
	contentHeight := len(lines)
	topPadding := (height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var result []string
	for i := 0; i < topPadding; i++ {
		result = append(result, strings.Repeat(" ", width))
	}
	result = append(result, lines...)

	for len(result) < height {
		result = append(result, strings.Repeat(" ", width))
	}

	if len(result) > height {
		result = result[:height]
	}

	return strings.Join(result, "\n")
}
