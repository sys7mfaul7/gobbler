#!/usr/bin/env bash
# Set up every toolchain used by gobbler and its satellites.
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
for s in go dotnet nextcloud android; do
  echo "== setup-$s"; "$here/setup-$s.sh"
done
