#!/usr/bin/env bash
# Builds dist/Anscombe.app, a macOS application bundle for the rubric's ".app"
# requirement (decision Q5).
#
# The program is a command-line tool, and a double-clicked app has no terminal
# to show its output. So the bundle is a thin wrapper: its launcher opens
# Terminal, runs the program there, and waits for Return so the results stay on
# screen. The program itself is unchanged.
#
# The binary is "universal" (runs on both Apple Silicon and Intel Macs) when
# Apple's lipo tool is available; otherwise it is built for this Mac only.
#
# Usage, from any directory:
#   scripts/build-macos-app.sh            then double-click dist/Anscombe.app

set -euo pipefail
cd "$(dirname "$0")/.."

app=dist/Anscombe.app
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"

# 1. The program, as a universal binary where possible.
if command -v lipo >/dev/null; then
  GOOS=darwin GOARCH=arm64 go build -o dist/anscombe-arm64 ./cmd/anscombe
  GOOS=darwin GOARCH=amd64 go build -o dist/anscombe-amd64 ./cmd/anscombe
  lipo -create -output "$app/Contents/Resources/anscombe" dist/anscombe-arm64 dist/anscombe-amd64
  rm dist/anscombe-arm64 dist/anscombe-amd64
  # Re-sign ad hoc: Apple Silicon refuses to run binaries without a signature.
  codesign --force --sign - "$app/Contents/Resources/anscombe" 2>/dev/null || true
else
  echo "build-macos-app.sh: lipo not found; building for this Mac's processor only" >&2
  GOOS=darwin go build -o "$app/Contents/Resources/anscombe" ./cmd/anscombe
fi

# 2. A script Terminal runs: show the results, then wait so the window stays readable.
cat >"$app/Contents/Resources/run-in-terminal.command" <<'EOF'
#!/bin/bash
clear
"$(dirname "$0")/anscombe"
echo
read -r -p "Press Return to close this window. " _
EOF

# 3. The app's launcher: hand the script to Terminal.
cat >"$app/Contents/MacOS/Anscombe" <<'EOF'
#!/bin/bash
open -a Terminal "$(dirname "$0")/../Resources/run-in-terminal.command"
EOF
chmod +x "$app/Contents/MacOS/Anscombe" "$app/Contents/Resources/run-in-terminal.command"

# 4. The metadata macOS needs to treat the folder as an application.
cat >"$app/Contents/Info.plist" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>               <string>Anscombe</string>
  <key>CFBundleDisplayName</key>        <string>Anscombe</string>
  <key>CFBundleIdentifier</key>         <string>com.github.mioluigi.anscombe</string>
  <key>CFBundleExecutable</key>         <string>Anscombe</string>
  <key>CFBundlePackageType</key>        <string>APPL</string>
  <key>CFBundleShortVersionString</key> <string>1.0</string>
  <key>CFBundleVersion</key>            <string>1</string>
  <key>LSMinimumSystemVersion</key>     <string>11.0</string>
</dict>
</plist>
EOF

echo "Built $app — double-click it in Finder, or run: open $app"
