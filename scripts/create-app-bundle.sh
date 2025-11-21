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
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
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

# Generate and copy app icon
if [ -f "assets/app/logo.svg" ]; then
	echo "Generating app icon from logo.svg"
	ICON_FILE="$RESOURCES/AppIcon.icns"
	./scripts/generate-icon.sh "assets/app/logo.svg" "$ICON_FILE"
fi

# Sign the app bundle if certificate is available
if [ -n "$MACOS_SIGN_IDENTITY" ]; then
	echo "Signing app bundle with identity: $MACOS_SIGN_IDENTITY"
	codesign --force --deep --sign "$MACOS_SIGN_IDENTITY" \
		--options runtime \
		--timestamp \
		"$APP_BUNDLE"

	# Verify signature
	echo "Verifying signature..."
	codesign --verify --deep --verbose=2 "$APP_BUNDLE"

	# Notarize if credentials are available
	if [ -n "$MACOS_NOTARY_ISSUER_ID" ] && [ -n "$MACOS_NOTARY_KEY_ID" ] && [ -n "$MACOS_NOTARY_KEY_BASE64" ]; then
		echo "Notarizing app bundle..."

		# Create temporary directory for notarization key
		TEMP_DIR=$(mktemp -d)
		KEY_FILE="$TEMP_DIR/AuthKey_${MACOS_NOTARY_KEY_ID}.p8"

		# Decode and write the key
		echo "$MACOS_NOTARY_KEY_BASE64" | base64 --decode > "$KEY_FILE"

		# Create a ZIP for notarization (notarytool requires ZIP or DMG)
		NOTARY_ZIP="$DIST_DIR/AirDash-notarize.zip"
		echo "Creating ZIP for notarization..."
		ditto -c -k --keepParent "$APP_BUNDLE" "$NOTARY_ZIP"

		# Submit for notarization
		echo "Submitting to Apple notary service..."
		xcrun notarytool submit "$NOTARY_ZIP" \
			--key "$KEY_FILE" \
			--key-id "$MACOS_NOTARY_KEY_ID" \
			--issuer "$MACOS_NOTARY_ISSUER_ID" \
			--wait

		# Check notarization status
		if [ $? -eq 0 ]; then
			echo "Notarization successful! Stapling ticket to app bundle..."
			xcrun stapler staple "$APP_BUNDLE"
			echo "Notarization complete"
		else
			echo "WARNING: Notarization failed - app may show security warnings"
		fi

		# Cleanup
		rm -f "$NOTARY_ZIP"
		rm -rf "$TEMP_DIR"
	else
		echo "WARNING: Notarization credentials not provided - app will show security warnings"
		echo "         Set MACOS_NOTARY_ISSUER_ID, MACOS_NOTARY_KEY_ID, and MACOS_NOTARY_KEY_BASE64"
	fi
else
	echo "WARNING: No signing identity provided - app will not be signed"
	echo "         Set MACOS_SIGN_IDENTITY environment variable to sign the app"
fi

echo "App bundle created successfully"
echo "  Bundle: $APP_BUNDLE"
echo "  Binary: $MACOS/airdash"
echo "  Info.plist: $CONTENTS/Info.plist"
