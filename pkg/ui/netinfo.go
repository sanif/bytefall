package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// NetInfoBar displays network information in a status bar
type NetInfoBar struct {
	width int
	info  data.NetworkInfo
}

// NewNetInfoBar creates a new network info bar
func NewNetInfoBar(width int) *NetInfoBar {
	return &NetInfoBar{width: width}
}

// UpdateInfo sets the network information
func (n *NetInfoBar) UpdateInfo(info data.NetworkInfo) {
	n.info = info
}

// Resize updates the width
func (n *NetInfoBar) Resize(width int) {
	n.width = width
}

// View renders the network info bar
func (n *NetInfoBar) View() string {
	theme := CurrentTheme

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	accentStyle := lipgloss.NewStyle().
		Foreground(theme.Accent)

	downStyle := lipgloss.NewStyle().
		Foreground(theme.Primary)

	sepStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	sep := sepStyle.Render(" │ ")

	// Build info segments
	var segments []string

	// Interface + IP combined
	if n.info.Interface != "" && n.info.IPv4 != "" {
		segments = append(segments,
			labelStyle.Render("⌘ ")+valueStyle.Render(n.info.Interface)+
				labelStyle.Render(":")+valueStyle.Render(n.info.IPv4))
	} else if n.info.Interface != "" {
		segments = append(segments,
			labelStyle.Render("⌘ ")+valueStyle.Render(n.info.Interface))
	}

	// Public IP (shortened)
	if n.info.PublicIP != "" {
		ip := n.info.PublicIP
		if len(ip) > 15 {
			ip = ip[:12] + "..."
		}
		segments = append(segments,
			labelStyle.Render("☁ ")+accentStyle.Render(ip))
	}

	// Traffic rates with arrows
	downRate := formatRate(n.info.BytesPerSec)
	segments = append(segments,
		downStyle.Render("↓")+valueStyle.Render(downRate))

	// Packets per second
	if n.info.PacketsPerSec > 0 {
		segments = append(segments,
			labelStyle.Render("◈ ")+valueStyle.Render(fmt.Sprintf("%d/s", n.info.PacketsPerSec)))
	}

	return strings.Join(segments, sep)
}

func formatRate(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	} else {
		return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/(1024*1024))
	}
}
