package data

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Connection represents a network connection with process info
type Connection struct {
	LocalIP    string
	LocalPort  uint16
	RemoteIP   string
	RemotePort uint16
	PID        int
	Process    string
}

// ProcessMapper maps network connections to processes using lsof
type ProcessMapper struct {
	cache     map[string]Connection
	cacheMu   sync.RWMutex
	stop      chan struct{}
	pollRate  time.Duration
}

// NewProcessMapper creates a new ProcessMapper
func NewProcessMapper(pollRate time.Duration) *ProcessMapper {
	return &ProcessMapper{
		cache:    make(map[string]Connection),
		stop:     make(chan struct{}),
		pollRate: pollRate,
	}
}

// Start begins periodic lsof polling
func (pm *ProcessMapper) Start() {
	go pm.pollLoop()
}

// Stop halts the polling loop
func (pm *ProcessMapper) Stop() {
	close(pm.stop)
}

// Lookup returns the process for a given connection
func (pm *ProcessMapper) Lookup(localIP string, localPort uint16, remoteIP string, remotePort uint16) (string, int, bool) {
	pm.cacheMu.RLock()
	defer pm.cacheMu.RUnlock()

	// Try exact match first
	key := fmt.Sprintf("%s:%d-%s:%d", localIP, localPort, remoteIP, remotePort)
	if conn, ok := pm.cache[key]; ok {
		return conn.Process, conn.PID, true
	}

	// Try with wildcard local IP
	key = fmt.Sprintf("*:%d-%s:%d", localPort, remoteIP, remotePort)
	if conn, ok := pm.cache[key]; ok {
		return conn.Process, conn.PID, true
	}

	// Try just local port match
	for _, conn := range pm.cache {
		if conn.LocalPort == localPort && conn.RemotePort == remotePort {
			return conn.Process, conn.PID, true
		}
	}

	return "", 0, false
}

func (pm *ProcessMapper) pollLoop() {
	ticker := time.NewTicker(pm.pollRate)
	defer ticker.Stop()

	// Initial poll
	pm.poll()

	for {
		select {
		case <-pm.stop:
			return
		case <-ticker.C:
			pm.poll()
		}
	}
}

var lsofRegex = regexp.MustCompile(`^(\S+)\s+(\d+).*TCP\s+(\S+)->(\S+)`)

func (pm *ProcessMapper) poll() {
	cmd := exec.Command("lsof", "-i", "TCP", "-n", "-P")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	newCache := make(map[string]Connection)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		conn := pm.parseLsofLine(line)
		if conn != nil {
			key := fmt.Sprintf("%s:%d-%s:%d", conn.LocalIP, conn.LocalPort, conn.RemoteIP, conn.RemotePort)
			newCache[key] = *conn
		}
	}

	pm.cacheMu.Lock()
	pm.cache = newCache
	pm.cacheMu.Unlock()
}

func (pm *ProcessMapper) parseLsofLine(line string) *Connection {
	matches := lsofRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	processName := matches[1]
	pid, _ := strconv.Atoi(matches[2])
	local := matches[3]
	remote := matches[4]

	localIP, localPort := parseAddress(local)
	remoteIP, remotePort := parseAddress(remote)

	if localPort == 0 || remotePort == 0 {
		return nil
	}

	return &Connection{
		LocalIP:    localIP,
		LocalPort:  localPort,
		RemoteIP:   remoteIP,
		RemotePort: remotePort,
		PID:        pid,
		Process:    processName,
	}
}

func parseAddress(addr string) (string, uint16) {
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return addr, 0
	}
	ip := addr[:idx]
	port, _ := strconv.Atoi(addr[idx+1:])
	return ip, uint16(port)
}
