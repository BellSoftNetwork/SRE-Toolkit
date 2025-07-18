#!/bin/zsh

# Claude Account Switcher - Zsh Completion

_claude_get_accounts() {
    local accounts_dir="$HOME/.claude/accounts"
    local accounts=()
    
    # Get all accounts from metadata files
    if [[ -d "$accounts_dir" ]]; then
        for metadata in "$accounts_dir"/*/metadata.json(N); do
            if [[ -f "$metadata" ]]; then
                # Extract email and alias
                local email=$(grep -o '"email":"[^"]*"' "$metadata" 2>/dev/null | cut -d'"' -f4)
                local alias=$(grep -o '"alias":"[^"]*"' "$metadata" 2>/dev/null | cut -d'"' -f4)
                
                [[ -n "$email" ]] && accounts+=("$email")
                [[ -n "$alias" && "$alias" != "$email" ]] && accounts+=("$alias")
            fi
        done
    fi
    
    echo "${accounts[@]}"
}

_claude_switch() {
    local accounts=($(_claude_get_accounts))
    _describe 'account' accounts
}

_claude_save() {
    # No completion for save command (user provides custom alias)
    return 0
}

_claude_switcher() {
    local -a commands
    commands=(
        'save:Save current account with optional alias'
        'switch:Switch to a saved account'
        'list:List all saved accounts'
        'current:Show current account'
    )
    
    if (( CURRENT == 2 )); then
        _describe 'command' commands
    else
        case "$words[2]" in
            switch|load)
                _claude_switch
                ;;
            save)
                _claude_save
                ;;
        esac
    fi
}

# Register completions
compdef _claude_switch claude-switch
compdef _claude_save claude-save
compdef _claude_switcher "$HOME/.claude/scripts/claude-switcher.sh"