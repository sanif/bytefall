package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sanif/bytefall/pkg/capture"
	"github.com/sanif/bytefall/pkg/data"
)

// matrixOnlyMode shows only the matrix animation
var matrixOnlyMode bool

// MatrixOptions configures what to show in matrix mode
type MatrixOptions struct {
	ShowInfo     bool // Show status bar
	ShowDownload bool // Show download speed
	ShowUpload   bool // Show upload speed
	ShowDomains  bool // Show domain count
	ShowIP       bool // Show local IP address
	ShowPublicIP bool // Show public IP address
	Minimal      bool // No status bar at all
}

var matrixOpts = MatrixOptions{
	ShowInfo:     true,
	ShowDownload: true,
	ShowUpload:   true,
	ShowDomains:  true,
	ShowIP:       false,
	ShowPublicIP: false,
	Minimal:      false,
}

// SetMatrixOnly enables or disables matrix-only mode
func SetMatrixOnly(enabled bool) {
	matrixOnlyMode = enabled
}

// SetMatrixOptions sets display options for matrix mode
func SetMatrixOptions(opts MatrixOptions) {
	matrixOpts = opts
}

// FocusPanel indicates which panel is focused
type FocusPanel int

const (
	FocusMatrix FocusPanel = iota
	FocusLeaderboard
	FocusProcessMap
	FocusTimeline
)

// KeyMap defines keybindings
type KeyMap struct {
	Quit       key.Binding
	Pause      key.Binding
	Reset      key.Binding
	Toggle     key.Binding
	Help       key.Binding
	SpeedTest  key.Binding
	Theme      key.Binding
	Fullscreen key.Binding
	Details    key.Binding
	ConnGraph  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		SpeedTest: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "speed"),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
		Fullscreen: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fullscreen"),
		),
		Details: key.NewBinding(
			key.WithKeys("enter", "d"),
			key.WithHelp("d", "details"),
		),
		ConnGraph: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "graph"),
		),
	}
}

// ShortHelp returns keybindings for the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Pause, k.Toggle, k.Theme, k.Fullscreen, k.SpeedTest, k.Details, k.ConnGraph, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Pause, k.Reset, k.Toggle},
		{k.Theme, k.Fullscreen, k.SpeedTest, k.Details, k.ConnGraph, k.Quit},
	}
}

// tickMsg triggers animation updates
type tickMsg time.Time

// statsMsg delivers updated statistics
type statsMsg struct{}

// Model is the main application model
type Model struct {
	// Dimensions
	width  int
	height int

	// Panels
	matrix      *MatrixPanel
	leaderboard *LeaderboardPanel
	processMap  *ProcessMapPanel
	timeline    *TimelinePanel
	netInfo     *NetInfoBar
	speedTest   *SpeedTestPopup
	domainInfo  *DomainInfoPopup
	connGraph   *ConnectionGraphPanel

	// Animation
	animBorder *AnimatedBorder
	tick       int

	// State
	focus         FocusPanel
	fullscreen    bool
	paused        bool
	miniMode      bool
	showConnGraph bool
	keymap        KeyMap
	help          help.Model

	// Data
	capture     *capture.Capture
	processMap_ *data.ProcessMapper
	stats       data.StatsProvider
	netMonitor  *data.NetworkMonitor
	lastStats   []*data.DomainStats
	selectedIdx int

	// Debug counters
	totalPackets int64
	totalDomains int

	// Traffic rate tracking
	lastTotalBytes   int64
	lastPacketCount  int64
	bytesPerSec      int64 // Download
	uploadBytesPerSec int64 // Upload (simulated as ~30% of download for demo)
}

