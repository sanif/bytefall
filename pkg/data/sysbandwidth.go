package data

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SystemBandwidth tracks actual system-wide network bandwidth
type SystemBandwidth struct {
	mu sync.RWMutex

	// Current rates (bytes per second)
	DownloadRate int64
	UploadRate   int64

	// Totals
	TotalDownload int64
	TotalUpload   int64

	// Peaks (since start)
	PeakDownload int64
	PeakUpload   int64

	// Interface being monitored
	Interface string

	// Internal tracking
	lastBytesIn  int64
	lastBytesOut int64
	lastUpdate   time.Time

	stop chan struct{}
}

// NewSystemBandwidth creates a new system bandwidth monitor
func NewSystemBandwidth(iface string) *SystemBandwidth {
	if iface == "" {
		iface = "en0" // Default macOS interface
	}
	return &SystemBandwidth{
		Interface: iface,
		stop:      make(chan struct{}),
	}
}

// Start begins monitoring system bandwidth
func (sb *SystemBandwidth) Start() {
	// Get initial values
	sb.updateStats()
	sb.lastUpdate = time.Now()

	go sb.monitorLoop()
}

// Stop halts monitoring
func (sb *SystemBandwidth) Stop() {
	close(sb.stop)
}

// GetRates returns current download and upload rates in bytes/sec
func (sb *SystemBandwidth) GetRates() (download, upload int64) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.DownloadRate, sb.UploadRate
}

// GetTotals returns total bytes transferred
func (sb *SystemBandwidth) GetTotals() (download, upload int64) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.TotalDownload, sb.TotalUpload
}

// GetPeaks returns peak rates since start
func (sb *SystemBandwidth) GetPeaks() (download, upload int64) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.PeakDownload, sb.PeakUpload
}

func (sb *SystemBandwidth) monitorLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sb.stop:
			return
		case <-ticker.C:
			sb.updateStats()
		}
	}
}

func (sb *SystemBandwidth) updateStats() {
	bytesIn, bytesOut := sb.readInterfaceStats()

	sb.mu.Lock()
	defer sb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(sb.lastUpdate).Seconds()

	if elapsed > 0 && sb.lastBytesIn > 0 {
		// Calculate rates
		deltaIn := bytesIn - sb.lastBytesIn
		deltaOut := bytesOut - sb.lastBytesOut

		// Handle counter wrap-around (rare but possible)
		if deltaIn < 0 {
			deltaIn = bytesIn
		}
		if deltaOut < 0 {
			deltaOut = bytesOut
		}

		sb.DownloadRate = int64(float64(deltaIn) / elapsed)
		sb.UploadRate = int64(float64(deltaOut) / elapsed)

		// Update peaks
		if sb.DownloadRate > sb.PeakDownload {
			sb.PeakDownload = sb.DownloadRate
		}
		if sb.UploadRate > sb.PeakUpload {
			sb.PeakUpload = sb.UploadRate
		}
	}

	sb.lastBytesIn = bytesIn
	sb.lastBytesOut = bytesOut
	sb.TotalDownload = bytesIn
	sb.TotalUpload = bytesOut
	sb.lastUpdate = now
}

// readInterfaceStats reads bytes in/out from netstat on macOS
func (sb *SystemBandwidth) readInterfaceStats() (bytesIn, bytesOut int64) {
	// Use netstat -ib to get interface statistics
	// Format: Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll
	out, err := exec.Command("netstat", "-ib").Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		// Match interface name (first column)
		if fields[0] != sb.Interface {
			continue
		}

		// Skip header row and link-level entries without IP
		// We want the row with actual IP address
		if fields[2] == "Link#" || fields[2] == "<Link#" || strings.HasPrefix(fields[2], "Link") {
			// This is the link-level row, use it for byte counts
			// Columns: Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
			//          0    1   2       3       4     5     6      7     8     9      10
			ibytes, _ := strconv.ParseInt(fields[6], 10, 64)
			obytes, _ := strconv.ParseInt(fields[9], 10, 64)
			return ibytes, obytes
		}
	}

	// Fallback: try to find any row for this interface
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 10 && fields[0] == sb.Interface {
			// Try parsing byte columns (positions vary by row type)
			for i := 4; i < len(fields)-3; i++ {
				if ibytes, err := strconv.ParseInt(fields[i], 10, 64); err == nil {
					if obytes, err := strconv.ParseInt(fields[i+3], 10, 64); err == nil {
						if ibytes > 0 || obytes > 0 {
							return ibytes, obytes
						}
					}
				}
			}
		}
	}

	return 0, 0
}

