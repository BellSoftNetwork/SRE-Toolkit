#!/bin/bash

# Claude Code Account Switcher
# Enhanced with email-based auto-save and alias support

CLAUDE_DIR="$HOME/.claude"
ACCOUNTS_DIR="$CLAUDE_DIR/accounts"
CREDENTIALS_FILE="$CLAUDE_DIR/.credentials.json"
CONFIG_FILE="$HOME/.claude.json"
STATSIG_DIR="$CLAUDE_DIR/statsig"
SETTINGS_FILE="$CLAUDE_DIR/settings.json"

# Ensure accounts directory exists
mkdir -p "$ACCOUNTS_DIR"

# Function to get account UUID from credentials
get_account_uuid() {
    if [ -f "$CONFIG_FILE" ]; then
        # Extract account UUID from oauthAccount section
        local uuid=$(grep -o '"accountUuid":[[:space:]]*"[^"]*"' "$CONFIG_FILE" 2>/dev/null | sed 's/.*"accountUuid":[[:space:]]*"\([^"]*\)".*/\1/')
        echo "$uuid"
    fi
}

# Function to get email from config
get_email() {
    if [ -f "$CONFIG_FILE" ]; then
        # Extract email from oauthAccount section
        grep -o '"emailAddress":[[:space:]]*"[^"]*"' "$CONFIG_FILE" 2>/dev/null | sed 's/.*"emailAddress":[[:space:]]*"\([^"]*\)".*/\1/'
    fi
}

