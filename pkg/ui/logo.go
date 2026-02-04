package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ASCII art logos
var logoLarge = `
██████╗ ██╗   ██╗████████╗███████╗███████╗ █████╗ ██╗     ██╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██╔════╝██╔══██╗██║     ██║
██████╔╝ ╚████╔╝    ██║   █████╗  █████╗  ███████║██║     ██║
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██╔══╝  ██╔══██║██║     ██║
██████╔╝   ██║      ██║   ███████╗██║     ██║  ██║███████╗███████╗
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚═╝     ╚═╝  ╚═╝╚══════╝╚══════╝`

var logoMedium = `
▄▄▄▄   ▓██   ██▓▄▄▄█████▓▓█████   █████▒▄▄▄       ██▓     ██▓
▓█████▄ ▒██  ██▒▓  ██▒ ▓▒▓█   ▀ ▓██   ▒▒████▄    ▓██▒    ▓██▒
▒██▒ ▄██ ▒██ ██░▒ ▓██░ ▒░▒███   ▒████ ░▒██  ▀█▄  ▒██░    ▒██░
▒██░█▀   ░ ▐██▓░░ ▓██▓ ░ ▒▓█  ▄ ░▓█▒  ░░██▄▄▄▄██ ▒██░    ▒██░
░▓█  ▀█▓ ░ ██▒▓░  ▒██▒ ░ ░▒████▒░▒█░    ▓█   ▓██▒░██████▒░██████▒
░▒▓███▀▒  ██▒▒▒   ▒ ░░   ░░ ▒░ ░ ▒ ░    ▒▒   ▓▒█░░ ▒░▓  ░░ ▒░▓  ░
▒░▒   ░ ▓██ ░▒░     ░     ░ ░  ░ ░       ▒   ▒▒ ░░ ░ ▒  ░░ ░ ▒  ░
 ░    ░ ▒ ▒ ░░    ░         ░    ░ ░     ░   ▒     ░ ░     ░ ░
 ░      ░ ░                 ░  ░             ░  ░    ░  ░    ░  ░
      ░ ░ ░`

var logoSmall = `
┏┓ ┓┏┏┳┓┏━┓┏━┓┏━┓┓  ┓
┣┫ ┗┫ ┃ ┣━ ┣━ ┣━┫┃  ┃
┗┛ ┗┛ ┻ ┗━┛┻  ┛ ┗┗━┛┗━┛`

var logoMini = `▀█▀ BYTEFALL █▀`

// RenderLogo returns the appropriately sized logo for the terminal width
func RenderLogo(width int, theme Theme) string {
	var logo string

	if width >= 75 {
		logo = logoLarge
	} else if width >= 60 {
		logo = logoMedium
	} else if width >= 30 {
		logo = logoSmall
	} else {
		logo = logoMini
	}

	// Apply gradient coloring
	lines := strings.Split(logo, "\n")
	var result []string

	colors := []lipgloss.Color{
		theme.Accent,
		theme.Primary,
		theme.Secondary,
		theme.Dim,
	}

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		colorIdx := i % len(colors)
		style := lipgloss.NewStyle().Foreground(colors[colorIdx])
		result = append(result, style.Render(line))
	}

	return strings.Join(result, "\n")
}

// RenderLogoCompact returns a single-line logo
func RenderLogoCompact(theme Theme) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	accentStyle := lipgloss.NewStyle().
		Foreground(theme.Accent)

	return accentStyle.Render("◆ ") + style.Render("BYTEFALL") + accentStyle.Render(" ◆")
}
