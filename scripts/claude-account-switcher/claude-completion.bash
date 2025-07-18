#!/bin/bash

# Claude Account Switcher - Bash Completion

_claude_get_accounts() {
    local accounts_dir="$HOME/.claude/accounts"
    local accounts=()
    
    # Get all accounts from metadata files
    if [ -d "$accounts_dir" ]; then
        for metadata in "$accounts_dir"/*/metadata.json; do
            if [ -f "$metadata" ]; then
                # Extract email and alias
                local email=$(grep -o '"email":"[^"]*"' "$metadata" 2>/dev/null | cut -d'"' -f4)
                local alias=$(grep -o '"alias":"[^"]*"' "$metadata" 2>/dev/null | cut -d'"' -f4)
                
                [ -n "$email" ] && accounts+=("$email")
                [ -n "$alias" ] && [ "$alias" != "$email" ] && accounts+=("$alias")
            fi
        done
    fi
    
    echo "${accounts[@]}"
}

_claude_switch() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local accounts=($(_claude_get_accounts))
    
    COMPREPLY=($(compgen -W "${accounts[*]}" -- "$cur"))
}

_claude_save() {
    # No completion for save command (user provides custom alias)
    COMPREPLY=()
}

_claude_switcher() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main commands
    local commands="save switch list current"
    
    # If we're on the first argument, complete commands
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
    else
        # Complete based on the command
        case "$prev" in
            switch|load)
                _claude_switch
                ;;
            save)
                _claude_save
                ;;
            *)
                COMPREPLY=()
                ;;
        esac
    fi
}

# Register completions
complete -F _claude_switch claude-switch
complete -F _claude_save claude-save
complete -F _claude_switcher "$HOME/.claude/scripts/claude-switcher.sh"