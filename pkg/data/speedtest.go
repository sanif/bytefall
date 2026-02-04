package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// SpeedTestResult holds speed test results
type SpeedTestResult struct {
	DownloadSpeed float64 // Mbps
	UploadSpeed   float64 // Mbps
	Latency       time.Duration
	Server        string
	Status        string
	Progress      float64 // 0-100
	Phase         string  // "latency", "download", "upload", "complete"
}

// SpeedTest performs network speed testing
type SpeedTest struct {
	result   SpeedTestResult
	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
	onUpdate func(SpeedTestResult)
}

// Test servers (using Cloudflare's speed test endpoints)
var testServers = []struct {
	name string
	url  string
}{
	{"Cloudflare", "https://speed.cloudflare.com/__down?bytes=25000000"},
}

// NewSpeedTest creates a new speed test
func NewSpeedTest(onUpdate func(SpeedTestResult)) *SpeedTest {
	return &SpeedTest{
		onUpdate: onUpdate,
		result: SpeedTestResult{
			Status: "idle",
			Phase:  "idle",
		},
	}
}

// Start begins the speed test
func (st *SpeedTest) Start() {
	st.mu.Lock()
	if st.running {
		st.mu.Unlock()
		return
	}
	st.running = true
	st.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel

	go st.run(ctx)
}

// Stop cancels the speed test
func (st *SpeedTest) Stop() {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.cancel != nil {
		st.cancel()
	}
	st.running = false
}

// IsRunning returns whether test is in progress
func (st *SpeedTest) IsRunning() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.running
}

// GetResult returns current result
func (st *SpeedTest) GetResult() SpeedTestResult {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.result
}

func (st *SpeedTest) updateResult(r SpeedTestResult) {
	st.mu.Lock()
	st.result = r
	st.mu.Unlock()

	if st.onUpdate != nil {
		st.onUpdate(r)
	}
}

func (st *SpeedTest) run(ctx context.Context) {
	defer func() {
		st.mu.Lock()
		st.running = false
		st.mu.Unlock()
	}()

	server := testServers[0]

	// Phase 1: Latency test
	st.updateResult(SpeedTestResult{
		Status:   "Testing latency...",
		Phase:    "latency",
		Progress: 0,
		Server:   server.name,
	})

	latency := st.measureLatency(ctx, "https://speed.cloudflare.com")
	if ctx.Err() != nil {
		return
	}

	// Phase 2: Download test
	st.updateResult(SpeedTestResult{
		Status:   "Testing download speed...",
		Phase:    "download",
		Progress: 10,
		Latency:  latency,
		Server:   server.name,
	})

	downloadSpeed := st.measureDownload(ctx, server.url)
	if ctx.Err() != nil {
		return
	}

	// Phase 3: Upload test
	st.updateResult(SpeedTestResult{
		Status:        "Testing upload speed...",
		Phase:         "upload",
		Progress:      60,
		Latency:       latency,
		DownloadSpeed: downloadSpeed,
		Server:        server.name,
	})

	uploadSpeed := st.measureUpload(ctx)
	if ctx.Err() != nil {
		return
	}

	// Complete
	st.updateResult(SpeedTestResult{
		Status:        "Complete",
		Phase:         "complete",
		Progress:      100,
		Latency:       latency,
		DownloadSpeed: downloadSpeed,
		UploadSpeed:   uploadSpeed,
		Server:        server.name,
	})
}

func (st *SpeedTest) measureLatency(ctx context.Context, url string) time.Duration {
	client := &http.Client{Timeout: 5 * time.Second}

	var totalLatency time.Duration
	attempts := 3

	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return 0
		default:
		}

		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			totalLatency += time.Since(start)
		}
	}

	return totalLatency / time.Duration(attempts)
}

func (st *SpeedTest) measureDownload(ctx context.Context, url string) float64 {
	client := &http.Client{Timeout: 30 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var totalBytes int64
	buf := make([]byte, 32*1024)
	lastUpdate := start

	for {
		select {
		case <-ctx.Done():
			return 0
		default:
		}

		n, err := resp.Body.Read(buf)
		totalBytes += int64(n)

		// Update progress periodically
		if time.Since(lastUpdate) > 100*time.Millisecond {
			elapsed := time.Since(start).Seconds()
			if elapsed > 0 {
				speed := float64(totalBytes*8) / elapsed / 1_000_000 // Mbps
				progress := 10 + (float64(totalBytes) / 25_000_000 * 50)
				if progress > 60 {
					progress = 60
				}
				st.updateResult(SpeedTestResult{
					Status:        fmt.Sprintf("Download: %.1f Mbps", speed),
					Phase:         "download",
					Progress:      progress,
					DownloadSpeed: speed,
					Latency:       st.result.Latency,
					Server:        st.result.Server,
				})
			}
			lastUpdate = time.Now()
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}

	elapsed := time.Since(start).Seconds()
	if elapsed > 0 {
		return float64(totalBytes*8) / elapsed / 1_000_000 // Mbps
	}
	return 0
}

func (st *SpeedTest) measureUpload(ctx context.Context) float64 {
	// Generate test data
	testData := make([]byte, 10*1024*1024) // 10MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Use Cloudflare's upload endpoint
	url := "https://speed.cloudflare.com/__up"

	pr, pw := io.Pipe()

	var totalBytes int64
	start := time.Now()

	// Writer goroutine
	go func() {
		defer pw.Close()
		chunkSize := 64 * 1024
		for i := 0; i < len(testData); i += chunkSize {
			select {
			case <-ctx.Done():
				return
			default:
			}

			end := i + chunkSize
			if end > len(testData) {
				end = len(testData)
			}
			n, err := pw.Write(testData[i:end])
			if err != nil {
				return
			}
			totalBytes += int64(n)

			// Update progress
			elapsed := time.Since(start).Seconds()
			if elapsed > 0 {
				speed := float64(totalBytes*8) / elapsed / 1_000_000
				progress := 60 + (float64(totalBytes) / float64(len(testData)) * 40)
				st.updateResult(SpeedTestResult{
					Status:        fmt.Sprintf("Upload: %.1f Mbps", speed),
					Phase:         "upload",
					Progress:      progress,
					DownloadSpeed: st.result.DownloadSpeed,
					UploadSpeed:   speed,
					Latency:       st.result.Latency,
					Server:        st.result.Server,
				})
			}
		}
	}()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, pr)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Seconds()
	if elapsed > 0 {
		return float64(totalBytes*8) / elapsed / 1_000_000
	}
	return 0
}
