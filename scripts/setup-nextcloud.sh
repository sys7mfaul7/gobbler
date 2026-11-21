#!/usr/bin/env bash
# PHP + Composer + Node for playlist-gobbler.nextcloud (Debian/Ubuntu).
set -euo pipefail
sudo apt-get update -qq
sudo apt-get install -y -qq php-cli php-xml php-mbstring php-zip php-curl php-intl php-sqlite3 composer libxml2-utils
if ! command -v node >/dev/null; then
  echo "Install Node >= 20 (e.g. from https://nodejs.org or your distro)"; exit 1
fi
php --version | head -1; composer --version; node --version
