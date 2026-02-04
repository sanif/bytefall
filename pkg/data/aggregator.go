package data

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sanif/bytefall/pkg/capture"
)

// DomainStats holds aggregated statistics for a domain
type DomainStats struct {
	Domain     string
	Bytes      int64
	Packets    int64
	LastSeen   time.Time
	Processes  map[string]bool
	History    []int64 // Rolling 60-second history of bytes per second
	historyIdx int
}

// connKey creates a unique key for a connection
func connKey(srcIP string, srcPort uint16, dstIP string, dstPort uint16) string {
	return fmt.Sprintf("%s:%d->%s:%d", srcIP, srcPort, dstIP, dstPort)
}

// Aggregator collects and aggregates packet statistics
type Aggregator struct {
	domains     map[string]*DomainStats
	mu          sync.RWMutex

	// Connection to domain cache - tracks which domain each connection is for
	connCache   map[string]string // connKey -> domain
	connMu      sync.RWMutex

	// DNS cache - learned from SNI and reverse DNS
	dnsCache    *DNSCache

	// Reverse DNS cache (for failed lookups)
	reverseDNSCache    map[string]string
	reverseDNSMu       sync.RWMutex

	processMap  *ProcessMapper
	stop        chan struct{}
	historySize int
	Updates     chan struct{}
}

// NewAggregator creates a new Aggregator
func NewAggregator(pm *ProcessMapper) *Aggregator {
	return &Aggregator{
		domains:         make(map[string]*DomainStats),
		connCache:       make(map[string]string),
		dnsCache:        NewDNSCache(),
		reverseDNSCache: make(map[string]string),
		processMap:      pm,
		stop:            make(chan struct{}),
		historySize:     60,
		Updates:         make(chan struct{}, 1),
	}
}

// Start begins processing packets
func (a *Aggregator) Start(ctx context.Context, packets <-chan capture.Packet) {
	go a.processLoop(ctx, packets)
	go a.historyLoop(ctx)
	go a.cleanupLoop(ctx)
}

// Stop halts the aggregator
func (a *Aggregator) Stop() {
	close(a.stop)
}

// GetStats returns a snapshot of all domain statistics
func (a *Aggregator) GetStats() []*DomainStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := make([]*DomainStats, 0, len(a.domains))
	for _, s := range a.domains {
		copied := &DomainStats{
			Domain:    s.Domain,
			Bytes:     s.Bytes,
			Packets:   s.Packets,
			LastSeen:  s.LastSeen,
			Processes: make(map[string]bool),
			History:   make([]int64, len(s.History)),
		}
		for p := range s.Processes {
			copied.Processes[p] = true
		}
		copy(copied.History, s.History)
		stats = append(stats, copied)
	}
	return stats
}

// GetDomainCount returns the number of tracked domains
func (a *Aggregator) GetDomainCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.domains)
}

func (a *Aggregator) processLoop(ctx context.Context, packets <-chan capture.Packet) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stop:
			return
		case pkt, ok := <-packets:
			if !ok {
				return
			}
			a.processPacket(pkt)
		}
	}
}

func (a *Aggregator) processPacket(pkt capture.Packet) {
	// Create connection keys (both directions)
	outKey := connKey(pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
	inKey := connKey(pkt.DstIP, pkt.DstPort, pkt.SrcIP, pkt.SrcPort)

	var domain string

	// Check if we already know the domain for this connection
	a.connMu.RLock()
	if d, ok := a.connCache[outKey]; ok {
		domain = d
	} else if d, ok := a.connCache[inKey]; ok {
		domain = d
	}
	a.connMu.RUnlock()

	// If not cached, try to extract from this packet
	if domain == "" {
		// Try SNI extraction from TLS ClientHello
		if len(pkt.Payload) > 0 {
			domain = ExtractSNI(pkt.Payload)
			if domain != "" {
				// Cache this SNI for future packets to this IP
				a.dnsCache.AddFromSNI(pkt.DstIP, domain)
			}
		}

		// Try our learned DNS cache
		if domain == "" {
			domain, _ = a.dnsCache.Lookup(pkt.DstIP)
		}

		// Try well-known IP ranges (Google, Cloudflare, etc)
		if domain == "" {
			domain = LookupWellKnown(pkt.DstIP)
		}

		// Try reverse DNS (async, may not have result immediately)
		if domain == "" {
			domain = a.reverseDNS(pkt.DstIP)
		}

		// Fallback to IP
		if domain == "" {
			domain = pkt.DstIP
		}

		// Cache for this connection
		a.connMu.Lock()
		a.connCache[outKey] = domain
		a.connCache[inKey] = domain
		a.connMu.Unlock()
	}

	// Get process info
	var process string
	if a.processMap != nil {
		proc, _, found := a.processMap.Lookup(pkt.SrcIP, pkt.SrcPort, pkt.DstIP, pkt.DstPort)
		if found {
			process = proc
		}
	}

	// Update stats
	a.mu.Lock()
	stats, ok := a.domains[domain]
	if !ok {
		stats = &DomainStats{
			Domain:    domain,
			Processes: make(map[string]bool),
			History:   make([]int64, a.historySize),
		}
		a.domains[domain] = stats
	}
	stats.Bytes += int64(pkt.Length)
	stats.Packets++
	stats.LastSeen = time.Now()
	if process != "" {
		stats.Processes[process] = true
	}
	a.mu.Unlock()

	// Signal update
	select {
	case a.Updates <- struct{}{}:
	default:
	}
}

func (a *Aggregator) reverseDNS(ip string) string {
	a.reverseDNSMu.RLock()
	if domain, ok := a.reverseDNSCache[ip]; ok {
		a.reverseDNSMu.RUnlock()
		return domain
	}
	a.reverseDNSMu.RUnlock()

	// Perform reverse DNS lookup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var domain string
	done := make(chan struct{})

	go func() {
		names, err := net.LookupAddr(ip)
		if err == nil && len(names) > 0 {
			domain = names[0]
			// Remove trailing dot
			if len(domain) > 0 && domain[len(domain)-1] == '.' {
				domain = domain[:len(domain)-1]
			}
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		// Timeout - cache empty result
		domain = ""
	case <-done:
		// Got result
	}

	a.reverseDNSMu.Lock()
	a.reverseDNSCache[ip] = domain
	a.reverseDNSMu.Unlock()

	// Also add to main DNS cache if we got a result
	if domain != "" {
		a.dnsCache.Add(ip, domain)
	}

	return domain
}

func (a *Aggregator) historyLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	lastBytes := make(map[string]int64)

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stop:
			return
		case <-ticker.C:
			a.mu.Lock()
			for domain, stats := range a.domains {
				// Calculate bytes in the last second
				bytesNow := stats.Bytes
				bytesDelta := bytesNow - lastBytes[domain]
				lastBytes[domain] = bytesNow

				// Shift history and add new value
				copy(stats.History[1:], stats.History[:len(stats.History)-1])
				stats.History[0] = bytesDelta
			}
			a.mu.Unlock()
		}
	}
}

// cleanupLoop removes old connection cache entries
func (a *Aggregator) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stop:
			return
		case <-ticker.C:
			// Keep cache from growing too large
			a.connMu.Lock()
			if len(a.connCache) > 10000 {
				// Clear half the cache
				count := 0
				for k := range a.connCache {
					delete(a.connCache, k)
					count++
					if count > 5000 {
						break
					}
				}
			}
			a.connMu.Unlock()
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
