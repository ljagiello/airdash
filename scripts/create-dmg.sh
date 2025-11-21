#!/bin/bash
set -e

APP_PATH="$1"
VERSION="$2"
DMG_NAME="AirDash-${VERSION}-arm64.dmg"

if [ -z "$APP_PATH" ] || [ -z "$VERSION" ]; then
    echo "Usage: $0 <app-path> <version>"
    echo "Example: $0 dist/AirDash.app 0.0.2"
    exit 1
fi

if [ ! -d "$APP_PATH" ]; then
    echo "Error: App bundle not found at $APP_PATH"
    exit 1
fi

DIST_DIR=$(dirname "$APP_PATH")
DMG_PATH="$DIST_DIR/$DMG_NAME"
DMG_TMP_DIR="$DIST_DIR/dmg_tmp"

echo "Creating DMG: $DMG_NAME"

# Create temporary directory for DMG contents
rm -rf "$DMG_TMP_DIR"
mkdir -p "$DMG_TMP_DIR"

# Copy app bundle to temp directory
echo "Copying app bundle..."
cp -R "$APP_PATH" "$DMG_TMP_DIR/"

# Create Applications symlink
echo "Creating Applications symlink..."
ln -s /Applications "$DMG_TMP_DIR/Applications"

# Create DMG using hdiutil
echo "Creating DMG..."
hdiutil create -volname "AirDash ${VERSION}" \
    -srcfolder "$DMG_TMP_DIR" \
    -ov \
    -format UDZO \
    "$DMG_PATH"

# Clean up
rm -rf "$DMG_TMP_DIR"

echo "DMG created successfully: $DMG_PATH"
