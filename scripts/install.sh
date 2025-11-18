#!/bin/bash
# This script installs the coolify-cli from GitHub releases
# Supports Linux and macOS on amd64/arm64 architectures
# Windows is not supported by this installer

set -e  # Exit on error

# Configuration
REPO="coollabsio/coolify-cli"
BINARY_NAME="coolify"
GLOBAL_INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cleanup trap
TEMP_FILE=""
cleanup() {
  if [ -n "$TEMP_FILE" ] && [ -f "$TEMP_FILE" ]; then
    rm -f "$TEMP_FILE"
  fi
}
trap cleanup EXIT

# Error handler
error_exit() {
  echo -e "${RED}Error: $1${NC}" >&2
  exit 1
}

# Show help
show_help() {
  cat << EOF
Coolify CLI Installer

Usage: $0 [OPTIONS] [VERSION]

OPTIONS:
  --user          Install to ~/.local/bin (no sudo required)
  --help          Show this help message
  --version       Show installer version

ARGUMENTS:
  VERSION         Specific version to install (e.g., v1.0.0)
                  If not specified, installs the latest release

EXAMPLES:
  $0              Install latest version to /usr/local/bin
  $0 --user       Install latest version to ~/.local/bin
  $0 v1.0.0       Install specific version to /usr/local/bin
  $0 --user v1.0.0  Install specific version to ~/.local/bin

EOF
  exit 0
}

# Parse arguments
USER_INSTALL=false
CUSTOM_VERSION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --user)
      USER_INSTALL=true
      shift
      ;;
    --help|-h)
      show_help
      ;;
    --version)
      echo "Coolify CLI Installer v1.0.0"
      exit 0
      ;;
    *)
      CUSTOM_VERSION="$1"
      shift
      ;;
  esac
done

# Check required tools
check_requirements() {
  local missing_tools=()

  if ! command -v curl &> /dev/null; then
    missing_tools+=("curl")
  fi

  if ! command -v tar &> /dev/null; then
    missing_tools+=("tar")
  fi

  if [ ${#missing_tools[@]} -gt 0 ]; then
    error_exit "Missing required tools: ${missing_tools[*]}\nPlease install them and try again."
  fi
}

# Function to detect platform, architecture, etc.
detect_platform() {
  OS=$(uname -s)
  ARCH=$(uname -m)

  case $OS in
    Linux) OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)
      error_exit "Unsupported operating system: $OS\nSupported: Linux, macOS"
      ;;
  esac

  case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *)
      error_exit "Unsupported architecture: $ARCH\nSupported: x86_64/amd64, aarch64/arm64"
      ;;
  esac
}

# Fetch latest release version from GitHub
get_latest_version() {
  echo "Fetching latest release version..." >&2
  local latest_version
  latest_version=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$latest_version" ]; then
    error_exit "Failed to fetch latest release version from GitHub"
  fi

  echo "$latest_version"
}

# Validate version format (should start with v or be a semantic version)
validate_version() {
  local version=$1
  # Check if version starts with 'v' or is a plain semantic version
  if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
    error_exit "Invalid version format: $version\nExpected format: v1.0.0 or 1.0.0"
  fi

  # Ensure version starts with 'v' for GitHub releases
  if [[ ! "$version" =~ ^v ]]; then
    version="v${version}"
  fi

  echo "$version"
}

# Check if coolify is already installed
check_existing_installation() {
  if command -v coolify &> /dev/null; then
    local current_version
    current_version=$(coolify version 2>/dev/null | head -n1 || echo "unknown")
    echo -e "${YELLOW}Coolify CLI is already installed: ${current_version}${NC}"
    echo -e "This will upgrade/reinstall to version ${GREEN}${1}${NC}"
    read -p "Continue? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Installation cancelled."
      exit 0
    fi
  fi
}

