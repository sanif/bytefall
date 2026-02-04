package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// MatrixColumn represents a single falling column of characters
type MatrixColumn struct {
	chars    []rune
	position int
	speed    int
	domain   string
	active   bool
	intensity int // 0-4, affects brightness
	fade     []int
}

// MatrixPanel displays the Matrix-style waterfall animation
type MatrixPanel struct {
	width   int
	height  int
	columns []MatrixColumn
	domains map[string]int64 // Domain to bytes mapping
	chars   []rune
	rng     *rand.Rand

	// Traffic tracking for animation intensity
	lastBytes     int64
	bytesPerSec   int64
	packetsPerSec int64
	peakBytes     int64
}

// Matrix characters for the waterfall effect
// Greek + Cyrillic + symbols for unique hacker aesthetic
var matrixChars = []rune("ΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩαβγδεζηθικλμνξοπρστυφχψω0123456789∆∇∂∑∏√∞≈≠±×÷")

// NewMatrixPanel creates a new Matrix animation panel
func NewMatrixPanel(width, height int) *MatrixPanel {
	m := &MatrixPanel{
		width:     width,
		height:    height,
		columns:   make([]MatrixColumn, width),
		domains:   make(map[string]int64),
		chars:     matrixChars,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		peakBytes: 1000, // Initial baseline
	}

	m.initColumns()
	return m
}

func (m *MatrixPanel) initColumns() {
	for i := range m.columns {
		m.columns[i] = MatrixColumn{
			chars:     make([]rune, m.height),
			fade:      make([]int, m.height),
			position:  -m.rng.Intn(m.height * 2),
			speed:     1,
			intensity: 1,
			active:    false,
		}
		for j := range m.columns[i].chars {
			m.columns[i].chars[j] = m.chars[m.rng.Intn(len(m.chars))]
		}
	}
}

// Update advances the animation state based on traffic
func (m *MatrixPanel) Update() {
	// Calculate traffic intensity (0.0 to 1.0)
	var totalBytes int64
	for _, bytes := range m.domains {
		totalBytes += bytes
	}

	// Update peak tracking
	bytesDelta := totalBytes - m.lastBytes
	if bytesDelta < 0 {
		bytesDelta = 0
	}
	m.bytesPerSec = bytesDelta * 20 // Since we update ~20 times/sec (50ms tick)
	m.lastBytes = totalBytes

	if m.bytesPerSec > m.peakBytes {
		m.peakBytes = m.bytesPerSec
	}
	// Decay peak slowly
	m.peakBytes = m.peakBytes * 99 / 100
	if m.peakBytes < 1000 {
		m.peakBytes = 1000
	}

	// Calculate intensity ratio (0.0 to 1.0)
	intensity := float64(m.bytesPerSec) / float64(m.peakBytes)
	if intensity > 1.0 {
		intensity = 1.0
	}

	// Determine how many columns should be active based on traffic
	minActive := len(m.columns) / 10        // At least 10% always active
	maxActive := len(m.columns) * 8 / 10    // Up to 80% when busy
	targetActive := minActive + int(float64(maxActive-minActive)*intensity)

	// Count current active columns
	activeCount := 0
	for _, col := range m.columns {
		if col.active && col.position < m.height+10 {
			activeCount++
		}
	}

	// Activate more columns if needed
	for activeCount < targetActive {
		// Find an inactive column
		idx := m.rng.Intn(len(m.columns))
		if !m.columns[idx].active || m.columns[idx].position > m.height {
			m.activateColumn(idx, intensity)
			activeCount++
		}
	}

	// Update each column
	for i := range m.columns {
		col := &m.columns[i]

		if !col.active {
			continue
		}

		// Speed based on intensity (1-4)
		baseSpeed := 1 + int(intensity*3)
		col.speed = baseSpeed
		if m.rng.Float64() < intensity*0.3 {
			col.speed++ // Random burst
		}

		// Advance position
		col.position += col.speed

		// Reset if past bottom
		if col.position > m.height+15 {
			if m.rng.Float64() < intensity*0.8 || m.rng.Float64() < 0.1 {
				m.activateColumn(i, intensity)
			} else {
				col.active = false
			}
		}

		// Update fade levels
		for j := 0; j < m.height; j++ {
			dist := col.position - j
			if dist < 0 {
				col.fade[j] = 0
			} else if dist < 2 {
				col.fade[j] = 4 // Brightest - the head
			} else if dist < 4 {
				col.fade[j] = 3
			} else if dist < 8 {
				col.fade[j] = 2
			} else if dist < 14 {
				col.fade[j] = 1
			} else {
				col.fade[j] = 0
			}

			// Boost intensity based on traffic
			if col.intensity > 2 && col.fade[j] > 0 {
				col.fade[j] = min(4, col.fade[j]+1)
			}
		}

		// Occasionally mutate characters (more often with more traffic)
		mutateChance := 5 + int(intensity*15)
		if m.rng.Intn(100) < mutateChance && col.position > 0 && col.position < m.height {
			idx := col.position
			if idx >= 0 && idx < m.height {
				col.chars[idx] = m.chars[m.rng.Intn(len(m.chars))]
			}
		}
	}
}

