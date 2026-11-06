package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

// TestRenderPreviewImages is a manual visual-check harness: it renders the
// status bar title in every icon style to PNG files. Run with:
//
//	AIRDASH_RENDER_DIR=/tmp/airdash-preview go test -run TestRenderPreviewImages
func TestRenderPreviewImages(t *testing.T) {
	dir := os.Getenv("AIRDASH_RENDER_DIR")
	if dir == "" {
		t.Skip("set AIRDASH_RENDER_DIR to render preview PNGs")
	}
	runtime.LockOSThread()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	measures := AirGradientMeasures{Atmp: 24.5, Pm02: 4, Rhum: 66, Rco2: 679}
	for _, style := range []IconStyle{IconStyleSolid, IconStyleHairline, IconStyleLabels} {
		s := defaultDisplaySettings()
		s.IconStyle = style
		s.Visible[MetricHeatIndex] = true
		title := buildStatusTitle(measures, "F", s, true)

		size := attributedSize(title)
		width := size.Width + 24
		height := 24.0
		img := appkit.NewImageWithSize(foundation.Size{Width: width, Height: height})
		objc.Call[objc.Void](img, objc.Sel("lockFocus"))
		appkit.Color_WindowBackgroundColor().SetFill()
		appkit.BezierPath_BezierPathWithRect(frame(0, 0, width, height)).Fill()
		objc.Call[objc.Void](title, objc.Sel("drawAtPoint:"), foundation.Point{X: 12, Y: (height - size.Height) / 2})
		objc.Call[objc.Void](img, objc.Sel("unlockFocus"))

		rep := appkit.BitmapImageRep_ImageRepWithData(img.TIFFRepresentation())
		png := rep.RepresentationUsingTypeProperties(appkit.BitmapImageFileTypePNG, nil)
		out := filepath.Join(dir, "preview-"+string(style)+".png")
		if err := os.WriteFile(out, png, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", out)
	}
}