# Function to save account metadata
save_account_metadata() {
    local account_dir="$1"
    local alias="$2"
    local email="$3"
    local uuid="$4"
    
    cat > "$account_dir/metadata.json" << EOF
{
    "uuid": "$uuid",
    "email": "$email",
    "alias": "$alias",
    "saved_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
}

# Function to save current account
save_account() {
    local alias="$1"
    
    # Get account info
    local uuid=$(get_account_uuid)
    local email=$(get_email)
    
    if [ -z "$uuid" ]; then
        echo "Error: Cannot determine account UUID. Make sure you're logged in."
        exit 1
    fi
    
    # Use email as default if no alias provided
    if [ -z "$alias" ]; then
        if [ -z "$email" ]; then
            echo "Error: No email found and no alias provided. Please provide an alias."
            echo "Usage: $0 save [alias]"
            exit 1
        fi
        alias="$email"
    fi
    
    echo "Saving account: $alias (UUID: ${uuid:0:8}...)"
    
    # Create account directory
    local account_dir="$ACCOUNTS_DIR/$uuid"
    mkdir -p "$account_dir"
    
    # Save all files
    [ -f "$CREDENTIALS_FILE" ] && cp "$CREDENTIALS_FILE" "$account_dir/credentials.json"
    [ -f "$CONFIG_FILE" ] && cp "$CONFIG_FILE" "$account_dir/claude.json"
    [ -f "$SETTINGS_FILE" ] && cp "$SETTINGS_FILE" "$account_dir/settings.json"
    
    # Save statsig directory
    if [ -d "$STATSIG_DIR" ] && [ "$(ls -A $STATSIG_DIR)" ]; then
        rm -rf "$account_dir/statsig" 2>/dev/null || true
        cp -r "$STATSIG_DIR" "$account_dir/statsig"
    fi
    
    # Save metadata
    save_account_metadata "$account_dir" "$alias" "$email" "$uuid"
    
    echo "✓ Account saved: $alias"
}

# Function to find account by email or alias
find_account() {
    local query="$1"
    
    # Search through all account directories
    for account_dir in "$ACCOUNTS_DIR"/*; do
        if [ -f "$account_dir/metadata.json" ]; then
            # Check if query matches email or alias
            if grep -q "\"email\": \"$query\"" "$account_dir/metadata.json" 2>/dev/null || \
               grep -q "\"alias\": \"$query\"" "$account_dir/metadata.json" 2>/dev/null; then
                echo "$account_dir"
                return 0
            fi
        fi
    done
    
    return 1
}

# Function to switch account
switch_account() {
    local query="$1"
    
    # Find account directory
    local account_dir=$(find_account "$query")
    
    if [ -z "$account_dir" ]; then
        echo "Error: Account '$query' not found!"
        list_accounts
        exit 1
    fi
    
    # Get account info
    local metadata=$(cat "$account_dir/metadata.json" 2>/dev/null)
    local email=$(echo "$metadata" | grep -o '"email":[[:space:]]*"[^"]*"' | sed 's/.*"email":[[:space:]]*"\([^"]*\)".*/\1/')
    local alias=$(echo "$metadata" | grep -o '"alias":[[:space:]]*"[^"]*"' | sed 's/.*"alias":[[:space:]]*"\([^"]*\)".*/\1/')
    
    echo "Switching to account: $alias ($email)"
    
    # Clear current session data
    echo "Clearing session data..."
    rm -rf "$STATSIG_DIR"
    mkdir -p "$STATSIG_DIR"
    
    # Restore account files
    [ -f "$account_dir/credentials.json" ] && cp "$account_dir/credentials.json" "$CREDENTIALS_FILE"
    [ -f "$account_dir/claude.json" ] && cp "$account_dir/claude.json" "$CONFIG_FILE"
    [ -f "$account_dir/settings.json" ] && cp "$account_dir/settings.json" "$SETTINGS_FILE"
    
    # Restore statsig data if exists
    if [ -d "$account_dir/statsig" ]; then
        echo "Restoring session data..."
        cp -r "$account_dir/statsig"/* "$STATSIG_DIR"/ 2>/dev/null || true
    fi
    
    echo "✓ Switched to account: $alias"
    echo "Note: Restart Claude Code to ensure clean session state"
}

# Function to list accounts
list_accounts() {
    echo "Available accounts:"
    echo ""
    
    local found=0
    for account_dir in "$ACCOUNTS_DIR"/*; do
        if [ -f "$account_dir/metadata.json" ]; then
            local metadata=$(cat "$account_dir/metadata.json")
            local email=$(echo "$metadata" | grep -o '"email":[[:space:]]*"[^"]*"' | sed 's/.*"email":[[:space:]]*"\([^"]*\)".*/\1/')
            local alias=$(echo "$metadata" | grep -o '"alias":[[:space:]]*"[^"]*"' | sed 's/.*"alias":[[:space:]]*"\([^"]*\)".*/\1/')
            local uuid=$(echo "$metadata" | grep -o '"uuid":[[:space:]]*"[^"]*"' | sed 's/.*"uuid":[[:space:]]*"\([^"]*\)".*/\1/')
            
            echo "  • $alias"
            echo "    Email: $email"
            echo "    UUID: ${uuid:0:8}..."
            echo ""
            found=1
        fi
    done
    
    if [ $found -eq 0 ]; then
        echo "  No saved accounts found"
    fi
}

# Main logic
case "$1" in
    save)
        save_account "$2"
        ;;
    switch|load)
        if [ -z "$2" ]; then
            echo "Usage: $0 switch <email|alias>"
            list_accounts
            exit 1
        fi
        switch_account "$2"
        ;;
    list|ls)
        list_accounts
        ;;
    current)
        email=$(get_email)
        uuid=$(get_account_uuid)
        if [ -n "$email" ]; then
            echo "Current account: $email"
            echo "UUID: ${uuid:0:8}..."
        else
            echo "No active account"
        fi
        ;;
    *)
        echo "Claude Code Account Switcher"
        echo ""
        echo "Usage:"
        echo "  $0 save [alias]      Save current account (uses email if no alias)"
        echo "  $0 switch <email|alias>   Switch to saved account"
        echo "  $0 list              List all saved accounts"
        echo "  $0 current           Show current account"
        echo ""
        echo "Examples:"
        echo "  $0 save              # Save with email as identifier"
        echo "  $0 save work         # Save with 'work' alias"
        echo "  $0 switch work       # Switch using alias"
        echo "  $0 switch user@example.com  # Switch using email"
        ;;
esac