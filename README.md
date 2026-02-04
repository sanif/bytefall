# ByteFall

**Real-time network traffic visualization with Matrix-style aesthetics**

ByteFall answers the question: *"Where is my data actually going right now?"* — displaying your machine's outbound network traffic as animated Matrix-rain streams, grouped by destination domain and source process.

![macOS](https://img.shields.io/badge/platform-macOS-blue)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

---

## Features

- **Matrix Rain Visualization** — Animated falling character streams colored by domain, with intensity reflecting traffic volume
- **Domain Leaderboard** — Real-time ranking of destinations by bytes/second
- **Process Mapping** — See which applications are connecting to which domains
- **Activity Timeline** — 60-second sparkline history per domain
- **Connection Graph** — Animated visualization of active connections
- **SNI Extraction** — Identifies domains from TLS ClientHello handshakes
- **Process Detection** — Maps network connections to their source processes
- **5 Color Themes** — Matrix, Cyberpunk, Amber, Ocean, Blood
- **Demo Mode** — Explore the interface without root privileges
- **Single Binary** — No runtime dependencies, just one executable

---

## Screenshots

<!-- Add screenshots or GIF demo here -->
```
┌─ Matrix ─────────────────┐┌─ Leaderboard ────────────┐
│ ╪ ╫ ║ ╬ ╪ ║ ╫ ╪ ╬ ║ ╪ ╫ ││ github.com      1.2 MB/s │
│ ╫ ╬ ╪ ║ ╫ ╬ ╪ ║ ╬ ╪ ╫ ║ ││ cloudflare.com  890 KB/s │
│ ║ ╪ ╫ ╬ ║ ╪ ╫ ╬ ║ ╫ ╬ ╪ ││ api.openai.com  256 KB/s │
│ ╬ ╫ ║ ╪ ╬ ╫ ║ ╪ ╫ ╬ ║ ╪ ││ ...                      │
└──────────────────────────┘└──────────────────────────┘
┌─ Process Map ────────────┐┌─ Timeline ───────────────┐
│ Chrome    → github.com   ││ github.com     ▁▃▅▇█▇▅▃▁ │
│ Slack     → slack.com    ││ cloudflare.com ▂▄▆▇▆▄▂▁▁ │
│ Terminal  → api.openai   ││ api.openai.com ▁▁▂▃▅▇█▇▅ │
└──────────────────────────┘└──────────────────────────┘
```

---

## Requirements

- **macOS** (v1 supports macOS only)
- **Root privileges** for packet capture (or use demo mode)
- Terminal with 24-bit color support recommended

---

## Installation

### Using Go

```bash
go install github.com/sanif/bytefall/cmd/bytefall@latest
```

### From Source

```bash
git clone https://github.com/sanif/bytefall.git
cd bytefall
go build -o bytefall ./cmd/bytefall
```

### Pre-built Binary

Download the latest release from the [Releases](https://github.com/sanif/bytefall/releases) page.

---

## Usage

### Basic Usage (requires sudo)

```bash
sudo ./bytefall
```

### Demo Mode (no privileges required)

```bash
./bytefall -demo
```

### Command-Line Options

```
bytefall [options]

Options:
  -i <interface>    Network interface to capture from (auto-detects default)
  -list             List available network interfaces
  -demo             Run in demo mode with simulated traffic
  -theme <name>     Color theme: matrix, cyberpunk, amber, ocean, blood
  -matrix           Show only the matrix animation panel
  -minimal          Hide status bar completely

Status Bar Options:
  -down             Show download speed
  -up               Show upload speed
  -domains          Show active domain count
  -ip               Show local IP address
  -public-ip        Show public IP address
```

### Examples

```bash
# List available interfaces
./bytefall -list

# Capture on specific interface
sudo ./bytefall -i en0

# Cyberpunk theme with matrix-only view
sudo ./bytefall -theme cyberpunk -matrix

# Demo mode with ocean theme
./bytefall -demo -theme ocean
```

---

## Key Bindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `p` | Pause/Resume capture |
| `r` | Reset statistics |
| `Tab` | Focus next panel |
| `Shift+Tab` | Focus previous panel |
| `f` | Toggle fullscreen for focused panel |
| `t` | Cycle through themes |
| `s` | Run speed test |
| `d` | Show domain details |
| `g` | Toggle connection graph |
| `?` | Show help |

---

## Themes

ByteFall includes 5 built-in color themes:

| Theme | Description |
|-------|-------------|
| **Matrix** | Classic green-on-black |
| **Cyberpunk** | Neon pink and cyan |
| **Amber** | Retro terminal amber |
| **Ocean** | Cool blues and teals |
| **Blood** | Deep reds |

Switch themes with the `-theme` flag or press `t` during runtime.

---

## Architecture

```
bytefall/
├── cmd/bytefall/       # CLI entry point
└── pkg/
    ├── capture/        # libpcap packet sniffing
    ├── data/           # Domain resolution, process mapping, aggregation
    └── ui/             # BubbleTea TUI components
```

### Data Flow

```
Packets → Capture → SNI/DNS Resolution → Process Mapping → Aggregation → TUI
```

**Key Components:**

- **Capture**: Uses libpcap via gopacket to sniff TCP traffic on ports 80/443
- **SNI Extraction**: Parses TLS ClientHello to identify destination domains
- **Process Mapping**: Correlates connections to processes using `lsof`
- **Aggregator**: Maintains real-time stats with 60-second rolling history
- **TUI**: BubbleTea-based interface with 4 synchronized panels

---

## Limitations

- **macOS only** (v1) — Linux/Windows support planned for future versions
- **Ports 80/443 only** — HTTP/HTTPS traffic
- **Domain-level visibility** — Full URL paths are encrypted (no MITM)
- **Real-time only** — No historical data persistence
- **Outbound traffic only** — Inbound connections not tracked

---

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Styling and layout
- [gopacket](https://github.com/google/gopacket) — Packet capture and parsing

---

## Building

```bash
# Build
go build -o bytefall ./cmd/bytefall

# Run tests
go test ./...

# Build for release
go build -ldflags="-s -w" -o bytefall ./cmd/bytefall
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Charm](https://charm.sh/) for the excellent TUI libraries
- The Matrix (1999) for the visual inspiration
