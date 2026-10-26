package genreclassifier

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

// Photo carries the untouched file bytes: the model resizes internally, and
// re-encoding client-side would only add a quality loss nobody asked for.
type Photo struct {
	Path   string
	Bytes  []byte
	MIME   string
	Width  int
	Height int
}

func OpenPhoto(path string) (Photo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		abs, aerr := filepath.Abs(path)
		if aerr != nil {
			abs = path
		}
		return Photo{}, InputErrorf("photo %s: %v", abs, err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return Photo{}, InputErrorf("photo %s: not a decodable JPEG or PNG (%v)", path, err)
	}
	if err := checkDimensions(path, cfg.Width, cfg.Height); err != nil {
		return Photo{}, err
	}
	return Photo{
		Path:   path,
		Bytes:  raw,
		MIME:   mimeFor(format),
		Width:  cfg.Width,
		Height: cfg.Height,
	}, nil
}

func checkDimensions(path string, width, height int) error {
	if width <= 0 || height <= 0 {
		return InputErrorf("photo %s: zero-sized image %dx%d", path, width, height)
	}
	return nil
}

// mimeFor only ever sees the formats registered above, so a non-empty format
// is always jpeg or png. DecodeConfig rejects anything else.
func mimeFor(format string) string {
	if format == "png" {
		return "image/png"
	}
	return "image/jpeg"
}