// GetAllInterfaceStats returns bandwidth for all active interfaces
func GetAllInterfaceStats() map[string][2]int64 {
	result := make(map[string][2]int64)

	out, err := exec.Command("netstat", "-ib").Output()
	if err != nil {
		return result
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		iface := fields[0]
		if iface == "Name" || iface == "" {
			continue
		}

		// Only process link-level rows
		if strings.HasPrefix(fields[2], "Link") || fields[2] == "<Link#" {
			ibytes, _ := strconv.ParseInt(fields[6], 10, 64)
			obytes, _ := strconv.ParseInt(fields[9], 10, 64)
			if ibytes > 0 || obytes > 0 {
				result[iface] = [2]int64{ibytes, obytes}
			}
		}
	}

	return result
}

// DemoSystemBandwidth simulates realistic system bandwidth for demo mode
type DemoSystemBandwidth struct {
	mu sync.RWMutex

	DownloadRate int64
	UploadRate   int64
	PeakDownload int64
	PeakUpload   int64

	tick int
	stop chan struct{}
}

// NewDemoSystemBandwidth creates a simulated system bandwidth
func NewDemoSystemBandwidth() *DemoSystemBandwidth {
	return &DemoSystemBandwidth{
		stop: make(chan struct{}),
	}
}

// Start begins simulating bandwidth
func (dsb *DemoSystemBandwidth) Start() {
	go dsb.simulateLoop()
}

// Stop halts simulation
func (dsb *DemoSystemBandwidth) Stop() {
	close(dsb.stop)
}

// GetRates returns simulated download and upload rates
func (dsb *DemoSystemBandwidth) GetRates() (download, upload int64) {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()
	return dsb.DownloadRate, dsb.UploadRate
}

// GetPeaks returns peak rates
func (dsb *DemoSystemBandwidth) GetPeaks() (download, upload int64) {
	dsb.mu.RLock()
	defer dsb.mu.RUnlock()
	return dsb.PeakDownload, dsb.PeakUpload
}

func (dsb *DemoSystemBandwidth) simulateLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dsb.stop:
			return
		case <-ticker.C:
			dsb.generateTraffic()
		}
	}
}

func (dsb *DemoSystemBandwidth) generateTraffic() {
	dsb.mu.Lock()
	defer dsb.mu.Unlock()

	dsb.tick++

	// Simulate realistic traffic patterns:
	// Base: 500KB-2MB download, 100KB-500KB upload
	// With occasional bursts and wave patterns

	// Base rates
	baseDown := int64(500*1024 + (dsb.tick*1234)%(1500*1024))
	baseUp := int64(100*1024 + (dsb.tick*567)%(400*1024))

	// Add wave pattern (simulates periodic activity)
	wave := 1.0 + 0.3*float64((dsb.tick%60)-30)/30.0 // -0.3 to +0.3

	// Occasional bursts (simulates file downloads, etc.)
	burst := 1.0
	if dsb.tick%17 == 0 { // Every ~17 seconds, simulate a burst
		burst = 3.0 + float64(dsb.tick%5)
	}
	if dsb.tick%43 == 0 { // Less frequent large burst
		burst = 5.0 + float64(dsb.tick%10)
	}

	// Calculate final rates
	dsb.DownloadRate = int64(float64(baseDown) * wave * burst)
	dsb.UploadRate = int64(float64(baseUp) * wave * (burst * 0.3)) // Upload bursts less

	// Keep rates positive and reasonable
	if dsb.DownloadRate < 10*1024 {
		dsb.DownloadRate = 10*1024 + int64(dsb.tick%50)*1024
	}
	if dsb.UploadRate < 5*1024 {
		dsb.UploadRate = 5*1024 + int64(dsb.tick%20)*1024
	}

	// Cap at reasonable maxes (100 MB/s down, 50 MB/s up)
	if dsb.DownloadRate > 100*1024*1024 {
		dsb.DownloadRate = 100*1024*1024 - int64(dsb.tick%1000)*1024
	}
	if dsb.UploadRate > 50*1024*1024 {
		dsb.UploadRate = 50*1024*1024 - int64(dsb.tick%500)*1024
	}

	// Track peaks
	if dsb.DownloadRate > dsb.PeakDownload {
		dsb.PeakDownload = dsb.DownloadRate
	}
	if dsb.UploadRate > dsb.PeakUpload {
		dsb.PeakUpload = dsb.UploadRate
	}
}
