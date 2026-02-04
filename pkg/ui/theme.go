package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the UI
type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Dim         lipgloss.Color
	Background  lipgloss.Color
	Text        lipgloss.Color
	TextDim     lipgloss.Color
	Border      lipgloss.Color
	BorderFocus lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	MatrixChars []rune
	// Rain colors for matrix effect (darkest to brightest)
	Rain1       lipgloss.Color
	Rain2       lipgloss.Color
	Rain3       lipgloss.Color
	Rain4       lipgloss.Color
	Rain5       lipgloss.Color
}

// Available themes
var (
	ThemeMatrix = Theme{
		Name:        "Matrix",
		Primary:     lipgloss.Color("#4fff4f"),
		Secondary:   lipgloss.Color("#2fbf2f"),
		Accent:      lipgloss.Color("#aaffaa"),
		Dim:         lipgloss.Color("#0f3f0f"),
		Background:  lipgloss.Color("#000000"),
		Text:        lipgloss.Color("#4fff4f"),
		TextDim:     lipgloss.Color("#1a5f1a"),
		Border:      lipgloss.Color("#1a6f1a"),
		BorderFocus: lipgloss.Color("#4fff4f"),
		Success:     lipgloss.Color("#4fff4f"),
		Warning:     lipgloss.Color("#ffff4f"),
		Error:       lipgloss.Color("#ff4f4f"),
		MatrixChars: []rune("ΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩαβγδεζηθικλμνξοπρστυφχψω0123456789∆∇∂∑∏√∞≈≠±×÷"),
		// Smooth green gradient for rain
		Rain1:       lipgloss.Color("#001f00"),
		Rain2:       lipgloss.Color("#004f00"),
		Rain3:       lipgloss.Color("#00af00"),
		Rain4:       lipgloss.Color("#00df00"),
		Rain5:       lipgloss.Color("#aaffaa"),
	}

	ThemeCyberpunk = Theme{
		Name:        "Cyberpunk",
		Primary:     lipgloss.Color("#ff00ff"),
		Secondary:   lipgloss.Color("#00ffff"),
		Accent:      lipgloss.Color("#ff6ec7"),
		Dim:         lipgloss.Color("#6f1a6f"),
		Background:  lipgloss.Color("#0a0510"),
		Text:        lipgloss.Color("#ff00ff"),
		TextDim:     lipgloss.Color("#6f3f6f"),
		Border:      lipgloss.Color("#6f1a6f"),
		BorderFocus: lipgloss.Color("#00ffff"),
		Success:     lipgloss.Color("#00ff9f"),
		Warning:     lipgloss.Color("#ffff00"),
		Error:       lipgloss.Color("#ff0055"),
		MatrixChars: []rune("╔╗╚╝║═▓▒░█▄▀●○◐◑◒◓∞≈≠±×÷√∑∏∆ΩΨΦΘΞΛ01"),
		Rain1:       lipgloss.Color("#1a0020"),
		Rain2:       lipgloss.Color("#4a0060"),
		Rain3:       lipgloss.Color("#aa00aa"),
		Rain4:       lipgloss.Color("#00cccc"),
		Rain5:       lipgloss.Color("#ffffff"),
	}

	ThemeAmber = Theme{
		Name:        "Amber",
		Primary:     lipgloss.Color("#ffbf00"),
		Secondary:   lipgloss.Color("#ff8c00"),
		Accent:      lipgloss.Color("#ffd700"),
		Dim:         lipgloss.Color("#5f4f00"),
		Background:  lipgloss.Color("#0f0a00"),
		Text:        lipgloss.Color("#ffbf00"),
		TextDim:     lipgloss.Color("#5f4f00"),
		Border:      lipgloss.Color("#5f4f00"),
		BorderFocus: lipgloss.Color("#ffbf00"),
		Success:     lipgloss.Color("#7fff00"),
		Warning:     lipgloss.Color("#ff6600"),
		Error:       lipgloss.Color("#ff3300"),
		MatrixChars: []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&*<>[]{}"),
		Rain1:       lipgloss.Color("#1a1000"),
		Rain2:       lipgloss.Color("#4a3000"),
		Rain3:       lipgloss.Color("#8a6000"),
		Rain4:       lipgloss.Color("#cc9000"),
		Rain5:       lipgloss.Color("#ffdd66"),
	}

	ThemeOcean = Theme{
		Name:        "Ocean",
		Primary:     lipgloss.Color("#00bfff"),
		Secondary:   lipgloss.Color("#0080ff"),
		Accent:      lipgloss.Color("#00ffff"),
		Dim:         lipgloss.Color("#004060"),
		Background:  lipgloss.Color("#000a10"),
		Text:        lipgloss.Color("#00bfff"),
		TextDim:     lipgloss.Color("#004060"),
		Border:      lipgloss.Color("#004060"),
		BorderFocus: lipgloss.Color("#00ffff"),
		Success:     lipgloss.Color("#00ff7f"),
		Warning:     lipgloss.Color("#ffa500"),
		Error:       lipgloss.Color("#ff4444"),
		MatrixChars: []rune("~≈∼∽≀≁∿⊰⊱⋰⋱⋮⋯░▒▓█▀▄◢◣◤◥○●◐◑01"),
		Rain1:       lipgloss.Color("#001020"),
		Rain2:       lipgloss.Color("#003050"),
		Rain3:       lipgloss.Color("#0060a0"),
		Rain4:       lipgloss.Color("#00a0e0"),
		Rain5:       lipgloss.Color("#88ffff"),
	}

	ThemeBlood = Theme{
		Name:        "Blood",
		Primary:     lipgloss.Color("#ff0040"),
		Secondary:   lipgloss.Color("#cc0033"),
		Accent:      lipgloss.Color("#ff3366"),
		Dim:         lipgloss.Color("#4f0015"),
		Background:  lipgloss.Color("#0a0005"),
		Text:        lipgloss.Color("#ff0040"),
		TextDim:     lipgloss.Color("#4f0015"),
		Border:      lipgloss.Color("#4f0015"),
		BorderFocus: lipgloss.Color("#ff0040"),
		Success:     lipgloss.Color("#00ff00"),
		Warning:     lipgloss.Color("#ff8800"),
		Error:       lipgloss.Color("#ff0000"),
		MatrixChars: []rune("†‡§¶∞≠≈×÷±√∑∏ΩΨΦΔΛΞ666BLOODDEATH01"),
		Rain1:       lipgloss.Color("#100005"),
		Rain2:       lipgloss.Color("#400015"),
		Rain3:       lipgloss.Color("#880030"),
		Rain4:       lipgloss.Color("#cc0040"),
		Rain5:       lipgloss.Color("#ff6688"),
	}

	// All available themes
	Themes = []Theme{ThemeMatrix, ThemeCyberpunk, ThemeAmber, ThemeOcean, ThemeBlood}
)

// CurrentTheme is the active theme
var CurrentTheme = ThemeMatrix

// CycleTheme switches to the next theme
func CycleTheme() Theme {
	for i, t := range Themes {
		if t.Name == CurrentTheme.Name {
			CurrentTheme = Themes[(i+1)%len(Themes)]
			return CurrentTheme
		}
	}
	CurrentTheme = Themes[0]
	return CurrentTheme
}

// SetTheme sets the theme by name (case-insensitive)
func SetTheme(name string) {
	nameLower := strings.ToLower(name)
	for _, t := range Themes {
		if strings.ToLower(t.Name) == nameLower {
			CurrentTheme = t
			return
		}
	}
}
