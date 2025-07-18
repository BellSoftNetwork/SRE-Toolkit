#!/bin/bash

# Claude Code Account Switching Aliases
# Add this to your ~/.bashrc or ~/.zshrc:
# source ~/programming/projects/git/goland/sre-toolkit/scripts/claude-account-switcher/claude-aliases.sh

# Path to the switch script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_SWITCH_SCRIPT="$SCRIPT_DIR/claude-switcher.sh"

# Function to switch Claude accounts
claude-switch() {
    if [ -z "$1" ]; then
        "$CLAUDE_SWITCH_SCRIPT" list
        return 1
    fi
    
    "$CLAUDE_SWITCH_SCRIPT" switch "$1"
}

# Function to save current account
claude-save() {
    "$CLAUDE_SWITCH_SCRIPT" save "$1"
}

# Show current account
claude-current() {
    "$CLAUDE_SWITCH_SCRIPT" current
}

# List all saved accounts
claude-list() {
    "$CLAUDE_SWITCH_SCRIPT" list
}

# Quick save with default (email)
claude-save-default() {
    "$CLAUDE_SWITCH_SCRIPT" save
}

# Migrate old backups to new format
claude-migrate() {
    echo "Migrating old account backups..."
    
    # Check for old style backups
    for backup in ~/.claude/.credentials.json.* ~/.claude.json.*; do
        if [ -f "$backup" ]; then
            local name=$(basename "$backup" | sed 's/.*\.//')
            echo "Found old backup: $name"
            
            # You would need to manually switch to this account first
            echo "Please switch to account '$name' manually, then run:"
            echo "  claude-save $name"
        fi
    done
}