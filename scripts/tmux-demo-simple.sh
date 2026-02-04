#!/bin/bash
# ByteFall Demo (Simple 2x2 grid)
# Usage: ./scripts/tmux-demo-simple.sh

SESSION="bytefall-demo"
BYTEFALL="${BYTEFALL:-bytefall}"

# Check if bytefall is available
if ! command -v "$BYTEFALL" &> /dev/null; then
    if [[ -f "./bytefall" ]]; then
        BYTEFALL="./bytefall"
    else
        echo "bytefall not found. Build it first: go build -o bytefall ./cmd/bytefall"
        exit 1
    fi
fi

# Kill existing session
tmux kill-session -t "$SESSION" 2>/dev/null

# Create session
tmux new-session -d -s "$SESSION"

# Layout: 2x2
#  ┌───────────┬───────────┐
#  │  matrix   │   speed   │
#  ├───────────┼───────────┤
#  │   apps    │ bandwidth │
#  └───────────┴───────────┘

# Top-left: matrix
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -matrix -theme matrix" C-m

# Top-right: speed
tmux split-window -h -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -speed -theme cyberpunk" C-m

# Bottom-left: apps
tmux select-pane -t "$SESSION:0.0"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -apps -theme ocean" C-m

# Bottom-right: bandwidth
tmux select-pane -t "$SESSION:0.1"
tmux split-window -v -t "$SESSION"
tmux send-keys -t "$SESSION" "$BYTEFALL -demo -bandwidth -theme amber" C-m

# Attach
tmux attach-session -t "$SESSION"
