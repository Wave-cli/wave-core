#!/bin/bash
# Build and setup script for wave-core and wave-flow
# Builds both projects and installs wave-flow plugin

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Paths
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WAVE_CORE="$SCRIPT_DIR"
WAVE_FLOW="$SCRIPT_DIR/../wave-flow"

section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# =============================================================================
# Build wave-core
# =============================================================================

section "Building wave-core"
cd "$WAVE_CORE"
go build -ldflags "-X github.com/wave-cli/wave-core/internal/version.version=$(git describe --tags --always 2>/dev/null || echo 'dev') -X github.com/wave-cli/wave-core/internal/version.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'none') -X github.com/wave-cli/wave-core/internal/version.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o wave .
success "Built: $WAVE_CORE/wave"

# =============================================================================
# Build wave-flow
# =============================================================================

section "Building wave-flow"
cd "$WAVE_FLOW"
go build -ldflags "-X github.com/wave-cli/wave-core/internal/version.version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o bin/flow .
success "Built: $WAVE_FLOW/bin/flow"

# =============================================================================
# Install wave-flow to ~/.wave/plugins
# =============================================================================

section "Installing wave-flow plugin"

# Get flow version from Waveplugin
FLOW_VERSION=$(grep 'version = ' "$WAVE_FLOW/Waveplugin" | tr -d '"' | awk '{print $3}')
success "Flow version: $FLOW_VERSION"

# Create plugin directory
PLUGIN_DIR="$HOME/.wave/plugins/flow"
mkdir -p "$PLUGIN_DIR/bin"

# Copy binary and Waveplugin
cp "$WAVE_FLOW/bin/flow" "$PLUGIN_DIR/bin/flow"
cp "$WAVE_FLOW/Waveplugin" "$PLUGIN_DIR/Waveplugin"

success "Installed to: $PLUGIN_DIR"

# =============================================================================
# Update config
# =============================================================================

section "Updating config"

WAVE_CONFIG="$HOME/.wave/config"

# Create config if it doesn't exist
if [ ! -f "$WAVE_CONFIG" ]; then
    mkdir -p "$HOME/.wave"
    cat > "$WAVE_CONFIG" << 'EOF'
[core]
logs_dir = "~/.wave/logs"

[plugins]
  "flow" = "local"
EOF
    success "Created config: $WAVE_CONFIG"
else
    # Update existing config
    if grep -q '\[plugins\]' "$WAVE_CONFIG"; then
        if ! grep -qE '"flow"|flow' "$WAVE_CONFIG" | grep -v "wave-cli/flow" | grep -q "flow"; then
            sed -i '/\[plugins\]/a\  "flow" = "local"' "$WAVE_CONFIG"
            success "Added flow to plugins"
        fi
    else
        echo '[plugins]' >> "$WAVE_CONFIG"
        echo '  "flow" = "local"' >> "$WAVE_CONFIG"
        success "Added plugins section"
    fi
fi

# =============================================================================
# Verify
# =============================================================================

section "Verification"

echo ""
echo "Testing wave:"
"$WAVE_CORE/wave" version

echo ""
echo "Testing flow plugin:"
cd "$WAVE_FLOW/../test"
"$WAVE_CORE/wave" flow --list || echo "(No Wavefile in test dir - expected)"

success "All done!"
echo ""
echo "Usage:"
echo "  $WAVE_CORE/wave flow <command>"
echo "  $WAVE_CORE/wave flow --list"