// NewModel creates a new application model
func NewModel(cap *capture.Capture, pm *data.ProcessMapper, stats data.StatsProvider) Model {
	iface := "en0"
	if cap != nil {
		iface = capture.DefaultInterface()
	}

	netMon := data.NewNetworkMonitor(iface)
	netMon.Start()

	return Model{
		matrix:      NewMatrixPanel(80, 20),
		leaderboard: NewLeaderboardPanel(60, 15),
		processMap:  NewProcessMapPanel(60, 15),
		timeline:    NewTimelinePanel(120, 10),
		netInfo:     NewNetInfoBar(120),
		speedTest:   NewSpeedTestPopup(),
		domainInfo:  NewDomainInfoPopup(),
		connGraph:   NewConnectionGraphPanel(80, 20),
		animBorder:  NewAnimatedBorder(),
		focus:       FocusMatrix,
		keymap:      DefaultKeyMap(),
		help:        help.New(),
		capture:     cap,
		processMap_: pm,
		stats:       stats,
		netMonitor:  netMon,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		statsCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func statsCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return statsMsg{}
	})
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle escape for popups
		if msg.String() == "esc" {
			if m.speedTest.IsVisible() {
				m.speedTest.Hide()
				return m, nil
			}
			if m.domainInfo.IsVisible() {
				m.domainInfo.Hide()
				return m, nil
			}
			if m.showConnGraph {
				m.showConnGraph = false
				return m, nil
			}
			if m.fullscreen {
				m.fullscreen = false
				return m, nil
			}
		}

		// Arrow keys for selection in leaderboard
		if m.focus == FocusLeaderboard && !m.speedTest.IsVisible() && !m.domainInfo.IsVisible() {
			switch msg.String() {
			case "up", "k":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil
			case "down", "j":
				if m.selectedIdx < len(m.lastStats)-1 {
					m.selectedIdx++
				}
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keymap.Quit):
			if m.netMonitor != nil {
				m.netMonitor.Stop()
			}
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Pause):
			m.paused = !m.paused

		case key.Matches(msg, m.keymap.Reset):
			m.selectedIdx = 0

		case key.Matches(msg, m.keymap.Toggle):
			m.focus = (m.focus + 1) % 4
			m.updateFocus()

		case key.Matches(msg, m.keymap.SpeedTest):
			if m.speedTest.IsVisible() {
				m.speedTest.Restart()
			} else {
				m.speedTest.Show()
			}

		case key.Matches(msg, m.keymap.Theme):
			CycleTheme()
			m.matrix.chars = CurrentTheme.MatrixChars

		case key.Matches(msg, m.keymap.Fullscreen):
			m.fullscreen = !m.fullscreen

		case key.Matches(msg, m.keymap.Details):
			if m.focus == FocusLeaderboard && len(m.lastStats) > m.selectedIdx {
				m.domainInfo.Show(m.lastStats[m.selectedIdx])
			}

		case key.Matches(msg, m.keymap.ConnGraph):
			m.showConnGraph = !m.showConnGraph
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.miniMode = msg.Height < 20
		m.resizePanels()

	case tickMsg:
		m.tick++
		m.animBorder.Update()
		if !m.paused {
			m.matrix.Update()
		}
		return m, tickCmd()

	case statsMsg:
		m.updateStats()
		if m.netMonitor != nil {
			m.netInfo.UpdateInfo(m.netMonitor.GetInfo())
		}
		return m, statsCmd()
	}

	return m, nil
}

func (m *Model) updateFocus() {
	m.animBorder.SetFocused(true)
	m.leaderboard.SetFocused(m.focus == FocusLeaderboard)
	m.processMap.SetFocused(m.focus == FocusProcessMap)
	m.timeline.SetFocused(m.focus == FocusTimeline)
}

