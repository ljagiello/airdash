#!/bin/bash
set -e

SVG_FILE="$1"
OUTPUT_ICNS="$2"

if [ -z "$SVG_FILE" ] || [ -z "$OUTPUT_ICNS" ]; then
    echo "Usage: $0 <input.svg> <output.icns>"
    exit 1
fi

if [ ! -f "$SVG_FILE" ]; then
    echo "Error: SVG file not found at $SVG_FILE"
    exit 1
fi

# Create temporary directory for iconset
ICONSET_DIR=$(mktemp -d)/AppIcon.iconset
mkdir -p "$ICONSET_DIR"

echo "Generating icon from $SVG_FILE..."

# Required icon sizes for macOS
# Format: icon_SIZExSIZE[.scale].png where scale is optional (@2x)
SIZES=(
    "16:icon_16x16.png"
    "32:icon_16x16@2x.png"
    "32:icon_32x32.png"
    "64:icon_32x32@2x.png"
    "128:icon_128x128.png"
    "256:icon_128x128@2x.png"
    "256:icon_256x256.png"
    "512:icon_256x256@2x.png"
    "512:icon_512x512.png"
    "1024:icon_512x512@2x.png"
)

# Check if we have rsvg-convert (from librsvg)
if command -v rsvg-convert &> /dev/null; then
    echo "Using rsvg-convert to generate PNG files..."
    for size_entry in "${SIZES[@]}"; do
        IFS=: read -r size filename <<< "$size_entry"
        echo "  Generating ${filename} (${size}x${size})..."
        rsvg-convert -w "$size" -h "$size" "$SVG_FILE" -o "$ICONSET_DIR/$filename"
    done
elif command -v qlmanage &> /dev/null && command -v sips &> /dev/null; then
    echo "Using qlmanage + sips to generate PNG files..."

    # First, convert SVG to a high-res PNG using qlmanage
    TEMP_PNG=$(mktemp).png
    qlmanage -t -s 1024 -o "$(dirname $TEMP_PNG)" "$SVG_FILE" > /dev/null 2>&1
    mv "$(dirname $TEMP_PNG)/$(basename $SVG_FILE .svg).png" "$TEMP_PNG" 2>/dev/null || {
        # qlmanage might use different naming
        mv "$(dirname $TEMP_PNG)/"*.png "$TEMP_PNG" 2>/dev/null || {
            echo "Error: Failed to convert SVG to PNG"
            rm -rf "$ICONSET_DIR"
            exit 1
        }
    }

    # Resize to all required sizes
    for size_entry in "${SIZES[@]}"; do
        IFS=: read -r size filename <<< "$size_entry"
        echo "  Generating ${filename} (${size}x${size})..."
        sips -z "$size" "$size" "$TEMP_PNG" --out "$ICONSET_DIR/$filename" > /dev/null 2>&1
    done

    rm -f "$TEMP_PNG"
else
    echo "Error: Neither rsvg-convert nor qlmanage/sips are available"
    echo "Please install librsvg: brew install librsvg"
    rm -rf "$ICONSET_DIR"
    exit 1
fi

# Convert iconset to icns
echo "Creating .icns file..."
iconutil -c icns "$ICONSET_DIR" -o "$OUTPUT_ICNS"

# Clean up
rm -rf "$(dirname $ICONSET_DIR)"

echo "Icon created successfully: $OUTPUT_ICNS"
