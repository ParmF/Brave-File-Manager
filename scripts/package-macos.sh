#!/usr/bin/env bash
set -euo pipefail

# Builds a macOS .app bundle and DMG without the Fyne CLI (avoids Go 1.26 tool conflicts).
#
# Usage: package-macos.sh <goarch> <app-version> <output-dmg-path>

GOARCH="${1:?goarch required (amd64 or arm64)}"
APP_VERSION="${2:?app version required}"
OUTPUT_DMG="${3:?output dmg path required}"

APP_NAME="Brave File Manager"
EXEC_NAME="BraveFileManager"
BUNDLE_ID="com.parmf.bravefilemanager"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="${ROOT}/${APP_NAME}.app"
DIST_DIR="$(dirname "${OUTPUT_DMG}")"

export CGO_ENABLED=1
export GOOS=darwin
export GOARCH="${GOARCH}"

rm -rf "${APP_DIR}"
mkdir -p "${DIST_DIR}"
mkdir -p "${APP_DIR}/Contents/MacOS"
mkdir -p "${APP_DIR}/Contents/Resources"

echo "Building ${APP_NAME} for darwin/${GOARCH}..."
(
  cd "${ROOT}"
  go build -trimpath -ldflags="-s -w" -o "${APP_DIR}/Contents/MacOS/${EXEC_NAME}" .
)

cat > "${APP_DIR}/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleDisplayName</key>
  <string>${APP_NAME}</string>
  <key>CFBundleExecutable</key>
  <string>${EXEC_NAME}</string>
  <key>CFBundleIdentifier</key>
  <string>${BUNDLE_ID}</string>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>${APP_VERSION}</string>
  <key>CFBundleVersion</key>
  <string>${APP_VERSION}</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
EOF

rm -f "${OUTPUT_DMG}"
echo "Creating DMG at ${OUTPUT_DMG}..."
hdiutil create \
  -volname "${APP_NAME}" \
  -srcfolder "${APP_DIR}" \
  -ov -format UDZO \
  "${OUTPUT_DMG}"

rm -rf "${APP_DIR}"
echo "Done: ${OUTPUT_DMG}"