# Download and verify file from GitHub release
download_from_github() {
  local repo=$1
  local release=$2
  local name=$3
  local install_dir=$4

  local filename="${name}_${release#v}_${OS}_${ARCH}.tar.gz"
  local download_url="https://github.com/${repo}/releases/download/${release}/${filename}"

  echo -e "${GREEN}Downloading ${name} ${release}${NC}"
  echo "Platform: ${OS}/${ARCH}"
  echo "URL: ${download_url}"

  # Create temp file
  TEMP_FILE=$(mktemp)

  # Download with progress and error handling
  if ! curl -fSL --progress-bar -o "${TEMP_FILE}" "${download_url}"; then
    error_exit "Failed to download from ${download_url}\nPlease check if the version exists or try again later."
  fi

  # Verify downloaded file is not empty
  if [ ! -s "$TEMP_FILE" ]; then
    error_exit "Downloaded file is empty"
  fi

  # Create install directory if it doesn't exist (for user install)
  if [ "$USER_INSTALL" = true ] && [ ! -d "$install_dir" ]; then
    echo "Creating directory: ${install_dir}"
    mkdir -p "$install_dir" || error_exit "Failed to create directory ${install_dir}"
  fi

  # Extract binary
  echo "Installing ${name} to ${install_dir}/${BINARY_NAME}"

  if [ "$USER_INSTALL" = true ]; then
    if ! tar -xzf "${TEMP_FILE}" -C "${install_dir}"; then
      error_exit "Failed to extract binary"
    fi
    chmod +x "${install_dir}/${BINARY_NAME}" || error_exit "Failed to make binary executable"
  else
    if ! sudo tar -xzf "${TEMP_FILE}" -C "${install_dir}"; then
      error_exit "Failed to extract binary (sudo required)"
    fi
    if ! sudo chmod +x "${install_dir}/${BINARY_NAME}"; then
      error_exit "Failed to make binary executable (sudo required)"
    fi
  fi

  # Verify installation
  if [ ! -f "${install_dir}/${BINARY_NAME}" ]; then
    error_exit "Binary was not installed to ${install_dir}/${BINARY_NAME}"
  fi

  echo -e "${GREEN}✓ ${name} installed successfully to ${install_dir}/${BINARY_NAME}${NC}"

  # Check if install directory is in PATH
  if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
    echo -e "${YELLOW}Warning: ${install_dir} is not in your PATH${NC}"
    if [ "$USER_INSTALL" = true ]; then
      echo "Add it to your PATH by adding this line to your ~/.bashrc or ~/.zshrc:"
      echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
  fi

  # Show installed version
  local installed_version
  if installed_version=$("${install_dir}/${BINARY_NAME}" version 2>/dev/null | head -n1); then
    echo -e "Installed version: ${GREEN}${installed_version}${NC}"
  fi
}

# Main installation flow
main() {
  echo "Coolify CLI Installer"
  echo "===================="
  echo

  # Check requirements first
  check_requirements

  # Detect platform
  detect_platform

  # Determine version to install
  local version_to_install
  if [ -z "$CUSTOM_VERSION" ]; then
    version_to_install=$(get_latest_version)
  else
    version_to_install=$(validate_version "$CUSTOM_VERSION")
  fi

  echo "Version to install: ${version_to_install}"
  echo

  # Check existing installation
  check_existing_installation "$version_to_install"

  # Determine install directory
  local install_dir
  if [ "$USER_INSTALL" = true ]; then
    install_dir="$USER_INSTALL_DIR"
    echo "Install mode: User (no sudo required)"
  else
    install_dir="$GLOBAL_INSTALL_DIR"
    echo "Install mode: Global (requires sudo)"
  fi

  echo "Install directory: ${install_dir}"
  echo

  # Download and install
  download_from_github "$REPO" "$version_to_install" "coolify-cli" "$install_dir"

  echo
  echo -e "${GREEN}Installation complete!${NC}"
  echo "Run 'coolify --help' to get started"
}

# Run main installation
main
