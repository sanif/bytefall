package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// AnimatedBorder creates borders with animation effects
type AnimatedBorder struct {
	tick    int
	focused bool
}

// NewAnimatedBorder creates a new animated border
func NewAnimatedBorder() *AnimatedBorder {
	return &AnimatedBorder{}
}

// Update advances the animation
func (ab *AnimatedBorder) Update() {
	ab.tick++
}

// SetFocused sets the focus state
func (ab *AnimatedBorder) SetFocused(focused bool) {
	ab.focused = focused
}

// GetBorderStyle returns the current border style with animation
func (ab *AnimatedBorder) GetBorderStyle(theme Theme) lipgloss.Style {
	if !ab.focused {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border)
	}

	// Animated pulsing for focused border
	pulseColors := []lipgloss.Color{
		theme.BorderFocus,
		theme.Accent,
		theme.Primary,
		theme.Accent,
	}

	colorIdx := (ab.tick / 3) % len(pulseColors)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(pulseColors[colorIdx])
}

// GlowBorder creates a glowing border effect
func GlowBorder(content string, width, height int, theme Theme, intensity int) string {
	// Glow characters based on intensity
	glowChars := []string{"░", "▒", "▓", "█"}
	glowIdx := intensity % len(glowChars)
	glowChar := glowChars[glowIdx]

	glowStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	brightStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	// Build glowing frame
	topBottom := brightStyle.Render("╭") +
		brightStyle.Render(strings.Repeat("─", width-2)) +
		brightStyle.Render("╮")

	glowTop := glowStyle.Render(glowChar) +
		glowStyle.Render(strings.Repeat(glowChar, width)) +
		glowStyle.Render(glowChar)

	var result strings.Builder
	result.WriteString(glowTop + "\n")
	result.WriteString(glowStyle.Render(glowChar) + topBottom + glowStyle.Render(glowChar) + "\n")

	lines := strings.Split(content, "\n")
	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		// Pad line to width
		lineWidth := lipgloss.Width(line)
		if lineWidth < width-2 {
			line += strings.Repeat(" ", width-2-lineWidth)
		}
		result.WriteString(glowStyle.Render(glowChar) +
			brightStyle.Render("│") + line + brightStyle.Render("│") +
			glowStyle.Render(glowChar) + "\n")
	}

	bottomBorder := brightStyle.Render("╰") +
		brightStyle.Render(strings.Repeat("─", width-2)) +
		brightStyle.Render("╯")
	result.WriteString(glowStyle.Render(glowChar) + bottomBorder + glowStyle.Render(glowChar) + "\n")
	result.WriteString(glowTop)

	return result.String()
}

// PulsingText creates text that pulses between colors
func PulsingText(text string, theme Theme) string {
	phase := int(time.Now().UnixMilli()/200) % 4
	colors := []lipgloss.Color{theme.Primary, theme.Accent, theme.Secondary, theme.Accent}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors[phase])

	return style.Render(text)
}

// ScanlineEffect adds retro scanline effect to content
func ScanlineEffect(content string, theme Theme) string {
	lines := strings.Split(content, "\n")
	var result []string

	scanStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	for i, line := range lines {
		if i%2 == 1 {
			// Dim every other line slightly
			result = append(result, scanStyle.Render(line))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
