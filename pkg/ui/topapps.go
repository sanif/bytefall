package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// App icons/emojis for common applications
var appIcons = map[string]string{
	"chrome":          "🌐",
	"google chrome":   "🌐",
	"firefox":         "🦊",
	"safari":          "🧭",
	"edge":            "🌐",
	"brave":           "🦁",
	"slack":           "💬",
	"discord":         "🎮",
	"telegram":        "✈️",
	"whatsapp":        "📱",
	"zoom":            "📹",
	"teams":           "👥",
	"spotify":         "🎵",
	"code":            "💻",
	"cursor":          "💻",
	"vscode":          "💻",
	"terminal":        "⌨️",
	"iterm":           "⌨️",
	"iterm2":          "⌨️",
	"wezterm":         "⌨️",
	"ssh":             "🔐",
	"git":             "📦",
	"node":            "📗",
	"python":          "🐍",
	"python3":         "🐍",
	"go":              "🐹",
	"docker":          "🐳",
	"kubectl":         "☸️",
	"curl":            "🔗",
	"wget":            "⬇️",
	"dropbox":         "📦",
	"onedrive":        "☁️",
	"icloud":          "☁️",
	"mail":            "📧",
	"outlook":         "📧",
	"finder":          "📁",
	"xcode":           "🔨",
	"notion":          "📝",
	"obsidian":        "💎",
	"figma":           "🎨",
	"photoshop":       "🖼️",
	"illustrator":     "✏️",
	"premiere":        "🎬",
	"vlc":             "🎬",
	"quicktime":       "🎬",
	"steam":           "🎮",
	"minecraft":       "⛏️",
	"default":         "◆",
}

// TopAppsWidget renders full-screen top applications by bandwidth
func RenderTopAppsWidget(width, height int, stats []*data.DomainStats, theme Theme) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	headerStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	appStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	bytesStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	// Build process -> bytes map
	appBytes := make(map[string]int64)
	appDomains := make(map[string]map[string]bool)

	for _, s := range stats {
		for proc := range s.Processes {
			appBytes[proc] += s.Bytes
			if appDomains[proc] == nil {
				appDomains[proc] = make(map[string]bool)
			}
			appDomains[proc][s.Domain] = true
		}
	}

	// Sort by bytes
	type appEntry struct {
		name    string
		bytes   int64
		domains int
	}
	var apps []appEntry
	for name, bytes := range appBytes {
		apps = append(apps, appEntry{name, bytes, len(appDomains[name])})
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].bytes > apps[j].bytes
	})

	// Find max bytes for bar scaling
	var maxBytes int64
	if len(apps) > 0 {
		maxBytes = apps[0].bytes
	}
	if maxBytes == 0 {
		maxBytes = 1
	}

	// Calculate layout
	barWidth := width * 40 / 100
	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > 50 {
		barWidth = 50
	}

	var lines []string

	// Title
	titleLine := dimStyle.Render("═══════") + "  " + titleStyle.Render("◆ TOP APPLICATIONS ◆") + "  " + dimStyle.Render("═══════")
	lines = append(lines, centerText(titleLine, width))
	lines = append(lines, "")

	// Header
	header := fmt.Sprintf("  %-3s %-20s %12s %8s  %-*s",
		"", "APPLICATION", "TRAFFIC", "DOMAINS", barWidth, "")
	lines = append(lines, centerText(headerStyle.Render(header), width))
	lines = append(lines, centerText(dimStyle.Render(strings.Repeat("─", 60+barWidth)), width))

	// App rows
	maxApps := height - 8
	if maxApps > len(apps) {
		maxApps = len(apps)
	}

	for i := 0; i < maxApps; i++ {
		app := apps[i]

		// Get icon
		icon := getAppIcon(app.name)

		// Truncate name
		name := app.name
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		// Create bar
		bar := renderAppBar(app.bytes, maxBytes, barWidth, theme, i)

		// Format row
		rank := fmt.Sprintf("%2d.", i+1)
		if i == 0 {
			rank = bytesStyle.Render(" ★")
		} else if i < 3 {
			rank = dimStyle.Render(fmt.Sprintf("%2d.", i+1))
		} else {
			rank = dimStyle.Render(fmt.Sprintf("%2d.", i+1))
		}

		row := fmt.Sprintf("  %s %s %-18s %12s %8d  %s",
			rank,
			icon,
			appStyle.Render(name),
			bytesStyle.Render(formatBytes(app.bytes)),
			app.domains,
			bar,
		)
		lines = append(lines, centerText(row, width))
	}

	// Empty state
	if len(apps) == 0 {
		lines = append(lines, "")
		lines = append(lines, centerText(dimStyle.Render("No application data"), width))
		lines = append(lines, centerText(dimStyle.Render("(requires sudo for process mapping)"), width))
	}

	// Summary footer
	lines = append(lines, "")
	lines = append(lines, centerText(dimStyle.Render(strings.Repeat("─", 60+barWidth)), width))

	var totalBytes int64
	for _, app := range apps {
		totalBytes += app.bytes
	}
	summary := fmt.Sprintf("Total: %d apps  •  %s transferred", len(apps), formatBytes(totalBytes))
	lines = append(lines, centerText(dimStyle.Render(summary), width))

	return verticalCenter(lines, width, height)
}

func getAppIcon(appName string) string {
	name := strings.ToLower(appName)

	// Check exact match first
	if icon, ok := appIcons[name]; ok {
		return icon
	}

	// Check partial matches
	for key, icon := range appIcons {
		if strings.Contains(name, key) {
			return icon
		}
	}

	return appIcons["default"]
}

func renderAppBar(value, maxValue int64, width int, theme Theme, rank int) string {
	if maxValue == 0 {
		maxValue = 1
	}

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

	// Color based on rank
	var barColor lipgloss.Color
	switch rank {
	case 0:
		barColor = theme.Accent
	case 1:
		barColor = theme.Rain5
	case 2:
		barColor = theme.Rain4
	default:
		barColor = theme.Rain3
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	return barStyle.Render(strings.Repeat("█", filled)) + dimStyle.Render(strings.Repeat("░", width-filled))
}
