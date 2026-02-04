package data

import (
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// NetworkInfo holds information about the network interface
type NetworkInfo struct {
	Interface    string
	IPv4         string
	IPv6         string
	MAC          string
	Gateway      string
	DNS          []string
	PublicIP     string
	BytesIn      int64
	BytesOut     int64
	PacketsIn    int64
	PacketsOut   int64
	BytesPerSec  int64
	PacketsPerSec int64
}

// NetworkMonitor tracks network statistics
type NetworkMonitor struct {
	info        NetworkInfo
	mu          sync.RWMutex
	iface       string
	lastBytes   int64
	lastPackets int64
	lastUpdate  time.Time
	stop        chan struct{}
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(iface string) *NetworkMonitor {
	return &NetworkMonitor{
		iface: iface,
		stop:  make(chan struct{}),
		info:  NetworkInfo{Interface: iface},
	}
}

// Start begins monitoring
func (nm *NetworkMonitor) Start() {
	nm.updateInfo()
	go nm.monitorLoop()
}

// Stop halts monitoring
func (nm *NetworkMonitor) Stop() {
	close(nm.stop)
}

// GetInfo returns current network info
func (nm *NetworkMonitor) GetInfo() NetworkInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.info
}

// AddTraffic records traffic stats
func (nm *NetworkMonitor) AddTraffic(bytesIn, bytesOut int64, packetsIn, packetsOut int64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.info.BytesIn += bytesIn
	nm.info.BytesOut += bytesOut
	nm.info.PacketsIn += packetsIn
	nm.info.PacketsOut += packetsOut
}

// SetRate directly sets the rate values (used when rate is calculated externally)
func (nm *NetworkMonitor) SetRate(bytesPerSec, packetsPerSec int64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.info.BytesPerSec = bytesPerSec
	nm.info.PacketsPerSec = packetsPerSec
}

func (nm *NetworkMonitor) monitorLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-nm.stop:
			return
		case <-ticker.C:
			nm.updateRates()
		}
	}
}

func (nm *NetworkMonitor) updateRates() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(nm.lastUpdate).Seconds()
	if elapsed > 0 {
		totalBytes := nm.info.BytesIn + nm.info.BytesOut
		totalPackets := nm.info.PacketsIn + nm.info.PacketsOut

		nm.info.BytesPerSec = int64(float64(totalBytes-nm.lastBytes) / elapsed)
		nm.info.PacketsPerSec = int64(float64(totalPackets-nm.lastPackets) / elapsed)

		nm.lastBytes = totalBytes
		nm.lastPackets = totalPackets
	}
	nm.lastUpdate = now
}

func (nm *NetworkMonitor) updateInfo() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Get interface info
	iface, err := net.InterfaceByName(nm.iface)
	if err == nil {
		nm.info.MAC = iface.HardwareAddr.String()

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					continue
				}
				if ip.To4() != nil {
					nm.info.IPv4 = ip.String()
				} else {
					nm.info.IPv6 = ip.String()
				}
			}
		}
	}

	// Get gateway (macOS specific)
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "gateway:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					nm.info.Gateway = parts[1]
				}
			}
		}
	}

	// Get DNS servers
	out, err = exec.Command("scutil", "--dns").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "nameserver[") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					nm.info.DNS = append(nm.info.DNS, parts[2])
				}
			}
		}
		// Deduplicate DNS
		seen := make(map[string]bool)
		unique := []string{}
		for _, dns := range nm.info.DNS {
			if !seen[dns] {
				seen[dns] = true
				unique = append(unique, dns)
			}
		}
		nm.info.DNS = unique
		if len(nm.info.DNS) > 3 {
			nm.info.DNS = nm.info.DNS[:3]
		}
	}

	// Get public IP (async to not block)
	go nm.fetchPublicIP()
}

func (nm *NetworkMonitor) fetchPublicIP() {
	// Try to get public IP from various services
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	for _, svc := range services {
		out, err := exec.Command("curl", "-s", "-m", "2", svc).Output()
		if err == nil {
			ip := strings.TrimSpace(string(out))
			if net.ParseIP(ip) != nil {
				nm.mu.Lock()
				nm.info.PublicIP = ip
				nm.mu.Unlock()
				return
			}
		}
	}
}
