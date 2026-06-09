#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Setup tmux
sudo apt-get update
sudo apt-get install -y tmux


# Setup agents
sudo chown -R vscode:vscode /home/vscode/.claude

curl -fsSL https://claude.ai/install.sh | bash

# curl -fsSL https://gh.io/copilot-install | bash


# Setup bash

# Setup custom prompt - hybrid of local + container features
cat >> ~/.bashrc << 'EOF'

cls ()
{
    clear && printf '\033[3J'
}

# Custom prompt - hybrid of local + container features
export PS1='\[\]`export XIT=$?; [ "$XIT" -ne 0 ] && echo -n "\[\033[1;31m\]" || echo -n "\[\033[0m\]"`container`export FOLDER=$(basename "$PWD"); export BRANCH="$(git --no-optional-locks symbolic-ref --short HEAD 2>/dev/null || git --no-optional-locks rev-parse --short HEAD 2>/dev/null)"; if [ "${BRANCH:-}" != "" ]; then [ "$FOLDER" != "$BRANCH" ] && echo -n " \[\033[32m\]$FOLDER"; echo -n " \[\033[33m\]($BRANCH)"; else echo -n " \[\033[32m\]$FOLDER"; fi`\[\033[00m\] $ \[\]'

# Change iTerm2 tab color to green cleanly without empty lines
if [ -n "$ITERM_SESSION_ID" ] || [ "$TERM_PROGRAM" = "iTerm.app" ] || [ -n "$TMUX" ]; then
  echo -ne "\033]6;1;bg;red;brightness;46\a"
  echo -ne "\033]6;1;bg;green;brightness;204\a"
  echo -ne "\033]6;1;bg;blue;brightness;113\a"
fi
EOF

