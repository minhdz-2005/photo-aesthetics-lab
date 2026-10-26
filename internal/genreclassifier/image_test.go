package genreclassifier

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeJPEG(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 120, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "photo.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenPhotoReadsJPEG(t *testing.T) {
	path := writeJPEG(t, 40, 30)
	photo, err := OpenPhoto(path)
	if err != nil {
		t.Fatalf("OpenPhoto: %v", err)
	}
	if photo.Width != 40 || photo.Height != 30 {
		t.Errorf("size = %dx%d, want 40x30", photo.Width, photo.Height)
	}
	if photo.MIME != "image/jpeg" {
		t.Errorf("MIME = %q, want image/jpeg", photo.MIME)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(photo.Bytes, raw) {
		t.Error("photo.Bytes must be the untouched file content")
	}
}

func TestOpenPhotoRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	textNamedJPG := filepath.Join(dir, "not-a-photo.jpg")
	if err := os.WriteFile(textNamedJPG, []byte("this is plain text, not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "empty.jpg")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		path    string
		wantMsg string
	}{
		{"missing file", filepath.Join(dir, "nope.jpg"), "photo"},
		{"empty file", empty, "not a decodable JPEG or PNG"},
		{"text renamed to jpg", textNamedJPG, "not a decodable JPEG or PNG"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := OpenPhoto(c.path)
			if err == nil {
				t.Fatalf("OpenPhoto(%s) = nil error, want failure", c.path)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isInput(err) {
				t.Errorf("error %v must be InputError so exit code is 1", err)
			}
		})
	}
}

func TestCheckDimensions(t *testing.T) {
	if err := checkDimensions("a.jpg", 0, 12); err == nil {
		t.Error("width 0 must be rejected")
	}
	if err := checkDimensions("a.jpg", 12, 0); err == nil {
		t.Error("height 0 must be rejected")
	}
	if err := checkDimensions("a.jpg", 12, 8); err != nil {
		t.Errorf("12x8 must pass, got %v", err)
	}
}