func (m *Model) resizePanels() {
	// Matrix-only mode - full screen
	if matrixOnlyMode {
		m.matrix.Resize(m.width, m.height-1)
		return
	}

	// Layout: header(1) + topRow + timeline + help(1) + buffer(1)
	headerHeight := 1
	helpHeight := 1
	bufferHeight := 1
	timelineHeight := maxInt(5, (m.height-3)/6)
	topHeight := m.height - headerHeight - helpHeight - timelineHeight - bufferHeight

	matrixWidth := m.width * 55 / 100
	rightWidth := m.width - matrixWidth
	leaderboardHeight := (topHeight + 1) / 2
	processMapHeight := topHeight - leaderboardHeight

	// Content dimensions (panel height - 2 for borders)
	m.matrix.Resize(matrixWidth-4, topHeight-2)
	m.leaderboard.Resize(rightWidth-4, leaderboardHeight-2)
	m.processMap.Resize(rightWidth-4, processMapHeight-2)
	m.timeline.Resize(m.width-4, timelineHeight-2)
	m.netInfo.Resize(m.width - 4)
	m.connGraph.Resize(m.width-4, m.height-6)
}

func (m *Model) updateStats() {
	if m.stats == nil {
		return
	}

	stats := m.stats.GetStats()
	m.lastStats = stats
	m.totalDomains = len(stats)

	// Count total packets and bytes
	var totalPkts int64
	var totalBytes int64
	for _, s := range stats {
		totalPkts += s.Packets
		totalBytes += s.Bytes
	}
	m.totalPackets = totalPkts

	// Calculate bytes per second (download)
	m.bytesPerSec = totalBytes - m.lastTotalBytes
	if m.bytesPerSec < 0 {
		m.bytesPerSec = 0
	}
	m.lastTotalBytes = totalBytes

	// Simulate upload as ~20-35% of download (realistic ratio for browsing)
	// In real capture mode, this would come from actual outgoing packets
	uploadRatio := 0.2 + float64(m.tick%15)*0.01 // Varies between 20-35%
	m.uploadBytesPerSec = int64(float64(m.bytesPerSec) * uploadRatio)

	// Update matrix with traffic rate
	m.matrix.UpdateTrafficRate(m.bytesPerSec, totalPkts)

	// Update domain traffic for matrix
	domains := make(map[string]int64)
	for _, s := range stats {
		domains[s.Domain] = s.Bytes
	}
	m.matrix.UpdateDomains(domains)

	// Update other panels
	m.leaderboard.UpdateStats(stats)
	m.leaderboard.SetSelected(m.selectedIdx)
	m.processMap.UpdateStats(stats)
	m.timeline.UpdateStats(stats)
	m.connGraph.UpdateStats(stats)

	// Update network monitor with traffic delta (not cumulative)
	if m.netMonitor != nil {
		m.netMonitor.SetRate(m.bytesPerSec, totalPkts-m.lastPacketCount)
		m.lastPacketCount = totalPkts
	}
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	theme := CurrentTheme

	// Handle popups first (full screen modals)
	if m.speedTest.IsVisible() {
		return m.speedTest.View(m.width, m.height)
	}

	if m.domainInfo.IsVisible() {
		return m.domainInfo.View(m.width, m.height, theme)
	}

	// Connection graph view
	if m.showConnGraph {
		return m.renderConnGraphView(theme)
	}

	// Matrix-only mode
	if matrixOnlyMode {
		return m.renderMatrixOnly(theme)
	}

	// Mini mode
	if m.miniMode {
		return m.renderMiniMode(theme)
	}

	// Fullscreen mode
	if m.fullscreen {
		return m.renderFullscreen(theme)
	}

	// Normal mode
	return m.renderNormal(theme)
}

