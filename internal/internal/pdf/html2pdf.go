package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Preferred engines in order (wkhtmltopdf → Chromium)
var candidates = [][]string{
	{"wkhtmltopdf", "--print-media-type", "-", "-"},                               // stdin → stdout
	{"chromium", "--headless", "--disable-gpu", "--print-to-pdf=/dev/stdout", ""}, // needs file:// path
	{"google-chrome", "--headless", "--disable-gpu", "--print-to-pdf=/dev/stdout", ""},
	{"chromium-browser", "--headless", "--disable-gpu", "--print-to-pdf=/dev/stdout", ""},
}

type Engine string

const (
	EngineWkhtml Engine = "wkhtmltopdf"
	EngineChrome Engine = "chromium"
)

// RenderHTMLToPDF tries wkhtmltopdf first; if not found, falls back to Chromium.
// For wkhtmltopdf we stream HTML via stdin. For Chromium we must write to a temp HTML file.
func RenderHTMLToPDF(html []byte) ([]byte, error) {
	// Validate HTML size (max 10MB to prevent DoS)
	const maxHTMLSize = 10 * 1024 * 1024 // 10MB
	if len(html) > maxHTMLSize {
		return nil, fmt.Errorf("HTML size exceeds maximum allowed size of 10MB")
	}

	// Set timeout for PDF generation (30 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		switch c[0] {
		case "wkhtmltopdf":
			cmd := exec.CommandContext(ctx, c[0], c[1:]...)
			cmd.Stdin = bytes.NewReader(html)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return nil, fmt.Errorf("PDF generation timed out")
				}
				return nil, err
			}
			return output, nil
		default:
			// Chromium needs a file URL - use secure random temp file
			tmpFile, err := os.CreateTemp("", "render-*.html")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file: %w", err)
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath) // Clean up temp file

			// Write HTML with secure permissions
			if err := os.WriteFile(tmpPath, html, 0600); err != nil {
				tmpFile.Close()
				return nil, fmt.Errorf("failed to write temp file: %w", err)
			}
			tmpFile.Close()

			args := []string{c[1], c[2], c[3], "file://" + tmpPath}
			cmd := exec.CommandContext(ctx, c[0], args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return nil, fmt.Errorf("PDF generation timed out")
				}
				return nil, err
			}
			return output, nil
		}
	}
	return nil, fmt.Errorf("no PDF engine found (install wkhtmltopdf or chromium)")
}

// small wrappers (keeps stdlib imports local)
func osTempDir() string                    { return filepath.Clean("/tmp") }
func osWriteFile(p string, b []byte) error { return os.WriteFile(p, b, 0o644) }
