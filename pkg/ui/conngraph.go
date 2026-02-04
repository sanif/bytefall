package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// connEntry represents a domain connEntry with bytes
type connEntry struct {
	domain string
	bytes  int64
}

// ConnectionGraphPanel displays a visual graph of process-to-domain connEntrys
type ConnectionGraphPanel struct {
	width   int
	height  int
	focused bool
	stats   []*data.DomainStats
	tick    int
}

// NewConnectionGraphPanel creates a new connEntry graph panel
func NewConnectionGraphPanel(width, height int) *ConnectionGraphPanel {
	return &ConnectionGraphPanel{
		width:  width,
		height: height,
	}
}

// Resize updates the panel dimensions
func (c *ConnectionGraphPanel) Resize(width, height int) {
	c.width = width
	c.height = height
}

// SetFocused sets the focus state
func (c *ConnectionGraphPanel) SetFocused(focused bool) {
	c.focused = focused
}

// UpdateStats updates the connEntry data
func (c *ConnectionGraphPanel) UpdateStats(stats []*data.DomainStats) {
	c.stats = stats
	c.tick++
}

// View renders the connection graph using Lipgloss components
func (c *ConnectionGraphPanel) View() string {
	theme := CurrentTheme

	// Build process -> domain mapping
	processConnections := make(map[string][]connEntry)

	for _, s := range c.stats {
		for proc := range s.Processes {
			processConnections[proc] = append(processConnections[proc], connEntry{s.Domain, s.Bytes})
		}
	}

	// Sort processes by total bytes
	type procEntry struct {
		name       string
		totalBytes int64
		conns      []connEntry
	}
	var procs []procEntry
	for proc, conns := range processConnections {
		var total int64
		for _, conn := range conns {
			total += conn.bytes
		}
		procs = append(procs, procEntry{proc, total, conns})
	}
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].totalBytes > procs[j].totalBytes
	})

	// If no connEntrys, show empty state
	if len(procs) == 0 {
		return c.renderEmptyState(theme)
	}

	// Calculate layout
	availableHeight := c.height - 4
	procBoxHeight := 6
	maxProcs := availableHeight / procBoxHeight
	if maxProcs < 1 {
		maxProcs = 1
	}
	if len(procs) > maxProcs {
		procs = procs[:maxProcs]
	}

	// Render each process as a styled box
	var processBoxes []string
	for _, proc := range procs {
		box := c.renderProcessCard(proc.name, proc.totalBytes, proc.conns, theme)
		processBoxes = append(processBoxes, box)
	}

	// Join all boxes vertically with spacing
	content := lipgloss.JoinVertical(lipgloss.Left, processBoxes...)

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)

	title := titleStyle.Render("  Network Connections")

	// Wrap in container
	containerStyle := lipgloss.NewStyle().
		Width(c.width).
		Height(c.height)

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (c *ConnectionGraphPanel) renderProcessCard(name string, totalBytes int64, conns []connEntry, theme Theme) string {
	// Process card style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(0, 1).
		MarginBottom(1).
		Width(c.width - 4)

	// Process name style
	procNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Secondary)

	// Bytes style
	bytesStyle := lipgloss.NewStyle().
		Foreground(theme.Accent)

	// Domain style
	domainStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	// Dim style
	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	// Header: process name + total bytes
	procName := name
	if len(procName) > 25 {
		procName = procName[:22] + "..."
	}
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		procNameStyle.Render(procName),
		dimStyle.Render(" ─ "),
		bytesStyle.Render(formatBytes(totalBytes)),
	)

	// Sort connEntrys
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].bytes > conns[j].bytes
	})

	// Limit domains
	maxDomains := 3
	if len(conns) > maxDomains {
		conns = conns[:maxDomains]
	}

	// Render domains
	var domainLines []string
	for _, conn := range conns {
		domainName := conn.domain
		if len(domainName) > 35 {
			domainName = domainName[:32] + "..."
		}

		// Traffic bar
		bar := c.renderTrafficBar(conn.bytes, theme)

		line := fmt.Sprintf("  %s %s %s",
			dimStyle.Render("→"),
			domainStyle.Render(fmt.Sprintf("%-36s", domainName)),
			bar)
		domainLines = append(domainLines, line)
	}

	// Combine header and domains
	cardContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(domainLines, "\n"),
	)

	return cardStyle.Render(cardContent)
}

func (c *ConnectionGraphPanel) renderTrafficBar(bytes int64, theme Theme) string {
	// Calculate bar length (0-10)
	var level int
	if bytes > 10000000 {
		level = 10
	} else if bytes > 5000000 {
		level = 8
	} else if bytes > 1000000 {
		level = 6
	} else if bytes > 500000 {
		level = 5
	} else if bytes > 100000 {
		level = 4
	} else if bytes > 50000 {
		level = 3
	} else if bytes > 10000 {
		level = 2
	} else if bytes > 1000 {
		level = 1
	}

	// Bar colors based on level
	var barColor lipgloss.Color
	if level >= 8 {
		barColor = theme.Rain5
	} else if level >= 6 {
		barColor = theme.Rain4
	} else if level >= 4 {
		barColor = theme.Rain3
	} else if level >= 2 {
		barColor = theme.Rain2
	} else {
		barColor = theme.Rain1
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	bytesStyle := lipgloss.NewStyle().Foreground(theme.TextDim)

	filled := strings.Repeat("█", level)
	empty := strings.Repeat("░", 10-level)

	return barStyle.Render(filled) + dimStyle.Render(empty) + " " + bytesStyle.Render(formatBytes(bytes))
}

func (c *ConnectionGraphPanel) renderEmptyState(theme Theme) string {
	emptyStyle := lipgloss.NewStyle().
		Width(c.width).
		Height(c.height).
		Align(lipgloss.Center, lipgloss.Center)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Dim).
		Padding(2, 4).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		MarginBottom(1)

	textStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("No Active Connections"),
		textStyle.Render("Process mapping requires"),
		textStyle.Render("sudo/root privileges"),
	)

	return emptyStyle.Render(boxStyle.Render(content))
}
