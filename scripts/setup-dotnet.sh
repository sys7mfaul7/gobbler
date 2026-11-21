#!/usr/bin/env bash
# .NET SDK for playlist-gobbler.jellyfin.
# Jellyfin 12.4's plugin packages (Jellyfin.Controller 12.4.0) target net10.0.
set -euo pipefail
CHANNEL="${DOTNET_CHANNEL:-10.0}"
DEST="${DOTNET_ROOT:-/opt/dotnet}"
if command -v dotnet >/dev/null && dotnet --list-sdks | grep -q "^${CHANNEL}"; then
  echo "dotnet ${CHANNEL} SDK already installed"; exit 0
fi
curl -fsSL https://dot.net/v1/dotnet-install.sh -o /tmp/dotnet-install.sh
sudo bash /tmp/dotnet-install.sh --channel "$CHANNEL" --install-dir "$DEST"
sudo ln -sf "$DEST/dotnet" /usr/local/bin/dotnet
dotnet --list-sdks
