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
- **Network Speed Widget** — Large fullscreen speed display with ↓/↑ gauges
- **Widget Modes** — Run any panel as a clean fullscreen widget
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

### Homebrew (Recommended)

```bash
brew tap sanif/tap
brew install bytefall
```

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
  -version          Show version
  -help             Show help message

Widget Modes (fullscreen, clean by default):
  -matrix           Fullscreen matrix rain widget
  -speed            Fullscreen network speed widget with gauges
  -leaderboard      Fullscreen domain rankings widget
  -processes        Fullscreen process map widget
  -timeline         Fullscreen activity timeline widget
  -graph            Fullscreen connection graph widget

Status Bar Options (for widget modes):
  -bar              Show status bar in widget mode (off by default)
  -down             Show download speed (default: true)
  -up               Show upload speed (default: true)
  -domains          Show active domain count (default: true)
  -ip               Show local IP address
  -public-ip        Show public IP address
```

### Examples

```bash
# List available interfaces
bytefall -list

# Capture on specific interface
sudo bytefall -i en0

# Demo mode (no sudo required)
bytefall -demo

# Cyberpunk theme
sudo bytefall -theme cyberpunk

# Widget modes (clean, fullscreen)
bytefall -demo -matrix           # Matrix rain
bytefall -demo -speed            # Network speed with ↓/↑ gauges
bytefall -demo -leaderboard      # Domain rankings
bytefall -demo -timeline         # Activity sparklines
bytefall -demo -graph            # Connection graph

# Widget with status bar
bytefall -demo -matrix -bar
bytefall -demo -speed -bar -ip -public-ip
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

## Widget Modes

ByteFall can run any panel as a clean, fullscreen widget — perfect for desktop dashboards or ambient displays.

| Flag | Widget | Description |
|------|--------|-------------|
| `-matrix` | Matrix Rain | Animated falling characters |
| `-speed` | Network Speed | Large ↓/↑ speed display with gauges |
| `-leaderboard` | Domain Rankings | Full-screen sorted domain list |
| `-processes` | Process Map | Process-to-domain tree view |
| `-timeline` | Activity Timeline | Sparklines for all domains |
| `-graph` | Connection Graph | Process cards with connections |

By default, widget modes show only the content (clean mode). Add `-bar` to include a status bar with network info.

```bash
# Clean matrix rain (no status bar)
bytefall -demo -matrix

# Speed widget with status bar showing IP
bytefall -demo -speed -bar -ip

# Leaderboard with full network info
bytefall -demo -leaderboard -bar -ip -public-ip
```

### Speed Widget Preview

```
                    ↓ 12.5 MB/s
         ████████████████░░░░░░░░░░░░░░

                    ↑ 3.2 MB/s
         █████░░░░░░░░░░░░░░░░░░░░░░░░░

                 ◈ 1,234 packets/s
```

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

## Shell Completion

ByteFall supports shell completion for bash, zsh, and fish.

```bash
# Bash (add to ~/.bashrc)
eval "$(bytefall -completion bash)"

# Zsh (add to ~/.zshrc)
eval "$(bytefall -completion zsh)"

# Fish (add to ~/.config/fish/config.fish)
bytefall -completion fish | source
```

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
