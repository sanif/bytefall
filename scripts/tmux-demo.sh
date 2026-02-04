#!/bin/bash
# ByteFall Demo - Shows all widgets in a tmux grid
# Usage: ./scripts/tmux-demo.sh

SESSION="bytefall-demo"
BYTEFALL="${BYTEFALL:-bytefall}"

# Check if bytefall is available
if ! command -v "$BYTEFALL" &> /dev/null; then
    # Try local build
    if [[ -f "./bytefall" ]]; then
        BYTEFALL="./bytefall"
    else
        echo "bytefall not found. Build it first: go build -o bytefall ./cmd/bytefall"
        exit 1
    fi
fi

# Kill existing session if it exists
tmux kill-session -t "$SESSION" 2>/dev/null

# Create new session with first widget (matrix)
tmux new-session -d -s "$SESSION" -x "$(tput cols)" -y "$(tput lines)"

# Layout: 2 rows x 4 columns
#  ┌─────────┬─────────┬─────────┬─────────┐
#  │ matrix  │  speed  │  apps   │bandwidth│
#  ├─────────┼─────────┼─────────┼─────────┤
#  │ leader  │processes│timeline │  graph  │
#  └─────────┴─────────┴─────────┴─────────┘

# First pane: matrix (already created)
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -matrix" C-m

# Split horizontally for speed
tmux split-window -h -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -speed -speed-style neon" C-m

# Split for apps
tmux split-window -h -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -apps" C-m

# Split for bandwidth
tmux split-window -h -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -bandwidth" C-m

# Select first pane and split vertically for leaderboard
tmux select-pane -t "$SESSION:0.0"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -leaderboard" C-m

# Select second pane (speed) and split for processes
tmux select-pane -t "$SESSION:0.2"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -processes" C-m

# Select third pane (apps) and split for timeline
tmux select-pane -t "$SESSION:0.4"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -timeline" C-m

# Select fourth pane (bandwidth) and split for graph
tmux select-pane -t "$SESSION:0.6"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -graph" C-m

# Even out the layout
tmux select-layout -t "$SESSION" tiled

# Attach to session
tmux attach-session -t "$SESSION"
