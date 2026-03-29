#!/bin/bash
# Build and setup script for wave-core and wave-flow
# Builds both projects, installs wave and wave-flow plugin

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

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# =============================================================================
# Build wave-core
# =============================================================================

section "Building wave-core"
cd "$WAVE_CORE"
go clean
VERSION=$(git describe --tags --always 2>/dev/null | sed 's/-[0-9]*-g.*//')
go build -ldflags "-X github.com/wave-cli/wave-core/internal/version.version=${VERSION:-dev} -X github.com/wave-cli/wave-core/internal/version.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'none') -X github.com/wave-cli/wave-core/internal/version.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o wave .
success "Built: $WAVE_CORE/wave"

# =============================================================================
# Install wave to ~/.local/bin (or /usr/local/bin)
# =============================================================================

section "Installing wave"

# Determine install directory
if [ "$(uname -s)" = "Darwin" ]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="$HOME/.local/bin"
fi

# Create directory if needed
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -w "$(dirname "$INSTALL_DIR")" ]; then
        mkdir -p "$INSTALL_DIR"
    else
        warn "Need sudo to create $INSTALL_DIR"
        sudo mkdir -p "$INSTALL_DIR"
    fi
fi

# Install wave
if [ -w "$INSTALL_DIR" ]; then
    chmod +x "$WAVE_CORE/wave"
    cp "$WAVE_CORE/wave" "$INSTALL_DIR/wave"
else
    warn "Need sudo to install to $INSTALL_DIR"
    sudo chmod +x "$WAVE_CORE/wave"
    sudo cp "$WAVE_CORE/wave" "$INSTALL_DIR/wave"
fi

success "Installed to: $INSTALL_DIR/wave"

# Check if PATH includes install dir
case ":${PATH}:" in
*":${INSTALL_DIR}:"*)
    ;;
*)
    warn "Add $INSTALL_DIR to your PATH"
    ;;
esac

# =============================================================================
# Build wave-flow
# =============================================================================

section "Building wave-flow"
cd "$WAVE_FLOW"
go clean
VERSION=$(git describe --tags --always 2>/dev/null | sed 's/-[0-9]*-g.*//')
go build -ldflags "-X github.com/wave-cli/wave-core/internal/version.version=${VERSION:-dev}" -o bin/flow .
success "Built: $WAVE_FLOW/bin/flow"

# =============================================================================
# Install wave-flow to ~/.wave/plugins
# =============================================================================

section "Installing wave-flow plugin"

FLOW_VERSION=$(grep 'version = ' "$WAVE_FLOW/Waveplugin" | tr -d '"' | awk '{print $3}')
success "Flow version: $FLOW_VERSION"

PLUGIN_DIR="$HOME/.wave/plugins/wave-cli/flow"
mkdir -p "$PLUGIN_DIR/bin"

cp "$WAVE_FLOW/bin/flow" "$PLUGIN_DIR/bin/flow"
cp "$WAVE_FLOW/Waveplugin" "$PLUGIN_DIR/Waveplugin"

success "Installed to: $PLUGIN_DIR"

# =============================================================================
# Update config
# =============================================================================

section "Updating config"

WAVE_CONFIG="$HOME/.wave/config"

if [ ! -f "$WAVE_CONFIG" ]; then
    mkdir -p "$HOME/.wave"
    cat > "$WAVE_CONFIG" << 'EOF'
[core]
logs_dir = "~/.wave/logs"

[plugins]
  "flow/flow" = "local"
EOF
    success "Created config: $WAVE_CONFIG"
else
    # Check if flow is already registered
    if grep -qE '^\s*"flow/flow"\s*=' "$WAVE_CONFIG" || grep -qE '^\s*"flow"\s*=' "$WAVE_CONFIG"; then
        success "Flow already in config"
    else
        if grep -q '\[plugins\]' "$WAVE_CONFIG"; then
            sed -i '/\[plugins\]/a\  "flow/flow" = "local"' "$WAVE_CONFIG"
            success "Added flow to plugins"
        else
            echo '[plugins]' >> "$WAVE_CONFIG"
            echo '  "flow/flow" = "local"' >> "$WAVE_CONFIG"
            success "Added plugins section"
        fi
    fi
fi

# =============================================================================
# Verify
# =============================================================================

section "Verification"

echo ""
echo "Testing wave:"
"$INSTALL_DIR/wave" version

echo ""
echo "Testing flow plugin:"
cd "$WAVE_FLOW/../test"
"$INSTALL_DIR/wave" flow --list || echo "(No Wavefile in test dir - expected)"

success "All done!"
echo ""
echo "Usage:"
echo "  wave flow <command>"
echo "  wave flow --list"
echo ""
echo "If wave is not found, run:"
echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
