package capture

import (
	"fmt"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packet represents a captured network packet with metadata
type Packet struct {
	SrcIP      string
	DstIP      string
	SrcPort    uint16
	DstPort    uint16
	Length     int
	Protocol   string
	Payload    []byte
	Timestamp  int64
}

// Capture manages packet capture from a network interface
type Capture struct {
	handle     *pcap.Handle
	iface      string
	packets    chan Packet
	stop       chan struct{}
	mu         sync.Mutex
	running    bool
}

// New creates a new Capture instance
func New(iface string) *Capture {
	return &Capture{
		iface:   iface,
		packets: make(chan Packet, 1000),
		stop:    make(chan struct{}),
	}
}

// Start begins packet capture
func (c *Capture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("capture already running")
	}

	// Open the device
	handle, err := pcap.OpenLive(c.iface, 65535, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("failed to open interface %s: %w", c.iface, err)
	}

	// Set BPF filter for TCP ports 80 and 443
	filter := "tcp port 80 or tcp port 443"
	if err := handle.SetBPFFilter(filter); err != nil {
		handle.Close()
		return fmt.Errorf("failed to set BPF filter: %w", err)
	}

	c.handle = handle
	c.running = true
	c.stop = make(chan struct{})

	go c.captureLoop()

	return nil
}

// Stop halts packet capture
func (c *Capture) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	close(c.stop)
	if c.handle != nil {
		c.handle.Close()
	}
	c.running = false
}

// Packets returns the channel of captured packets
func (c *Capture) Packets() <-chan Packet {
	return c.packets
}

func (c *Capture) captureLoop() {
	packetSource := gopacket.NewPacketSource(c.handle, c.handle.LinkType())
	packets := packetSource.Packets()

	for {
		select {
		case <-c.stop:
			return
		case pkt, ok := <-packets:
			if !ok {
				return
			}
			if p := c.parsePacket(pkt); p != nil {
				select {
				case c.packets <- *p:
				default:
					// Channel full, drop packet
				}
			}
		}
	}
}

func (c *Capture) parsePacket(pkt gopacket.Packet) *Packet {
	// Get IP layer
	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return nil
	}
	ip, _ := ipLayer.(*layers.IPv4)

	// Get TCP layer
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil
	}
	tcp, _ := tcpLayer.(*layers.TCP)

	p := &Packet{
		SrcIP:     ip.SrcIP.String(),
		DstIP:     ip.DstIP.String(),
		SrcPort:   uint16(tcp.SrcPort),
		DstPort:   uint16(tcp.DstPort),
		Length:    len(pkt.Data()),
		Protocol:  "TCP",
		Timestamp: pkt.Metadata().Timestamp.UnixNano(),
	}

	// Get application payload
	appLayer := pkt.ApplicationLayer()
	if appLayer != nil {
		p.Payload = appLayer.Payload()
	}

	return p
}

// ListInterfaces returns available network interfaces
func ListInterfaces() ([]string, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	var interfaces []string
	for _, dev := range devices {
		interfaces = append(interfaces, dev.Name)
	}
	return interfaces, nil
}

// DefaultInterface returns the default capture interface
func DefaultInterface() string {
	return "en0" // macOS default
}
