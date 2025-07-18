#!/bin/bash

# Claude Account Switcher Installation Script

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_DIR="$HOME/.claude"
INSTALL_DIR="$CLAUDE_DIR/scripts"
SHELL_RC=""

# Detect shell configuration file
detect_shell_rc() {
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    else
        # Try to detect from SHELL variable
        case "$SHELL" in
            */zsh)
                SHELL_RC="$HOME/.zshrc"
                ;;
            */bash)
                SHELL_RC="$HOME/.bashrc"
                ;;
            *)
                echo -e "${YELLOW}Warning: Could not detect shell type. Defaulting to .bashrc${NC}"
                SHELL_RC="$HOME/.bashrc"
                ;;
        esac
    fi
}

# Create installation directory
echo -e "${GREEN}Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

# Copy scripts to installation directory
echo -e "${GREEN}Installing scripts to $INSTALL_DIR...${NC}"
cp "$SCRIPT_DIR/claude-switcher.sh" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/claude-aliases.sh" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/claude-completion.bash" "$INSTALL_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR/claude-completion.zsh" "$INSTALL_DIR/" 2>/dev/null || true
chmod +x "$INSTALL_DIR/claude-switcher.sh"

# Update claude-aliases.sh to use installed location
sed -i 's|CLAUDE_SWITCH_SCRIPT=.*|CLAUDE_SWITCH_SCRIPT="$HOME/.claude/scripts/claude-switcher.sh"|' "$INSTALL_DIR/claude-aliases.sh"

# Detect shell RC file
detect_shell_rc
echo -e "${GREEN}Detected shell configuration file: $SHELL_RC${NC}"

# Check if already sourced
ALIAS_SOURCE_LINE="source \"\$HOME/.claude/scripts/claude-aliases.sh\""
MARKER_COMMENT="# Claude Account Switcher"

if [ -f "$SHELL_RC" ]; then
    if grep -q "$MARKER_COMMENT" "$SHELL_RC" 2>/dev/null; then
        echo -e "${YELLOW}Claude Account Switcher already installed in $SHELL_RC${NC}"
        echo "Updating installation..."
        
        # Remove old installation
        sed -i "/$MARKER_COMMENT/,+1d" "$SHELL_RC"
    fi
fi

# Add to shell RC file
echo -e "${GREEN}Adding aliases to $SHELL_RC...${NC}"

# Determine completion file based on shell
COMPLETION_SOURCE=""
if [[ "$SHELL_RC" == *"zshrc"* ]]; then
    COMPLETION_SOURCE="source \"\$HOME/.claude/scripts/claude-completion.zsh\""
else
    COMPLETION_SOURCE="source \"\$HOME/.claude/scripts/claude-completion.bash\""
fi

cat >> "$SHELL_RC" << EOF

$MARKER_COMMENT
$ALIAS_SOURCE_LINE
$COMPLETION_SOURCE
EOF

# Create uninstall script
cat > "$INSTALL_DIR/uninstall.sh" << 'EOF'
#!/bin/bash

# Claude Account Switcher Uninstall Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${RED}Uninstalling Claude Account Switcher...${NC}"

# Remove from shell RC files
for rc_file in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [ -f "$rc_file" ]; then
        sed -i '/# Claude Account Switcher/,+2d' "$rc_file" 2>/dev/null || true
        echo -e "${GREEN}Removed from $rc_file${NC}"
    fi
done

# Remove scripts
rm -rf "$HOME/.claude/scripts"
echo -e "${GREEN}Removed scripts directory${NC}"

echo -e "${GREEN}Uninstallation complete!${NC}"
echo "Note: Account data in ~/.claude/accounts/ was preserved"
EOF

chmod +x "$INSTALL_DIR/uninstall.sh"

# Success message
echo ""
echo -e "${GREEN}✓ Installation complete!${NC}"
echo ""
echo "Available commands:"
echo "  claude-save [alias]     - Save current account"
echo "  claude-switch <name>    - Switch to saved account"
echo "  claude-list             - List all accounts"
echo "  claude-current          - Show current account"
echo ""
echo "To activate the commands in your current session, run:"
echo -e "  ${YELLOW}source $SHELL_RC${NC}"
echo ""
echo "To uninstall later, run:"
echo -e "  ${YELLOW}$INSTALL_DIR/uninstall.sh${NC}"