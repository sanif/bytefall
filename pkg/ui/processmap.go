package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// ProcessMapPanel shows process to domain relationships
type ProcessMapPanel struct {
	width   int
	height  int
	stats   []*data.DomainStats
	focused bool
}

// NewProcessMapPanel creates a new process map panel
func NewProcessMapPanel(width, height int) *ProcessMapPanel {
	return &ProcessMapPanel{
		width:  width,
		height: height,
	}
}

// UpdateStats sets the domain statistics
func (p *ProcessMapPanel) UpdateStats(stats []*data.DomainStats) {
	p.stats = stats
}

// SetFocused sets the focus state
func (p *ProcessMapPanel) SetFocused(focused bool) {
	p.focused = focused
}

// Resize updates dimensions
func (p *ProcessMapPanel) Resize(width, height int) {
	p.width = width
	p.height = height
}

// View renders the process map
func (p *ProcessMapPanel) View() string {
	theme := CurrentTheme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	processStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Secondary)

	domainStyle := lipgloss.NewStyle().
		Foreground(theme.Text)

	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	connectorStyle := lipgloss.NewStyle().
		Foreground(theme.Border)

	// Build process -> domains map with byte counts
	type procDomain struct {
		domain string
		bytes  int64
	}
	procMap := make(map[string][]procDomain)
	procBytes := make(map[string]int64)

	for _, s := range p.stats {
		for proc := range s.Processes {
			procMap[proc] = append(procMap[proc], procDomain{s.Domain, s.Bytes})
			procBytes[proc] += s.Bytes
		}
	}

	// Sort processes by total bytes
	type procEntry struct {
		name    string
		bytes   int64
		domains []procDomain
	}
	var processes []procEntry
	for proc, domains := range procMap {
		processes = append(processes, procEntry{proc, procBytes[proc], domains})
	}
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].bytes > processes[j].bytes
	})

	// Build lines
	lines := make([]string, p.height)

	// Title
	lines[0] = titleStyle.Render("  Process Map")
	lines[1] = dimStyle.Render(strings.Repeat("─", p.width))

	// Fill entries
	lineIdx := 2
	for _, proc := range processes {
		if lineIdx >= p.height-1 {
			break
		}

		// Sort domains by bytes
		sort.Slice(proc.domains, func(i, j int) bool {
			return proc.domains[i].bytes > proc.domains[j].bytes
		})

		procName := proc.name
		if len(procName) > 12 {
			procName = procName[:9] + "..."
		}

		// Process line with icon
		procLine := fmt.Sprintf("  %s %s",
			processStyle.Render("▸"),
			processStyle.Render(procName))
		lines[lineIdx] = padRight(procLine, p.width)
		lineIdx++

		// Domain lines with tree connectors
		maxDomains := 3
		if len(proc.domains) > maxDomains {
			proc.domains = proc.domains[:maxDomains]
		}

		for i, d := range proc.domains {
			if lineIdx >= p.height {
				break
			}

			domain := d.domain
			maxDomainLen := p.width - 12
			if maxDomainLen < 15 {
				maxDomainLen = 15
			}
			if len(domain) > maxDomainLen {
				domain = domain[:maxDomainLen-2] + ".."
			}

			// Tree connector
			connector := "├─"
			if i == len(proc.domains)-1 {
				connector = "└─"
			}

			line := fmt.Sprintf("    %s %s",
				connectorStyle.Render(connector),
				domainStyle.Render(domain))
			lines[lineIdx] = padRight(line, p.width)
			lineIdx++
		}
	}

	// Empty state
	if len(processes) == 0 && lineIdx < p.height {
		lines[lineIdx] = dimStyle.Render("  No process data")
		lineIdx++
		lines[lineIdx] = dimStyle.Render("  (requires sudo)")
	}

	// Fill remaining with empty
	for i := range lines {
		if lines[i] == "" {
			lines[i] = strings.Repeat(" ", p.width)
		}
	}

	return strings.Join(lines, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
