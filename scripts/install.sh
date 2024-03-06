#!/bin/bash

# Function to detect platform, architecture, etc.
detect_platform() {
  OS=$(uname -s)
  ARCH=$(uname -m)

  case $OS in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
  esac

  case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
  esac
}

# Function to download file from GitHub release
download_from_github() {
  local repo=$1
  local release=$2
  local name=$3
  local filename=${name}_${release}_${OS}_${ARCH}.tar.gz
  # https://github.com/coollabsio/coolify-cli/releases/download/0.0.1/coolify-cli_0.0.1_linux_amd64.tar.gz
  # Construct download URL
  local download_url="https://github.com/${repo}/releases/download/${release}/${filename}"

  # Use curl to download the file
  curl -L -o "${filename}" $download_url

  # Determine the binary directory
  local binary_dir=""
  if [ "$OS" == "linux" ] || [ "$OS" == "darwin" ]; then
    binary_dir="/usr/local/bin"
  fi

  sudo tar -xzvf "${filename}" -C "${binary_dir}"

  # Cleanup
  rm "${filename}"

  echo "${name} installed successfully to ${binary_dir}"
}

detect_platform
download_from_github "coollabsio/coolify-cli" "0.0.1" "coolify-cli"