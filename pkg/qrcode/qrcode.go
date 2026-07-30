package qrcode

import (
	goqrcode "github.com/skip2/go-qrcode"
)

// GeneratePNG generates a QR code as PNG bytes for the given content string.
// size controls the width/height in pixels.
func GeneratePNG(content string, size int) ([]byte, error) {
	return goqrcode.Encode(content, goqrcode.Medium, size)
}
