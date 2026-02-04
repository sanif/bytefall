package ui

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/data"
)

// DomainInfoPopup shows detailed domain information
type DomainInfoPopup struct {
	visible bool
	domain  *data.DomainStats
	ipInfo  *IPInfo
	loading bool
}

// IPInfo holds resolved IP information
type IPInfo struct {
	IPs       []string
	Hostnames []string
	Resolved  time.Time
}

// NewDomainInfoPopup creates a new domain info popup
func NewDomainInfoPopup() *DomainInfoPopup {
	return &DomainInfoPopup{}
}

// Show displays the popup for a domain
func (d *DomainInfoPopup) Show(domain *data.DomainStats) {
	d.visible = true
	d.domain = domain
	d.loading = true

	// Resolve IP info in background
	go d.resolveInfo()
}

// Hide closes the popup
func (d *DomainInfoPopup) Hide() {
	d.visible = false
	d.domain = nil
	d.ipInfo = nil
}

// IsVisible returns visibility state
func (d *DomainInfoPopup) IsVisible() bool {
	return d.visible
}

func (d *DomainInfoPopup) resolveInfo() {
	if d.domain == nil {
		return
	}

	info := &IPInfo{Resolved: time.Now()}

	// Resolve IPs
	ips, err := net.LookupIP(d.domain.Domain)
	if err == nil {
		for _, ip := range ips {
			info.IPs = append(info.IPs, ip.String())
		}
	}

	// Reverse lookup for hostnames
	if len(info.IPs) > 0 {
		hosts, err := net.LookupAddr(info.IPs[0])
		if err == nil {
			info.Hostnames = hosts
		}
	}

	d.ipInfo = info
	d.loading = false
}

// View renders the popup
func (d *DomainInfoPopup) View(termWidth, termHeight int, theme Theme) string {
	if !d.visible || d.domain == nil {
		return ""
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Dim).
		Padding(0, 2)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.TextDim)

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Dim)

	// Build content
	var lines []string

	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("🔍 DOMAIN DETAILS"))
	lines = append(lines, "")

	// Domain name
	lines = append(lines, labelStyle.Render("  Domain: ")+valueStyle.Render(d.domain.Domain))
	lines = append(lines, "")

	// Traffic stats
	lines = append(lines, labelStyle.Render("  ── Traffic ──────────────────"))
	lines = append(lines, labelStyle.Render("  Bytes:    ")+valueStyle.Render(formatBytesLong(d.domain.Bytes)))
	lines = append(lines, labelStyle.Render("  Packets:  ")+valueStyle.Render(fmt.Sprintf("%d", d.domain.Packets)))
	lines = append(lines, labelStyle.Render("  Last seen: ")+valueStyle.Render(d.domain.LastSeen.Format("15:04:05")))
	lines = append(lines, "")

	// Processes
	lines = append(lines, labelStyle.Render("  ── Processes ────────────────"))
	if len(d.domain.Processes) > 0 {
		procs := make([]string, 0, len(d.domain.Processes))
		for p := range d.domain.Processes {
			procs = append(procs, p)
		}
		for i, p := range procs {
			if i >= 5 {
				lines = append(lines, labelStyle.Render(fmt.Sprintf("  ... and %d more", len(procs)-5)))
				break
			}
			lines = append(lines, labelStyle.Render("  • ")+valueStyle.Render(p))
		}
	} else {
		lines = append(lines, dimStyle.Render("  (no process info)"))
	}
	lines = append(lines, "")

	// IP Info
	lines = append(lines, labelStyle.Render("  ── Network ──────────────────"))
	if d.loading {
		lines = append(lines, dimStyle.Render("  Resolving..."))
	} else if d.ipInfo != nil {
		if len(d.ipInfo.IPs) > 0 {
			for i, ip := range d.ipInfo.IPs {
				if i >= 3 {
					lines = append(lines, dimStyle.Render(fmt.Sprintf("  ... +%d more IPs", len(d.ipInfo.IPs)-3)))
					break
				}
				lines = append(lines, labelStyle.Render("  IP: ")+valueStyle.Render(ip))
			}
		}
		if len(d.ipInfo.Hostnames) > 0 {
			lines = append(lines, labelStyle.Render("  Host: ")+valueStyle.Render(d.ipInfo.Hostnames[0]))
		}
	}
	lines = append(lines, "")

	// Sparkline
	lines = append(lines, labelStyle.Render("  ── Activity (60s) ───────────"))
	lines = append(lines, "  "+d.renderMiniSparkline(theme))
	lines = append(lines, "")

	// Help
	lines = append(lines, dimStyle.Render("  [ESC] close  [b] block domain"))
	lines = append(lines, "")

	// Create box
	contentWidth := 44
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.Primary).
		Background(theme.Background).
		Width(contentWidth).
		Padding(0, 1)

	content := strings.Join(lines, "\n")
	box := boxStyle.Render(content)

	// Center
	boxLines := strings.Split(box, "\n")
	boxHeight := len(boxLines)
	boxWidth := lipgloss.Width(box)

	padTop := (termHeight - boxHeight) / 2
	padLeft := (termWidth - boxWidth) / 2

	if padTop < 0 {
		padTop = 0
	}
	if padLeft < 0 {
		padLeft = 0
	}

	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color("#050505"))

	var output strings.Builder

	for i := 0; i < padTop; i++ {
		output.WriteString(bgStyle.Render(strings.Repeat(" ", termWidth)))
		output.WriteString("\n")
	}

	for i, line := range boxLines {
		leftPad := bgStyle.Render(strings.Repeat(" ", padLeft))
		rightPadWidth := termWidth - padLeft - lipgloss.Width(line)
		if rightPadWidth < 0 {
			rightPadWidth = 0
		}
		rightPad := bgStyle.Render(strings.Repeat(" ", rightPadWidth))
		output.WriteString(leftPad + line + rightPad)
		if i < len(boxLines)-1 {
			output.WriteString("\n")
		}
	}

	remainingHeight := termHeight - padTop - boxHeight
	for i := 0; i < remainingHeight; i++ {
		output.WriteString("\n")
		output.WriteString(bgStyle.Render(strings.Repeat(" ", termWidth)))
	}

	return output.String()
}

func (d *DomainInfoPopup) renderMiniSparkline(theme Theme) string {
	if d.domain == nil || len(d.domain.History) == 0 {
		return strings.Repeat("▁", 30)
	}

	sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var maxVal int64
	for _, v := range d.domain.History {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return lipgloss.NewStyle().Foreground(theme.Dim).Render(strings.Repeat("▁", 30))
	}

	var spark strings.Builder
	width := 30
	for i := 0; i < width; i++ {
		idx := i * len(d.domain.History) / width
		if idx >= len(d.domain.History) {
			idx = len(d.domain.History) - 1
		}

		val := d.domain.History[idx]
		level := int(val * int64(len(sparkChars)-1) / maxVal)
		if level < 0 {
			level = 0
		}
		if level >= len(sparkChars) {
			level = len(sparkChars) - 1
		}

		spark.WriteRune(sparkChars[level])
	}

	return lipgloss.NewStyle().Foreground(theme.Primary).Render(spark.String())
}

func formatBytesLong(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB (%d bytes)", float64(b)/float64(div), "KMGTPE"[exp], b)
}