func (m Model) renderMatrixOnly(theme Theme) string {
	// Minimal mode - just the matrix, no status bar
	if matrixOpts.Minimal {
		m.matrix.Resize(m.width, m.height)
		return m.matrix.View()
	}

	// No info mode - matrix only
	if !matrixOpts.ShowInfo {
		m.matrix.Resize(m.width, m.height)
		return m.matrix.View()
	}

	// Status bar at top
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	primaryStyle := lipgloss.NewStyle().Foreground(theme.Primary)
	accentStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)

	sep := dimStyle.Render(" │ ")

	// Build left part based on options
	var leftParts []string
	leftParts = append(leftParts, primaryStyle.Render("◆ BYTEFALL"))

	if matrixOpts.ShowDownload {
		leftParts = append(leftParts, accentStyle.Render("↓ "+formatRate(m.bytesPerSec)))
	}

	if matrixOpts.ShowUpload {
		uploadStyle := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)
		leftParts = append(leftParts, uploadStyle.Render("↑ "+formatRate(m.uploadBytesPerSec)))
	}

	if matrixOpts.ShowDomains {
		leftParts = append(leftParts, dimStyle.Render(fmt.Sprintf("%d domains", m.totalDomains)))
	}

	if matrixOpts.ShowIP && m.netMonitor != nil {
		info := m.netMonitor.GetInfo()
		if info.IPv4 != "" {
			leftParts = append(leftParts, dimStyle.Render(info.Interface+":")+primaryStyle.Render(info.IPv4))
		}
	}

	if matrixOpts.ShowPublicIP && m.netMonitor != nil {
		info := m.netMonitor.GetInfo()
		if info.PublicIP != "" {
			leftParts = append(leftParts, dimStyle.Render("pub:")+accentStyle.Render(info.PublicIP))
		}
	}

	leftPart := strings.Join(leftParts, sep)

	// Build right part
	var rightParts []string
	rightParts = append(rightParts, dimStyle.Render("[")+primaryStyle.Render(theme.Name)+dimStyle.Render("]"))

	if m.paused {
		rightParts = append(rightParts, lipgloss.NewStyle().Foreground(theme.Warning).Render("⏸ PAUSED"))
	} else {
		rightParts = append(rightParts, primaryStyle.Render("▶ LIVE"))
	}

	rightPart := strings.Join(rightParts, sep)

	// Calculate padding
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)
	padding := m.width - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	statusBar := leftPart + strings.Repeat(" ", padding) + rightPart

	// Full screen matrix (height - 1 for status bar)
	m.matrix.Resize(m.width, m.height-1)
	content := m.matrix.View()

	return statusBar + "\n" + content
}

func (m Model) renderMiniMode(theme Theme) string {
	// Single line compact display
	style := lipgloss.NewStyle().Foreground(theme.Primary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	var parts []string
	parts = append(parts, style.Render("◆ BYTEFALL"))
	parts = append(parts, dimStyle.Render(" │ "))

	if m.netMonitor != nil {
		info := m.netMonitor.GetInfo()
		parts = append(parts, style.Render(fmt.Sprintf("▼%s", formatRate(info.BytesPerSec))))
		parts = append(parts, dimStyle.Render(" │ "))
	}

	if len(m.lastStats) > 0 {
		top := m.lastStats[0]
		domain := top.Domain
		if len(domain) > 20 {
			domain = domain[:17] + "..."
		}
		parts = append(parts, style.Render(domain))
		parts = append(parts, dimStyle.Render(fmt.Sprintf(" %s", formatBytes(top.Bytes))))
	}

	if m.paused {
		parts = append(parts, dimStyle.Render(" │ "))
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.Warning).Render("PAUSED"))
	}

	return strings.Join(parts, "")
}

func (m Model) renderFullscreen(theme Theme) string {
	// Fullscreen focused panel
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.BorderFocus)

	var content string
	var title string

	switch m.focus {
	case FocusMatrix:
		m.matrix.Resize(m.width-4, m.height-4)
		content = m.matrix.View()
		title = "MATRIX VIEW"
	case FocusLeaderboard:
		m.leaderboard.Resize(m.width-4, m.height-6)
		content = m.leaderboard.View()
		title = "LEADERBOARD"
	case FocusProcessMap:
		m.processMap.Resize(m.width-4, m.height-6)
		content = m.processMap.View()
		title = "PROCESS MAP"
	case FocusTimeline:
		m.timeline.Resize(m.width-4, m.height-6)
		content = m.timeline.View()
		title = "TIMELINE"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Dim).
		Padding(0, 2)

	header := titleStyle.Render(" " + title + " ") +
		lipgloss.NewStyle().Foreground(theme.Dim).Render(" [ESC] exit fullscreen  [TAB] switch panel")

	panel := borderStyle.
		Width(m.width - 2).
		Height(m.height - 3).
		Render(truncateHeight(content, m.height-5))

	return lipgloss.JoinVertical(lipgloss.Left, header, panel)
}

