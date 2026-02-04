package data

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// DemoAggregator generates fake traffic data for demo mode
type DemoAggregator struct {
	domains     map[string]*DomainStats
	mu          sync.RWMutex
	stop        chan struct{}
	historySize int
	rng         *rand.Rand
	tick        int
}

// Sample domains for demo
var demoDomains = []struct {
	domain  string
	process string
	weight  int
	burst   bool // Can have traffic bursts
}{
	// High traffic
	{"github.com", "chrome", 95, true},
	{"api.github.com", "git", 85, true},
	{"google.com", "chrome", 90, false},
	{"www.googleapis.com", "chrome", 80, true},
	{"api.anthropic.com", "claude", 88, true},

	// Medium traffic
	{"cdn.jsdelivr.net", "node", 65, false},
	{"registry.npmjs.org", "npm", 55, true},
	{"slack.com", "Slack", 50, false},
	{"api.slack.com", "Slack", 45, true},
	{"s3.amazonaws.com", "aws-cli", 60, true},

	// Lower traffic
	{"twitter.com", "chrome", 35, false},
	{"cloudflare.com", "curl", 30, false},
	{"docker.io", "docker", 25, true},
	{"pkg.go.dev", "go", 20, false},
	{"crates.io", "cargo", 18, false},
	{"pypi.org", "pip", 22, true},

	// Occasional
	{"netflix.com", "chrome", 15, true},
	{"spotify.com", "Spotify", 12, false},
	{"discord.com", "Discord", 28, true},
	{"zoom.us", "zoom", 10, true},
}

// NewDemoAggregator creates a demo aggregator with fake data
func NewDemoAggregator() *DemoAggregator {
	return &DemoAggregator{
		domains:     make(map[string]*DomainStats),
		stop:        make(chan struct{}),
		historySize: 60,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Start begins generating demo data
func (d *DemoAggregator) Start() {
	// Initialize with some base data
	d.mu.Lock()
	for _, dd := range demoDomains {
		baseBytes := int64(d.rng.Intn(500000) + 50000)
		d.domains[dd.domain] = &DomainStats{
			Domain:    dd.domain,
			Bytes:     baseBytes,
			Packets:   baseBytes / 1000,
			LastSeen:  time.Now(),
			Processes: map[string]bool{dd.process: true},
			History:   make([]int64, d.historySize),
		}
		// Initialize history with wave pattern
		for i := range d.domains[dd.domain].History {
			wave := math.Sin(float64(i) * 0.2)
			d.domains[dd.domain].History[i] = int64(float64(d.rng.Intn(5000)+2000) * (1 + wave*0.5))
		}
	}
	d.mu.Unlock()

	go d.generateLoop()
}

// Stop halts demo data generation
func (d *DemoAggregator) Stop() {
	close(d.stop)
}

// GetStats returns a snapshot of all domain statistics
func (d *DemoAggregator) GetStats() []*DomainStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make([]*DomainStats, 0, len(d.domains))
	for _, s := range d.domains {
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

func (d *DemoAggregator) generateLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	historyTicker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer historyTicker.Stop()

	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			d.simulateTraffic()
		case <-historyTicker.C:
			d.rotateHistory()
		}
	}
}

func (d *DemoAggregator) simulateTraffic() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.tick++

	// Randomly add traffic to domains based on weight
	for _, dd := range demoDomains {
		if d.rng.Intn(100) < dd.weight {
			if stats, ok := d.domains[dd.domain]; ok {
				// Base traffic
				bytes := int64(d.rng.Intn(5000) + 500)

				// Occasional bursts for burst-capable domains
				if dd.burst && d.rng.Intn(100) < 5 {
					bytes *= int64(d.rng.Intn(10) + 5)
				}

				// Add some wave pattern based on tick
				wave := math.Sin(float64(d.tick) * 0.05)
				bytes = int64(float64(bytes) * (1 + wave*0.3))

				stats.Bytes += bytes
				stats.Packets += int64(d.rng.Intn(10) + 1)
				stats.LastSeen = time.Now()
			}
		}
	}
}

func (d *DemoAggregator) rotateHistory() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for domain, stats := range d.domains {
		// Shift history
		copy(stats.History[1:], stats.History[:len(stats.History)-1])

		// Find domain config for weight
		var weight int = 50
		for _, dd := range demoDomains {
			if dd.domain == domain {
				weight = dd.weight
				break
			}
		}

		// Generate current rate with some variation
		base := int64(weight * 100)
		wave := math.Sin(float64(d.tick) * 0.1)
		variation := int64(d.rng.Intn(3000))
		stats.History[0] = base + variation + int64(float64(base)*wave*0.5)

		if stats.History[0] < 0 {
			stats.History[0] = 0
		}
	}
}
