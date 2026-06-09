#!/bin/bash
set -e

cd "$(dirname "$0")/.."


# Setup tmux
sudo apt-get update
sudo apt-get install -y tmux


# Setup Node + nvm

sudo chown -R vscode:vscode $PROJECT_PATH/node_modules

# Install the Node version specified in .nvmrc
export NVM_DIR="/usr/local/share/nvm"
# shellcheck source=/dev/null
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
if [ -f .nvmrc ]; then
    nvm install
fi


# Install Playwright browser binaries + native system dependencies
# Must run after Node is available; browsers are stored in the container layer
if [ -f package.json ]; then
    npm install
    npx playwright install --with-deps
fi


# Setup agents
sudo chown -R vscode:vscode /home/vscode/.claude

curl -fsSL https://claude.ai/install.sh | bash

# curl -fsSL https://gh.io/copilot-install | bash


# Setup bash

# Setup custom prompt - hybrid of local + container features
cat >> ~/.bashrc << 'EOF'

# Automatically use the correct Node version when entering the project
export NVM_DIR="/usr/local/share/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
cd() { builtin cd "$@" && [ -f .nvmrc ] && nvm use --silent; }

# Auto-switch to .nvmrc version on shell start
[ -f .nvmrc ] && nvm use --silent

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

