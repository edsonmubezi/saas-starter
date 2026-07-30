package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// textExtensions is the set of file extensions whose contents should have
// string replacements applied. Files with other extensions are left untouched.
var textExtensions = map[string]bool{
	".go":         true,
	".mod":        true,
	".toml":       true,
	".yaml":       true,
	".yml":        true,
	".json":       true,
	".env":        true,
	".md":         true,
	".tsx":        true,
	".ts":         true,
	".js":         true,
	".jsx":        true,
	".css":        true,
	".html":       true,
	".sql":        true,
	".sh":         true,
	".dockerfile": true,
}

// textBasenames are exact filenames (no extension) that are also treated as
// text files regardless of their extension (or lack thereof).
var textBasenames = map[string]bool{
	"Dockerfile":    true,
	"Makefile":      true,
	".env.example":  true, // filepath.Ext(".env.example") == ".example", not ".env"
}

// WalkAndReplace walks root and applies every replacement pair to each text
// file it finds. Directories whose base name appears in skipDirs are skipped
// entirely. Files that do not change are not rewritten.
func WalkAndReplace(root string, pairs []Pair, skipDirs []string) error {
	skipSet := make(map[string]bool, len(skipDirs))
	for _, d := range skipDirs {
		skipSet[d] = true
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipSet[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if !isTextFile(path) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)
		modified := content
		for _, p := range pairs {
			modified = strings.ReplaceAll(modified, p.Old, p.New)
		}
		if modified == content {
			return nil
		}

		return os.WriteFile(path, []byte(modified), 0644)
	})
}

func isTextFile(path string) bool {
	base := filepath.Base(path)
	if textBasenames[base] {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	return textExtensions[ext]
}
