package export

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jung-kurt/gofpdf"
)

// fontsDir is the path to the directory containing TTF font files.
var fontsDir = "assets/fonts"

// builtinFonts are always available in gofpdf without TTF files.
var builtinFonts = []string{"Arial", "Helvetica", "Times", "Courier"}

var (
	cachedFonts []string
	fontsOnce   sync.Once
)

// SetFontsDir overrides the default fonts directory (useful for tests or custom paths).
func SetFontsDir(dir string) {
	fontsDir = dir
}

// discoverFonts scans the fonts directory for TTF files.
func discoverFonts() []string {
	seen := make(map[string]bool)
	for _, f := range builtinFonts {
		seen[f] = true
	}

	absDir, err := filepath.Abs(fontsDir)
	if err != nil {
		log.Printf("fonts: could not resolve fonts dir %q: %v", fontsDir, err)
		result := append([]string{}, builtinFonts...)
		sort.Strings(result)
		return result
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		log.Printf("fonts: could not read fonts dir %q: %v", absDir, err)
		result := append([]string{}, builtinFonts...)
		sort.Strings(result)
		return result
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".ttf") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		name = strings.ReplaceAll(name, "_", " ")
		seen[name] = true
	}

	var result []string
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	log.Printf("fonts: %d fonts available (%d built-in + %d custom TTF) from %s", len(result), len(builtinFonts), len(result)-len(builtinFonts), absDir)
	return result
}

// AvailableFonts returns a sorted list of all font names that can be used in PDFs.
// This includes the 4 gofpdf built-in fonts plus any TTF files found in the fonts directory.
// Results are cached after first call; restart server to pick up new fonts.
func AvailableFonts() []string {
	fontsOnce.Do(func() {
		cachedFonts = discoverFonts()
	})
	return cachedFonts
}

// IsBuiltinFont returns true if the font is one of the 4 gofpdf built-in fonts.
func IsBuiltinFont(name string) bool {
	for _, b := range builtinFonts {
		if strings.EqualFold(b, name) {
			return true
		}
	}
	return false
}

// ttfPath returns the absolute TTF file path for a custom font name.
func ttfPath(fontName string) string {
	filename := strings.ReplaceAll(fontName, " ", "_") + ".ttf"
	absDir, err := filepath.Abs(fontsDir)
	if err != nil {
		return filepath.Join(fontsDir, filename)
	}
	return filepath.Join(absDir, filename)
}

// RegisterFont registers a custom TTF font with the given gofpdf instance.
// For built-in fonts this is a no-op. Returns the font name to use with SetFont.
func RegisterFont(pdf *gofpdf.Fpdf, fontName string) string {
	if fontName == "" {
		return "Arial"
	}

	// Built-in fonts don't need registration
	if IsBuiltinFont(fontName) {
		return fontName
	}

	path := ttfPath(fontName)
	if _, err := os.Stat(path); err != nil {
		log.Printf("fonts: TTF file not found for %q at %s, falling back to Arial", fontName, path)
		return "Arial"
	}

	// Use the font name as the gofpdf family identifier
	pdf.AddUTF8Font(fontName, "", path)
	pdf.AddUTF8Font(fontName, "B", path)  // Bold uses same file (variable fonts)
	pdf.AddUTF8Font(fontName, "I", path)  // Italic uses same file
	pdf.AddUTF8Font(fontName, "BI", path) // Bold-italic uses same file

	return fontName
}