func (m Model) renderNormal(theme Theme) string {
	// Styles with current theme
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border)

	// Animated border for focused panel
	pulseColors := []lipgloss.Color{theme.BorderFocus, theme.Accent, theme.Primary, theme.Accent}
	animatedBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(pulseColors[(m.tick/3)%len(pulseColors)])

	// Header styles
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	primaryStyle := lipgloss.NewStyle().Foreground(theme.Primary)
	accentStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)

	sep := dimStyle.Render(" │ ")

	// Left side: Logo + Download/Upload Speed + Interface
	logo := primaryStyle.Render("◆ BYTEFALL")
	downloadSpeed := accentStyle.Render("↓ " + formatRate(m.bytesPerSec))
	uploadStyle := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)
	uploadSpeed := uploadStyle.Render("↑ " + formatRate(m.uploadBytesPerSec))

	var ifaceInfo string
	if m.netMonitor != nil {
		info := m.netMonitor.GetInfo()
		if info.Interface != "" {
			ifaceInfo = dimStyle.Render(info.Interface)
			if info.IPv4 != "" {
				ifaceInfo += dimStyle.Render(":") + primaryStyle.Render(info.IPv4)
			}
		}
	}

	leftPart := logo + sep + downloadSpeed + " " + uploadSpeed
	if ifaceInfo != "" {
		leftPart += sep + ifaceInfo
	}

	// Right side: Stats + Theme + Status
	statsInfo := dimStyle.Render(fmt.Sprintf("%d domains", m.totalDomains))
	themeText := dimStyle.Render("[") + primaryStyle.Render(theme.Name) + dimStyle.Render("]")

	var statusText string
	if m.paused {
		statusText = lipgloss.NewStyle().Foreground(theme.Warning).Render("⏸ PAUSED")
	} else {
		statusText = primaryStyle.Render("▶ LIVE")
	}

	rightPart := statsInfo + sep + themeText + sep + statusText

	// Calculate padding
	headerLeftWidth := lipgloss.Width(leftPart)
	headerRightWidth := lipgloss.Width(rightPart)
	headerPadding := m.width - headerLeftWidth - headerRightWidth
	if headerPadding < 0 {
		headerPadding = 0
	}

	headerLine := leftPart + strings.Repeat(" ", headerPadding) + rightPart

	// Calculate dimensions to fill terminal
	// Layout: header(1) + topRow + timeline + help(1) + buffer(1)
	headerHeight := 1
	helpHeight := 1
	bufferHeight := 1
	timelineHeight := maxInt(5, (m.height-3)/6)
	topHeight := m.height - headerHeight - helpHeight - timelineHeight - bufferHeight

	matrixWidth := m.width * 55 / 100
	rightWidth := m.width - matrixWidth
	// Split right column evenly, giving extra line to top panel if odd
	leaderboardHeight := (topHeight + 1) / 2
	processMapHeight := topHeight - leaderboardHeight

	// Render panels with themed borders
	getBorder := func(panel FocusPanel) lipgloss.Style {
		if m.focus == panel {
			return animatedBorderStyle
		}
		return borderStyle
	}

	// Content height = panel height - 2 (for top/bottom border)
	matrixContentHeight := topHeight - 2
	leaderboardContentHeight := leaderboardHeight - 2
	processMapContentHeight := processMapHeight - 2
	timelineContentHeight := timelineHeight - 2

	// Render panels - use only truncateHeight to control exact line count
	// Don't use lipgloss Height() as it can cause inconsistent rendering
	matrixContent := m.matrix.View()
	matrixPanel := getBorder(FocusMatrix).
		Width(matrixWidth - 2).
		Render(truncateHeight(matrixContent, matrixContentHeight))

	leaderboardContent := m.leaderboard.View()
	leaderboardPanel := getBorder(FocusLeaderboard).
		Width(rightWidth - 2).
		Render(truncateHeight(leaderboardContent, leaderboardContentHeight))

	processMapContent := m.processMap.View()
	processMapPanel := getBorder(FocusProcessMap).
		Width(rightWidth - 2).
		Render(truncateHeight(processMapContent, processMapContentHeight))

	timelineContent := m.timeline.View()
	timelinePanel := getBorder(FocusTimeline).
		Width(m.width - 2).
		Render(truncateHeight(timelineContent, timelineContentHeight))

	// Compose layout - ensure exact heights before joining
	matrixPanel = padToHeight(matrixPanel, topHeight)
	leaderboardPanel = padToHeight(leaderboardPanel, leaderboardHeight)
	processMapPanel = padToHeight(processMapPanel, processMapHeight)
	timelinePanel = padToHeight(timelinePanel, timelineHeight)

	rightColumn := lipgloss.JoinVertical(lipgloss.Left, leaderboardPanel, processMapPanel)
	topRow := joinHorizontalFixed(matrixPanel, rightColumn, topHeight)

	// Help with theme colors
	m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(theme.Primary)
	m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(theme.Dim)
	helpView := m.help.View(m.keymap)

	return lipgloss.JoinVertical(lipgloss.Left, headerLine, topRow, timelinePanel, helpView)
}

