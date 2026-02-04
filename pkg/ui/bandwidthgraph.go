package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// BandwidthHistory stores historical bandwidth data
type BandwidthHistory struct {
	mu          sync.RWMutex
	download    []int64 // bytes per second samples
	upload      []int64
	timestamps  []time.Time
	maxSamples  int
	peakDown    int64
	peakUp      int64
}

var bandwidthHistory = &BandwidthHistory{
	maxSamples: 120, // 2 minutes of history at 1 sample/sec
}

// RecordBandwidth adds a new bandwidth sample
func RecordBandwidth(download, upload int64) {
	bandwidthHistory.mu.Lock()
	defer bandwidthHistory.mu.Unlock()

	bandwidthHistory.download = append(bandwidthHistory.download, download)
	bandwidthHistory.upload = append(bandwidthHistory.upload, upload)
	bandwidthHistory.timestamps = append(bandwidthHistory.timestamps, time.Now())

	// Track peaks
	if download > bandwidthHistory.peakDown {
		bandwidthHistory.peakDown = download
	}
	if upload > bandwidthHistory.peakUp {
		bandwidthHistory.peakUp = upload
	}

	// Trim old samples
	if len(bandwidthHistory.download) > bandwidthHistory.maxSamples {
		bandwidthHistory.download = bandwidthHistory.download[1:]
		bandwidthHistory.upload = bandwidthHistory.upload[1:]
		bandwidthHistory.timestamps = bandwidthHistory.timestamps[1:]
	}
}

// Graph characters for different heights (8 levels)
var graphChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderBandwidthGraphWidget renders full-screen bandwidth history graph
func RenderBandwidthGraphWidget(width, height int, currentDown, currentUp int64, theme Theme) string {
	bandwidthHistory.mu.RLock()
	defer bandwidthHistory.mu.RUnlock()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	downloadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	uploadStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Text)

	var lines []string

	// Title (system-wide bandwidth)
	titleLine := dimStyle.Render("═══════") + "  " + titleStyle.Render("◆ SYSTEM BANDWIDTH ◆") + "  " + dimStyle.Render("═══════")
	lines = append(lines, centerText(titleLine, width))
	lines = append(lines, centerText(dimStyle.Render("All Network Traffic"), width))
	lines = append(lines, "")

	// Current speed display
	currentLine := downloadStyle.Render(fmt.Sprintf("↓ %s", formatRateLarge(currentDown))) +
		dimStyle.Render("  │  ") +
		uploadStyle.Render(fmt.Sprintf("↑ %s", formatRateLarge(currentUp)))
	lines = append(lines, centerText(currentLine, width))
	lines = append(lines, "")

	// Calculate graph dimensions
	graphWidth := width - 16 // Leave room for Y-axis labels
	if graphWidth < 30 {
		graphWidth = 30
	}
	if graphWidth > 120 {
		graphWidth = 120
	}
	graphHeight := (height - 16) / 2 // Split between download and upload
	if graphHeight < 5 {
		graphHeight = 5
	}
	if graphHeight > 15 {
		graphHeight = 15
	}

	// Find max value for scaling
	maxVal := bandwidthHistory.peakDown
	if bandwidthHistory.peakUp > maxVal {
		maxVal = bandwidthHistory.peakUp
	}
	if maxVal < 1024 {
		maxVal = 1024 // Minimum 1 KB/s scale
	}

	// Download graph
	lines = append(lines, centerText(downloadStyle.Render("▼ DOWNLOAD"), width))
	downloadGraph := renderGraph(bandwidthHistory.download, graphWidth, graphHeight, maxVal, theme, true)
	for _, line := range downloadGraph {
		lines = append(lines, centerText(line, width))
	}
	lines = append(lines, "")

	// Separator with time axis
	timeAxis := renderTimeAxis(graphWidth, len(bandwidthHistory.download), dimStyle)
	lines = append(lines, centerText(timeAxis, width))
	lines = append(lines, "")

	// Upload graph
	lines = append(lines, centerText(uploadStyle.Render("▲ UPLOAD"), width))
	uploadGraph := renderGraph(bandwidthHistory.upload, graphWidth, graphHeight, maxVal, theme, false)
	for _, line := range uploadGraph {
		lines = append(lines, centerText(line, width))
	}
	lines = append(lines, "")

	// Stats footer
	lines = append(lines, centerText(dimStyle.Render(strings.Repeat("─", graphWidth+10)), width))

	peakLine := labelStyle.Render("Peak: ") +
		downloadStyle.Render(fmt.Sprintf("↓ %s", formatRateLarge(bandwidthHistory.peakDown))) +
		dimStyle.Render("  •  ") +
		uploadStyle.Render(fmt.Sprintf("↑ %s", formatRateLarge(bandwidthHistory.peakUp)))
	lines = append(lines, centerText(peakLine, width))

	// Calculate averages
	var avgDown, avgUp int64
	if len(bandwidthHistory.download) > 0 {
		for _, v := range bandwidthHistory.download {
			avgDown += v
		}
		avgDown /= int64(len(bandwidthHistory.download))
	}
	if len(bandwidthHistory.upload) > 0 {
		for _, v := range bandwidthHistory.upload {
			avgUp += v
		}
		avgUp /= int64(len(bandwidthHistory.upload))
	}

	avgLine := labelStyle.Render("Avg:  ") +
		downloadStyle.Render(fmt.Sprintf("↓ %s", formatRateLarge(avgDown))) +
		dimStyle.Render("  •  ") +
		uploadStyle.Render(fmt.Sprintf("↑ %s", formatRateLarge(avgUp)))
	lines = append(lines, centerText(avgLine, width))

	return verticalCenter(lines, width, height)
}

