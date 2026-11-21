#!/usr/bin/env bash
# Android SDK/NDK + gomobile for the future Android agent
# (AIMP for Android + Nextcloud Android app as the sync path).
set -euo pipefail
ANDROID_HOME="${ANDROID_HOME:-/opt/android-sdk}"
CLT_BUILD="${CLT_BUILD:-16111833}"              # commandlinetools-linux-<build>_latest.zip
NDK_VERSION="${NDK_VERSION:-29.0.14206865}"
PLATFORM="${PLATFORM:-android-36}"
BUILD_TOOLS="${BUILD_TOOLS:-36.1.0}"
command -v java >/dev/null || sudo apt-get install -y -qq openjdk-21-jdk-headless unzip
sudo mkdir -p "$ANDROID_HOME" && sudo chown -R "$(id -u):$(id -g)" "$ANDROID_HOME"
if [ ! -x "$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager" ]; then
  curl -fsSL -o /tmp/clt.zip "https://dl.google.com/android/repository/commandlinetools-linux-${CLT_BUILD}_latest.zip"
  mkdir -p "$ANDROID_HOME/cmdline-tools" && unzip -q /tmp/clt.zip -d "$ANDROID_HOME/cmdline-tools"
  mv "$ANDROID_HOME/cmdline-tools/cmdline-tools" "$ANDROID_HOME/cmdline-tools/latest"
fi
SM="$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager"
yes | "$SM" --licenses >/dev/null || true
"$SM" "platform-tools" "platforms;${PLATFORM}" "build-tools;${BUILD_TOOLS}" "ndk;${NDK_VERSION}"
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/$NDK_VERSION"
"$(go env GOPATH)/bin/gomobile" init
echo "export ANDROID_HOME=$ANDROID_HOME ANDROID_NDK_HOME=$ANDROID_NDK_HOME"
