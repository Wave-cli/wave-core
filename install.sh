#!/usr/bin/env bash
if [ -z "${BASH_VERSION:-}" ]; then
  exec /usr/bin/env bash "$0" "$@"
fi
set -euo pipefail

# --- Configuration ---
REPO="Wave-cli/wave-core"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

# Global variable for cleanup
tmp_dir=""

# --- Helper Functions ---

die() {
  printf "\033[0;31mError: %s\033[0m\n" "$*" >&2
  exit 1
}

have() {
  command -v "$1" >/dev/null 2>&1
}

detect_platform() {
  local os arch
  have uname || die "uname is required to detect platform"
  os=$(uname -s)
  arch=$(uname -m)

  case "${os}" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) die "Unsupported OS: ${os}" ;;
  esac

  case "${arch}" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) die "Unsupported architecture: ${arch}" ;;
  esac
  printf "%s %s" "${os}" "${arch}"
}

select_http_client() {
  if have curl; then
    fetch() { curl -fsSL "$1"; }
    download() { curl -fsSL -o "$2" "$1"; }
  elif have wget; then
    fetch() { wget -qO- "$1"; }
    download() { wget -qO "$2" "$1"; }
  else
    die "curl or wget is required"
  fi
}

get_installed_path() {
  if have wave; then
    command -v wave
  else
    return 1
  fi
}

do_uninstall() {
  local binary_path
  binary_path=$(get_installed_path) || {
    echo "wave is not installed."
    exit 0
  }

  echo "Found wave at ${binary_path}. Uninstalling..."
  if [ -w "${binary_path}" ]; then
    rm "${binary_path}"
  else
    echo "Requesting sudo to remove ${binary_path}..."
    sudo rm "${binary_path}"
  fi
  echo "Successfully uninstalled wave."
  exit 0
}

resolve_install_dir() {
  if [ -n "${WAVE_INSTALL_DIR:-}" ]; then
    echo "${WAVE_INSTALL_DIR}"
    return
  fi

  if [ -t 0 ]; then
    echo "------------------------------------------------" >&2
    echo "Where would you like to install wave?" >&2
    echo "1) Only for me (~/.local/bin)" >&2
    echo "2) System-wide (/usr/local/bin) - Needs sudo" >&2
    echo "------------------------------------------------" >&2

    printf "Select [1-2, default 1]: " >&2
    read -r choice

    if [ "${choice}" = "2" ]; then
      echo "/usr/local/bin"
      return
    fi
  fi

  if [ "$(uname -s)" = "Darwin" ]; then
    echo "/usr/local/bin"
  else
    echo "${HOME}/.local/bin"
  fi
}

# --- The New Banner ---

print_banner() {
  printf "\033[0;36m"
  cat <<'EOF'

  ██╗    ██╗ █████╗ ██╗   ██╗███████╗     ██████╗██╗     ██╗
  ██║    ██║██╔══██╗██║   ██║██╔════╝    ██╔════╝██║     ██║
  ██║ █╗ ██║███████║██║   ██║█████╗      ██║     ██║     ██║
  ██║███╗██║██╔══██║╚██╗ ██╔╝██╔══╝      ██║     ██║     ██║
  ╚███╔███╔╝██║  ██║ ╚████╔╝ ███████╗    ╚██████╗███████╗██║
   ╚══╝╚══╝ ╚═╝  ╚═╝  ╚═══╝  ╚══════╝     ╚═════╝╚══════╝╚═╝

EOF
  printf "\033[0m\n"
}

# --- Main Logic ---

main() {
  if [[ "${1:-}" == "--uninstall" ]]; then
    do_uninstall
  fi

  print_banner

  select_http_client

  local platform os arch asset_name release_json latest_tag download_url install_dir
  local archive_path

  platform=$(detect_platform)
  os=$(echo "${platform}" | cut -d' ' -f1)
  arch=$(echo "${platform}" | cut -d' ' -f2)
  asset_name="wave-${os}-${arch}.tar.gz"

  echo "Checking for latest release..."
  release_json=$(fetch "${API_URL}")

  latest_tag=$(echo "${release_json}" | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/' | head -n 1)
  download_url=$(echo "${release_json}" | grep -o "https://github.com/${REPO}/releases/download/[^\" ]*${asset_name}" | head -n 1)

  if [ -z "${download_url}" ]; then
    die "Could not find asset ${asset_name} in release ${latest_tag:-unknown}"
  fi

  install_dir=$(resolve_install_dir)

  tmp_dir=$(mktemp -d)
  trap 'if [ -n "${tmp_dir:-}" ]; then rm -rf "${tmp_dir}"; fi' EXIT

  archive_path="${tmp_dir}/${asset_name}"

  echo "Downloading wave ${latest_tag}..."
  download "${download_url}" "${archive_path}"

  echo "Extracting..."
  tar -xzf "${archive_path}" -C "${tmp_dir}"

  if [ ! -f "${tmp_dir}/wave" ]; then
    die "Binary 'wave' not found in archive."
  fi

  if [ ! -d "${install_dir}" ]; then
    if [ -w "$(dirname "${install_dir}" 2>/dev/null || echo ".")" ]; then
      mkdir -p "${install_dir}"
    else
      sudo mkdir -p "${install_dir}"
    fi
  fi

  if [ -w "${install_dir}" ]; then
    chmod +x "${tmp_dir}/wave"
    mv "${tmp_dir}/wave" "${install_dir}/wave"
  else
    echo "Requesting sudo to install..."
    sudo chmod +x "${tmp_dir}/wave"
    sudo mv "${tmp_dir}/wave" "${install_dir}/wave"
  fi

  echo -e "\n\033[0;32m✓ wave ${latest_tag} installed to ${install_dir}/wave\033[0m"

  case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *)
    echo -e "\033[0;33mNote:\033[0m Add ${install_dir} to your PATH."
    ;;
  esac
}

main "$@"