func renderGraph(data []int64, width, height int, maxVal int64, theme Theme, isDownload bool) []string {
	if len(data) == 0 {
		// Empty graph
		lines := make([]string, height)
		dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)
		for i := 0; i < height; i++ {
			lines[i] = dimStyle.Render("│") + strings.Repeat(" ", width)
		}
		return lines
	}

	// Resample data to fit width
	samples := resampleData(data, width)

	// Create the graph lines (top to bottom)
	lines := make([]string, height)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dim)

	for row := 0; row < height; row++ {
		var lineBuilder strings.Builder

		// Y-axis label on first row
		if row == 0 {
			lineBuilder.WriteString(dimStyle.Render(fmt.Sprintf("%6s │", formatRateShort(maxVal))))
		} else if row == height-1 {
			lineBuilder.WriteString(dimStyle.Render("     0 │"))
		} else {
			lineBuilder.WriteString(dimStyle.Render("       │"))
		}

		// Calculate threshold for this row (top row = maxVal, bottom row = 0)
		threshold := float64(height-1-row) / float64(height-1)

		for _, val := range samples {
			normalizedVal := float64(val) / float64(maxVal)
			if normalizedVal > 1.0 {
				normalizedVal = 1.0
			}

			// Determine character based on how much of this cell is filled
			cellFill := (normalizedVal - threshold) * float64(height)
			if cellFill < 0 {
				cellFill = 0
			}
			if cellFill > 1 {
				cellFill = 1
			}

			charIdx := int(cellFill * float64(len(graphChars)-1))
			if charIdx < 0 {
				charIdx = 0
			}
			if charIdx >= len(graphChars) {
				charIdx = len(graphChars) - 1
			}

			var color lipgloss.Color
			if isDownload {
				// Gradient based on value
				if normalizedVal > 0.8 {
					color = theme.Rain5
				} else if normalizedVal > 0.6 {
					color = theme.Rain4
				} else if normalizedVal > 0.4 {
					color = theme.Rain3
				} else if normalizedVal > 0.2 {
					color = theme.Rain2
				} else {
					color = theme.Rain1
				}
			} else {
				color = theme.Secondary
			}

			style := lipgloss.NewStyle().Foreground(color)

			if cellFill > 0 {
				lineBuilder.WriteString(style.Render(string(graphChars[charIdx])))
			} else {
				lineBuilder.WriteString(" ")
			}
		}

		lines[row] = lineBuilder.String()
	}

	return lines
}

func resampleData(data []int64, targetWidth int) []int64 {
	if len(data) == 0 {
		return make([]int64, targetWidth)
	}

	if len(data) <= targetWidth {
		// Pad with zeros at the beginning
		result := make([]int64, targetWidth)
		offset := targetWidth - len(data)
		copy(result[offset:], data)
		return result
	}

	// Downsample by averaging
	result := make([]int64, targetWidth)
	ratio := float64(len(data)) / float64(targetWidth)

	for i := 0; i < targetWidth; i++ {
		start := int(float64(i) * ratio)
		end := int(float64(i+1) * ratio)
		if end > len(data) {
			end = len(data)
		}
		if start >= end {
			start = end - 1
		}
		if start < 0 {
			start = 0
		}

		var sum int64
		count := 0
		for j := start; j < end; j++ {
			sum += data[j]
			count++
		}
		if count > 0 {
			result[i] = sum / int64(count)
		}
	}

	return result
}

func renderTimeAxis(width, dataLen int, dimStyle lipgloss.Style) string {
	var axis strings.Builder
	axis.WriteString(dimStyle.Render("       └"))

	// Create time markers
	for i := 0; i < width; i++ {
		if i == 0 {
			axis.WriteString(dimStyle.Render("─"))
		} else if i == width-1 {
			axis.WriteString(dimStyle.Render("┘"))
		} else if i == width/2 {
			axis.WriteString(dimStyle.Render("┴"))
		} else {
			axis.WriteString(dimStyle.Render("─"))
		}
	}

	return axis.String()
}

func formatRateShort(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%dB", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.0fK", float64(bytesPerSec)/1024)
	} else {
		return fmt.Sprintf("%.1fM", float64(bytesPerSec)/(1024*1024))
	}
}
