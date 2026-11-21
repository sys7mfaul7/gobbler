#!/usr/bin/env bash
# Go toolchain for the gobbler core and the AIMP agent.
# Cross-compiling needs nothing extra: builds are CGO_ENABLED=0 static binaries.
set -euo pipefail
GO_VERSION="${GO_VERSION:-1.26.0}"   # minimum; matches go.mod
if command -v go >/dev/null; then
  echo "go already installed: $(go version)"; exit 0
fi
os=$(uname -s | tr '[:upper:]' '[:lower:]'); arch=$(uname -m)
case "$arch" in x86_64) arch=amd64;; aarch64|arm64) arch=arm64;; esac
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.${os}-${arch}.tar.gz" | sudo tar -C /usr/local -xz
echo 'add /usr/local/go/bin to PATH'
