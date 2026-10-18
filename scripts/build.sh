#!/usr/bin/env bash
# Builds every executable into dist/ (spec NFR-6, AC-4):
#   dist/anscombe        command-line program for this computer
#   dist/anscombe.exe    Windows program (64-bit Intel/AMD), cross-compiled
#   dist/Anscombe.app    macOS application (only when run on a Mac)
#
# Go cross-compiles by setting GOOS and GOARCH, so no Windows machine is needed.
# The data is embedded, so each executable runs on its own with no other files.
#
# Usage, from any directory:
#   scripts/build.sh

set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p dist

go build -o dist/anscombe ./cmd/anscombe
echo "Built dist/anscombe"

GOOS=windows GOARCH=amd64 go build -o dist/anscombe.exe ./cmd/anscombe
echo "Built dist/anscombe.exe"

if [[ "$(uname)" == "Darwin" ]]; then
  scripts/build-macos-app.sh
else
  echo "Skipped Anscombe.app (macOS only)"
fi
