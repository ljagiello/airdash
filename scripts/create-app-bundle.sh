#!/bin/bash
set -e

BINARY_PATH=$1
VERSION=$2

# Get the directory containing the binary
DIST_DIR=$(dirname "$BINARY_PATH")

# Create app bundle structure
APP_BUNDLE="$DIST_DIR/AirDash.app"
CONTENTS="$APP_BUNDLE/Contents"
MACOS="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"

echo "Creating app bundle at $APP_BUNDLE"

# Create directories
mkdir -p "$MACOS"
mkdir -p "$RESOURCES"

# Copy binary
echo "Copying binary to $MACOS/airdash"
cp "$BINARY_PATH" "$MACOS/airdash"
chmod +x "$MACOS/airdash"

# Create Info.plist
echo "Creating Info.plist"
cat > "$CONTENTS/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>AirDash</string>
	<key>CFBundleDisplayName</key>
	<string>AirDash</string>
	<key>CFBundleIdentifier</key>
	<string>com.github.ljagiello.airdash</string>
	<key>CFBundleVersion</key>
	<string>$VERSION</string>
	<key>CFBundleShortVersionString</key>
	<string>$VERSION</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleExecutable</key>
	<string>airdash</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHumanReadableCopyright</key>
	<string>Copyright © 2023-2025 Łukasz Jagiełło</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
EOF

# Copy logo if it exists (optional)
if [ -f "assets/app/logo.png" ]; then
	echo "Copying logo to Resources"
	cp "assets/app/logo.png" "$RESOURCES/"
fi

echo "✓ App bundle created successfully"
echo "  Bundle: $APP_BUNDLE"
echo "  Binary: $MACOS/airdash"
echo "  Info.plist: $CONTENTS/Info.plist"
