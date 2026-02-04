# ByteFall

**Real-time network traffic visualization with Matrix-style aesthetics**

ByteFall answers the question: *"Where is my data actually going right now?"* — displaying your machine's outbound network traffic as animated Matrix-rain streams, grouped by destination domain and source process.

![macOS](https://img.shields.io/badge/platform-macOS-blue)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

---

## Features

- **System-wide Bandwidth** — Real-time monitoring of all network traffic across all protocols and ports
- **Matrix Rain Visualization** — Animated falling character streams colored by domain, with intensity reflecting traffic volume
- **Domain Leaderboard** — Real-time ranking of HTTP/HTTPS destinations by bytes/second
- **Process Mapping** — See which applications are connecting to which domains
- **Activity Timeline** — 60-second sparkline history per domain
- **Connection Graph** — Animated visualization of active connections
- **Network Speed Widget** — Large fullscreen speed display with 5 visual styles (system-wide)
- **Bandwidth History** — Historical graphs of download/upload rates with peak/average stats
- **Top Applications** — Ranked list of apps by bandwidth usage with icons
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
  -apps             Fullscreen top applications by bandwidth
  -bandwidth        Fullscreen bandwidth history graph

Speed Widget Styles:
  -speed-style <s>  Style: minimal, boxed, retro, neon, compact (default: boxed)

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
bytefall -demo -speed            # Network speed gauges
bytefall -demo -leaderboard      # Domain rankings
bytefall -demo -apps             # Top applications
bytefall -demo -bandwidth        # Bandwidth history
bytefall -demo -timeline         # Activity sparklines
bytefall -demo -graph            # Connection graph

# Speed widget with different styles
bytefall -demo -speed -speed-style neon      # Cyberpunk neon style
bytefall -demo -speed -speed-style retro     # Retro terminal style
bytefall -demo -speed -speed-style minimal   # Clean minimal style
bytefall -demo -speed -speed-style compact   # Compact dense style

# Widget with status bar
bytefall -demo -matrix -bar
bytefall -demo -speed -bar -ip
bytefall -demo -apps -bar -public-ip
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
| `s` | Run speed test (or cycle styles in speed widget) |
| `d` | Show domain details |
| `g` | Toggle connection graph |
| `?` | Show help |

---

## Widget Modes

ByteFall can run any panel as a clean, fullscreen widget — perfect for desktop dashboards, ambient displays, or monitoring setups.

| Flag | Widget | Description |
|------|--------|-------------|
| `-matrix` | Matrix Rain | Animated falling characters colored by traffic |
| `-speed` | Network Speed | Large ↓/↑ speed display with gauges |
| `-leaderboard` | Domain Rankings | Full-screen sorted domain list |
| `-processes` | Process Map | Process-to-domain tree view |
| `-timeline` | Activity Timeline | Sparklines for all domains |
| `-graph` | Connection Graph | Process cards with connections |
| `-apps` | Top Applications | Ranked apps by bandwidth with icons |
| `-bandwidth` | Bandwidth Graph | Historical download/upload line graphs |

By default, widget modes show only the content (clean mode). Add `-bar` to include a status bar with network info.

```bash
# Widget examples
bytefall -demo -matrix           # Matrix rain
bytefall -demo -speed            # Network speed gauges
bytefall -demo -leaderboard      # Domain rankings
bytefall -demo -apps             # Top applications
bytefall -demo -bandwidth        # Bandwidth history
bytefall -demo -timeline         # Activity sparklines
bytefall -demo -graph            # Connection graph

# Add status bar with -bar
bytefall -demo -speed -bar -ip
bytefall -demo -apps -bar -public-ip
```

---

### Speed Widget

Large **system-wide** network speed display with animated gauges and 5 visual styles. Shows actual bandwidth across all protocols and ports (not just HTTP/HTTPS).

**Styles** (set via `-speed-style` or press `s` to cycle):

| Style | Description |
|-------|-------------|
| `minimal` | Clean arrows and numbers |
| `boxed` | Decorative frames and gauges (default) |
| `retro` | Old terminal green aesthetic |
| `neon` | Cyberpunk pink/cyan glow |
| `compact` | Dense single-line display |

**Auto-sizing**: Digits scale based on terminal size (large/medium/small).

```
            ╭─────  NETWORK SPEED  ─────╮

            ▾  DOWNLOAD

            ┌───┐ ╶───┐   ╷
            │   │     │  ██
            │   │ ┌───┘ ╵ ╵   MB/s
            │   │ │
            └───┘ └───╴

            ⟨▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░⟩ 65%

            ╶─────────────────────────────╴

            ▴  UPLOAD

                ╶───┐   ┌───┐
                    │   │   │
                 ───┤   └───┤   KB/s
                    │       │
                ╶───┘   ╶───┘

            ⟨▓▓▓▓▓▓░░░░░░░░░░░░░⟩ 22%

            ╰─────────────────────────────────────╯
```

---

### Top Applications Widget

Shows which apps are using your bandwidth, ranked by traffic volume with icons.

```
═══════  ◆ TOP APPLICATIONS ◆  ═══════

     APPLICATION        TRAFFIC   DOMAINS
─────────────────────────────────────────
 ★ 🌐 Chrome           45.2 MB       12  ████████████████░░░░
 2. 💬 Slack           12.1 MB        3  ████░░░░░░░░░░░░░░░░
 3. 💻 Code             8.4 MB        8  ███░░░░░░░░░░░░░░░░░
 4. 🐳 Docker           3.2 MB        2  █░░░░░░░░░░░░░░░░░░░

Total: 4 apps  •  68.9 MB transferred
```

---

### Bandwidth History Widget

Real-time graphs showing **system-wide** download and upload speeds over time with peak/average stats. Tracks all network traffic across all protocols and ports.

```
═══════  ◆ BANDWIDTH HISTORY ◆  ═══════

         ↓ 12.5 MB/s  │  ↑ 3.2 MB/s

            ▼ DOWNLOAD
 12.5M │        ▄█▆▄
       │      ▂▆████▅▃
       │    ▁▄████████▆▄▂
     0 └──────────────────┘

            ▲ UPLOAD
  3.2M │    ▃▅▄
       │  ▂▅███▅▃▂
       │ ▁███████▆▄▃▂▁
     0 └──────────────────┘

Peak: ↓ 15.8 MB/s  •  ↑ 4.1 MB/s
Avg:  ↓ 8.2 MB/s   •  ↑ 1.8 MB/s
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
System Interface → netstat → System Bandwidth (all traffic)
Packets → Capture → SNI/DNS → Process Mapping → Domain Stats (HTTP/HTTPS)
```

**Key Components:**

- **System Bandwidth**: Reads actual interface statistics via `netstat -ib` for true system-wide bandwidth
- **Capture**: Uses libpcap via gopacket to sniff TCP traffic on ports 80/443 for domain tracking
- **SNI Extraction**: Parses TLS ClientHello to identify destination domains
- **Process Mapping**: Correlates connections to processes using `lsof`
- **Aggregator**: Maintains real-time stats with 60-second rolling history
- **TUI**: BubbleTea-based interface with synchronized panels

---

## Limitations

- **macOS only** (v1) — Linux/Windows support planned for future versions
- **Domain tracking: Ports 80/443 only** — Domain identification uses HTTP/HTTPS traffic (speed widgets show all traffic)
- **Domain-level visibility** — Full URL paths are encrypted (no MITM)
- **Real-time only** — No historical data persistence
- **Outbound traffic only** — Inbound connections not tracked for domain stats

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