func (m *MatrixPanel) activateColumn(idx int, intensity float64) {
	col := &m.columns[idx]
	col.active = true
	col.position = -m.rng.Intn(m.height / 2)
	col.speed = 1 + int(intensity*3)
	col.intensity = 1 + int(intensity*3)
	col.domain = m.pickWeightedDomain()

	// Refresh characters
	for j := range col.chars {
		col.chars[j] = m.chars[m.rng.Intn(len(m.chars))]
	}
}

// UpdateDomains sets the domain traffic data
func (m *MatrixPanel) UpdateDomains(domains map[string]int64) {
	m.domains = domains
}

// UpdateTrafficRate sets the current traffic rate for animation intensity
func (m *MatrixPanel) UpdateTrafficRate(bytesPerSec, packetsPerSec int64) {
	m.bytesPerSec = bytesPerSec
	m.packetsPerSec = packetsPerSec
}

func (m *MatrixPanel) pickWeightedDomain() string {
	if len(m.domains) == 0 {
		return ""
	}

	var total int64
	for _, bytes := range m.domains {
		total += bytes
	}

	if total == 0 {
		return ""
	}

	pick := m.rng.Int63n(total)
	var cumulative int64
	for domain, bytes := range m.domains {
		cumulative += bytes
		if pick < cumulative {
			return domain
		}
	}

	return ""
}

// Resize updates the panel dimensions
func (m *MatrixPanel) Resize(width, height int) {
	if width == m.width && height == m.height {
		return
	}

	m.width = width
	m.height = height
	m.columns = make([]MatrixColumn, width)
	m.initColumns()
}

// View renders the Matrix panel
func (m *MatrixPanel) View() string {
	theme := CurrentTheme

	// Color palette for fade levels - dedicated rain colors for smooth gradient
	colors := []lipgloss.Style{
		lipgloss.NewStyle().Foreground(theme.Rain1), // Darkest
		lipgloss.NewStyle().Foreground(theme.Rain2),
		lipgloss.NewStyle().Foreground(theme.Rain3),
		lipgloss.NewStyle().Foreground(theme.Rain4),
		lipgloss.NewStyle().Foreground(theme.Rain5).Bold(true), // Brightest (head)
	}

	var lines []string

	for row := 0; row < m.height; row++ {
		var line strings.Builder
		colPos := 0 // Track actual display column position

		for col := 0; colPos < m.width && col < len(m.columns); col++ {
			c := m.columns[col]
			if row < len(c.fade) && row < len(c.chars) {
				fade := c.fade[row]
				char := c.chars[row]
				charWidth := runewidth.RuneWidth(char)

				// Skip if this character would overflow the width
				if colPos+charWidth > m.width {
					// Fill remaining space
					for colPos < m.width {
						line.WriteRune(' ')
						colPos++
					}
					break
				}

				if fade > 0 && fade <= 4 {
					line.WriteString(colors[fade].Render(string(char)))
				} else {
					// Write spaces for the character width
					for i := 0; i < charWidth; i++ {
						line.WriteRune(' ')
					}
				}
				colPos += charWidth
			} else {
				line.WriteRune(' ')
				colPos++
			}
		}

		// Pad to exact width
		for colPos < m.width {
			line.WriteRune(' ')
			colPos++
		}

		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