func (m Model) renderConnGraphView(theme Theme) string {
	// Full screen connection graph view
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.BorderFocus)

	// Resize connection graph to fill screen
	m.connGraph.Resize(m.width-4, m.height-6)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Background(theme.Dim).
		Padding(0, 2)

	header := titleStyle.Render(" CONNECTION GRAPH ") +
		lipgloss.NewStyle().Foreground(theme.Dim).Render(" [ESC/G] close  [T] theme")

	content := m.connGraph.View()
	panel := borderStyle.
		Width(m.width - 2).
		Height(m.height - 3).
		Render(truncateHeight(content, m.height-5))

	return lipgloss.JoinVertical(lipgloss.Left, header, panel)
}

// truncateHeight ensures content doesn't exceed maxLines
func truncateHeight(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		for len(lines) < maxLines {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:maxLines], "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// padToHeight ensures a rendered string has exactly the specified number of lines
func padToHeight(content string, height int) string {
	lines := strings.Split(content, "\n")

	// Truncate if too many lines
	if len(lines) > height {
		lines = lines[:height]
	}

	// Pad if too few lines
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// joinHorizontalFixed joins two panels side by side, line by line, ensuring exact height
func joinHorizontalFixed(left, right string, height int) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	// Pad to exact height
	for len(leftLines) < height {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < height {
		rightLines = append(rightLines, "")
	}

	// Get the visual width of the left panel
	leftWidth := 0
	for _, line := range leftLines {
		w := lipgloss.Width(line)
		if w > leftWidth {
			leftWidth = w
		}
	}

	// Join line by line
	var result []string
	for i := 0; i < height; i++ {
		leftLine := ""
		rightLine := ""
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		// Pad left line to consistent width
		leftVisualWidth := lipgloss.Width(leftLine)
		if leftVisualWidth < leftWidth {
			leftLine = leftLine + strings.Repeat(" ", leftWidth-leftVisualWidth)
		}

		result = append(result, leftLine+rightLine)
	}

	return strings.Join(result, "\n")
}
