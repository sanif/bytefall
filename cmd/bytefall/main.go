package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanif/bytefall/pkg/capture"
	"github.com/sanif/bytefall/pkg/data"
	"github.com/sanif/bytefall/pkg/ui"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	// Parse flags
	iface := flag.String("i", capture.DefaultInterface(), "Network interface to capture from")
	listIfaces := flag.Bool("list", false, "List available network interfaces")
	demo := flag.Bool("demo", false, "Run in demo mode without packet capture")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")
	completion := flag.String("completion", "", "Generate shell completion (bash, zsh, fish)")
	theme := flag.String("theme", "matrix", "Color theme (matrix, cyberpunk, amber, ocean, blood)")

	// Widget mode flags (mutually exclusive)
	matrixOnly := flag.Bool("matrix", false, "Fullscreen matrix rain widget")
	speedOnly := flag.Bool("speed", false, "Fullscreen network speed widget")
	leaderboardOnly := flag.Bool("leaderboard", false, "Fullscreen domain leaderboard widget")
	processesOnly := flag.Bool("processes", false, "Fullscreen process map widget")
	timelineOnly := flag.Bool("timeline", false, "Fullscreen activity timeline widget")
	graphOnly := flag.Bool("graph", false, "Fullscreen connection graph widget")
	topAppsOnly := flag.Bool("apps", false, "Fullscreen top applications widget")
	bandwidthOnly := flag.Bool("bandwidth", false, "Fullscreen bandwidth history graph")

	// Speed widget style
	speedStyleFlag := flag.String("speed-style", "boxed", "Speed widget style (minimal, boxed, retro, neon, compact)")

	// Status bar options (off by default in widget mode)
	showBar := flag.Bool("bar", false, "Show status bar in widget mode")
	showDownload := flag.Bool("down", true, "Show download speed in status bar")
	showUpload := flag.Bool("up", true, "Show upload speed in status bar")
	showDomains := flag.Bool("domains", true, "Show domain count in status bar")
	showIP := flag.Bool("ip", false, "Show IP address in status bar")
	showPublicIP := flag.Bool("public-ip", false, "Show public IP address in status bar")

	// Legacy flags (kept for compatibility)
	showInfo := flag.Bool("info", true, "Show status bar with network info (legacy)")
	minimal := flag.Bool("minimal", false, "Minimal mode - no status bar at all (legacy)")
	flag.Parse()

	// Help
	if *showHelp {
		printHelp()
		return
	}

	// Version
	if *showVersion {
		fmt.Printf("bytefall version %s\n", version)
		return
	}

	// Shell completion
	if *completion != "" {
		generateCompletion(*completion)
		return
	}

	// List interfaces mode
	if *listIfaces {
		interfaces, err := capture.ListInterfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing interfaces: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Available interfaces:")
		for _, iface := range interfaces {
			fmt.Printf("  %s\n", iface)
		}
		return
	}

	// Set theme
	ui.SetTheme(*theme)

	// Determine widget mode (mutually exclusive, first one wins)
	widgetMode := ui.WidgetNone
	switch {
	case *matrixOnly:
		widgetMode = ui.WidgetMatrix
	case *speedOnly:
		widgetMode = ui.WidgetSpeed
	case *leaderboardOnly:
		widgetMode = ui.WidgetLeaderboard
	case *processesOnly:
		widgetMode = ui.WidgetProcesses
	case *timelineOnly:
		widgetMode = ui.WidgetTimeline
	case *graphOnly:
		widgetMode = ui.WidgetGraph
	case *topAppsOnly:
		widgetMode = ui.WidgetTopApps
	case *bandwidthOnly:
		widgetMode = ui.WidgetBandwidth
	}

	// Set widget mode and options
	ui.SetWidgetMode(widgetMode)

	// Set speed widget style
	switch *speedStyleFlag {
	case "minimal":
		ui.SetSpeedStyle(ui.SpeedStyleMinimal)
	case "boxed":
		ui.SetSpeedStyle(ui.SpeedStyleBoxed)
	case "retro":
		ui.SetSpeedStyle(ui.SpeedStyleRetro)
	case "neon":
		ui.SetSpeedStyle(ui.SpeedStyleNeon)
	case "compact":
		ui.SetSpeedStyle(ui.SpeedStyleCompact)
	}

	// Determine if status bar should be shown
	// -bar explicitly enables it, -minimal explicitly disables it
	// Legacy -info is respected if neither -bar nor -minimal is set
	showStatusBar := *showBar
	if *minimal {
		showStatusBar = false
	} else if !*showBar && *showInfo && widgetMode != ui.WidgetNone {
		// Legacy: if -info is true (default) and no explicit -bar, don't show bar by default in widget mode
		showStatusBar = false
	}

	ui.SetWidgetOptions(ui.WidgetOptions{
		ShowStatusBar: showStatusBar,
		ShowDownload:  *showDownload,
		ShowUpload:    *showUpload,
		ShowDomains:   *showDomains,
		ShowIP:        *showIP,
		ShowPublicIP:  *showPublicIP,
	})

	// Legacy compatibility
	ui.SetMatrixOptions(ui.MatrixOptions{
		ShowInfo:     *showInfo && !*minimal,
		ShowDownload: *showDownload,
		ShowUpload:   *showUpload,
		ShowDomains:  *showDomains,
		ShowIP:       *showIP,
		ShowPublicIP: *showPublicIP,
		Minimal:      *minimal,
	})

	// Demo mode - run without actual packet capture
	if *demo {
		runDemo()
		return
	}

	// Check for root privileges (needed for packet capture)
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "ByteFall requires root privileges for packet capture.\n")
		fmt.Fprintf(os.Stderr, "Run with: sudo bytefall\n")
		fmt.Fprintf(os.Stderr, "Or use demo mode: bytefall -demo\n")
		os.Exit(1)
	}

	// Initialize capture
	cap := capture.New(*iface)
	if err := cap.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start capture: %v\n", err)
		os.Exit(1)
	}
	defer cap.Stop()

	// Initialize process mapper
	pm := data.NewProcessMapper(time.Second)
	pm.Start()
	defer pm.Stop()

	// Initialize aggregator
	agg := data.NewAggregator(pm)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agg.Start(ctx, cap.Packets())

	// Create and run TUI
	model := ui.NewModel(cap, pm, agg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func runDemo() {
	// Create demo aggregator with simulated traffic
	demo := data.NewDemoAggregator()
	demo.Start()
	defer demo.Stop()

	model := ui.NewModel(nil, nil, demo)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func generateCompletion(shell string) {
	switch shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s (supported: bash, zsh, fish)\n", shell)
		os.Exit(1)
	}
}

func printHelp() {
	help := `ByteFall v` + version + ` - Real-time network traffic visualization

USAGE:
    sudo bytefall [OPTIONS]
    bytefall -demo [OPTIONS]

OPTIONS:
    -i <interface>    Network interface to capture from (default: auto-detect)
    -list             List available network interfaces
    -demo             Run in demo mode with simulated traffic (no sudo needed)
    -theme <name>     Color theme: matrix, cyberpunk, amber, ocean, blood
    -version          Show version
    -help             Show this help message
    -completion <sh>  Generate shell completion (bash, zsh, fish)

WIDGET MODES (fullscreen, clean by default):
    -matrix           Fullscreen matrix rain widget
    -speed            Fullscreen network speed widget with gauges
    -leaderboard      Fullscreen domain rankings widget
    -processes        Fullscreen process map widget
    -timeline         Fullscreen activity timeline widget
    -graph            Fullscreen connection graph widget
    -apps             Fullscreen top applications by bandwidth
    -bandwidth        Fullscreen bandwidth history graph

SPEED WIDGET STYLES:
    -speed-style <s>  Style for speed widget: minimal, boxed, retro, neon, compact

STATUS BAR OPTIONS (for widget modes):
    -bar              Show status bar in widget mode (off by default)
    -down             Show download speed (default: true)
    -up               Show upload speed (default: true)
    -domains          Show active domain count (default: true)
    -ip               Show local IP address
    -public-ip        Show public IP address

KEY BINDINGS:
    q, Ctrl+C         Quit
    p                 Pause/Resume capture
    r                 Reset statistics
    Tab               Focus next panel
    Shift+Tab         Focus previous panel
    f                 Toggle fullscreen for focused panel
    t                 Cycle through themes
    s                 Run speed test
    d                 Show domain details
    g                 Toggle connection graph
    ?                 Show help overlay

EXAMPLES:
    sudo bytefall                         # Run with default settings
    bytefall -demo                        # Demo mode (no privileges needed)
    bytefall -demo -matrix                # Matrix rain widget
    bytefall -demo -speed                 # Network speed widget
    bytefall -demo -speed -speed-style neon      # Neon style speed widget
    bytefall -demo -apps                  # Top applications by bandwidth
    bytefall -demo -bandwidth             # Bandwidth history graph
    bytefall -demo -leaderboard           # Domain leaderboard
    bytefall -demo -timeline              # Activity timeline
    bytefall -demo -graph                 # Connection graph

SHELL COMPLETION:
    # Bash (add to ~/.bashrc)
    eval "$(bytefall -completion bash)"

    # Zsh (add to ~/.zshrc)
    eval "$(bytefall -completion zsh)"

    # Fish (add to ~/.config/fish/config.fish)
    bytefall -completion fish | source

For more information: https://github.com/sanif/bytefall
`
	fmt.Print(help)
}

const bashCompletion = `# bytefall bash completion
_bytefall() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-i -list -demo -matrix -speed -leaderboard -processes -timeline -graph -apps -bandwidth -theme -speed-style -bar -down -up -domains -ip -public-ip -version -completion"

    case "${prev}" in
        -i)
            # Complete with network interfaces
            COMPREPLY=( $(compgen -W "$(bytefall -list 2>/dev/null | tail -n +2 | tr -d ' ')" -- ${cur}) )
            return 0
            ;;
        -theme)
            COMPREPLY=( $(compgen -W "matrix cyberpunk amber ocean blood" -- ${cur}) )
            return 0
            ;;
        -speed-style)
            COMPREPLY=( $(compgen -W "minimal boxed retro neon compact" -- ${cur}) )
            return 0
            ;;
        -completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
    esac

    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}
complete -F _bytefall bytefall
`

const zshCompletion = `#compdef bytefall

_bytefall() {
    local -a opts themes shells
    opts=(
        '-i[Network interface to capture from]:interface:->interfaces'
        '-list[List available network interfaces]'
        '-demo[Run in demo mode without packet capture]'
        '-matrix[Fullscreen matrix rain widget]'
        '-speed[Fullscreen network speed widget]'
        '-leaderboard[Fullscreen domain leaderboard widget]'
        '-processes[Fullscreen process map widget]'
        '-timeline[Fullscreen activity timeline widget]'
        '-graph[Fullscreen connection graph widget]'
        '-apps[Fullscreen top applications widget]'
        '-bandwidth[Fullscreen bandwidth history graph]'
        '-theme[Color theme]:theme:->themes'
        '-speed-style[Speed widget style]:style:->speedstyles'
        '-bar[Show status bar in widget mode]'
        '-down[Show download speed in status bar]'
        '-up[Show upload speed in status bar]'
        '-domains[Show domain count in status bar]'
        '-ip[Show IP address in status bar]'
        '-public-ip[Show public IP address in status bar]'
        '-version[Show version]'
        '-completion[Generate shell completion]:shell:->shells'
    )
    themes=(matrix cyberpunk amber ocean blood)
    speedstyles=(minimal boxed retro neon compact)
    shells=(bash zsh fish)

    _arguments -s $opts

    case "$state" in
        interfaces)
            local -a ifaces
            ifaces=(${(f)"$(bytefall -list 2>/dev/null | tail -n +2 | tr -d ' ')"})
            _describe 'interface' ifaces
            ;;
        themes)
            _describe 'theme' themes
            ;;
        speedstyles)
            _describe 'style' speedstyles
            ;;
        shells)
            _describe 'shell' shells
            ;;
    esac
}

_bytefall "$@"
`

const fishCompletion = `# bytefall fish completion
complete -c bytefall -f

complete -c bytefall -s i -d 'Network interface to capture from' -xa "(bytefall -list 2>/dev/null | tail -n +2 | string trim)"
complete -c bytefall -l list -d 'List available network interfaces'
complete -c bytefall -l demo -d 'Run in demo mode without packet capture'
complete -c bytefall -l matrix -d 'Fullscreen matrix rain widget'
complete -c bytefall -l speed -d 'Fullscreen network speed widget'
complete -c bytefall -l leaderboard -d 'Fullscreen domain leaderboard widget'
complete -c bytefall -l processes -d 'Fullscreen process map widget'
complete -c bytefall -l timeline -d 'Fullscreen activity timeline widget'
complete -c bytefall -l graph -d 'Fullscreen connection graph widget'
complete -c bytefall -l apps -d 'Fullscreen top applications widget'
complete -c bytefall -l bandwidth -d 'Fullscreen bandwidth history graph'
complete -c bytefall -l theme -d 'Color theme' -xa "matrix cyberpunk amber ocean blood"
complete -c bytefall -l speed-style -d 'Speed widget style' -xa "minimal boxed retro neon compact"
complete -c bytefall -l bar -d 'Show status bar in widget mode'
complete -c bytefall -l down -d 'Show download speed in status bar'
complete -c bytefall -l up -d 'Show upload speed in status bar'
complete -c bytefall -l domains -d 'Show domain count in status bar'
complete -c bytefall -l ip -d 'Show IP address in status bar'
complete -c bytefall -l public-ip -d 'Show public IP address in status bar'
complete -c bytefall -l version -d 'Show version'
complete -c bytefall -l completion -d 'Generate shell completion' -xa "bash zsh fish"
`
