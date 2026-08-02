// Package scaffold orchestrates cloning the saas-starter template, applying
// string replacements, and running post-processing steps.
package scaffold

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Options holds every value needed to scaffold a new project.
type Options struct {
	Slug        string // e.g. "todo-app"
	DisplayName string // e.g. "Todo App"
	GithubUser  string // e.g. "johndoe"
	Author      string // e.g. "John Doe"
	TemplateURL string // git clone URL of the template repo
	OutputDir   string // destination directory
}

// filesToDelete are paths (relative to the project root) removed immediately
// after cloning. Paths ending with "/" are directories.
var filesToDelete = []string{
	".env",
	".env.production",
	"go.sum",
	"frontend/.env",
	"frontend/node_modules",
	"frontend/dist",
	"bin",
	"tmp",
	"uploads",
	".claude",
	".git",
	"create-saas-app", // the CLI source itself — never leaks into generated projects
}

// skipDirs are directory names that WalkAndReplace will never descend into.
var skipDirs = []string{
	"node_modules",
	".git",
	"create-saas-app", // skip the CLI subtree during replacement pass
}

// Scaffold runs the full scaffold pipeline:
//  1. Clone the template repository
//  2. Remove template-specific files
//  3. Apply string replacements across all text files
//  4. Copy .env.example → .env (and frontend variant)
//  5. go mod tidy → npm install → git init + initial commit
//  6. Print next-steps guidance
func Scaffold(opts Options) error {
	// 1. Clone.
	fmt.Printf("Cloning template into %s...\n", opts.OutputDir)
	if err := runCmd(".", "git", "clone", "--depth=1", opts.TemplateURL, opts.OutputDir); err != nil {
		return fmt.Errorf("cloning template: %w", err)
	}

	// 2. Remove template-specific files.
	fmt.Println("Removing template-specific files...")
	for _, rel := range filesToDelete {
		full := filepath.Join(opts.OutputDir, filepath.FromSlash(rel))
		if err := os.RemoveAll(full); err != nil {
			return fmt.Errorf("removing %s: %w", rel, err)
		}
	}

	// 3. Build replacement pairs (order is significant — see comments).
	pairs := buildReplacementPairs(opts)

	// 4. Walk the tree and apply replacements.
	fmt.Println("Applying string replacements...")
	if err := WalkAndReplace(opts.OutputDir, pairs, skipDirs); err != nil {
		return fmt.Errorf("replacing strings: %w", err)
	}

	// 5. Create working .env files from the committed examples.
	fmt.Println("Creating .env files from examples...")
	envPairs := [][2]string{
		{".env.example", ".env"},
		{"frontend/.env.example", "frontend/.env"},
	}
	for _, ep := range envPairs {
		src := filepath.Join(opts.OutputDir, filepath.FromSlash(ep[0]))
		dst := filepath.Join(opts.OutputDir, filepath.FromSlash(ep[1]))
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copying %s → %s: %w", ep[0], ep[1], err)
		}
	}

	// 5b. Inject generated secrets into .env so the project starts without
	//     manual editing of security keys.
	fmt.Println("Generating secrets...")
	if err := injectSecrets(filepath.Join(opts.OutputDir, ".env")); err != nil {
		return fmt.Errorf("injecting secrets: %w", err)
	}

	// 6. Post-processing.
	fmt.Println("Running go mod tidy...")
	if err := RunGoModTidy(opts.OutputDir); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	fmt.Println("Running npm install...")
	if err := RunNpmInstall(filepath.Join(opts.OutputDir, "frontend")); err != nil {
		return fmt.Errorf("npm install: %w", err)
	}

	fmt.Println("Initialising git repository...")
	if err := RunGitInit(opts.OutputDir, "Initial commit"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	// 7. Success.
	printNextSteps(opts)
	return nil
}

// Pair is a single old→new string substitution.
type Pair struct {
	Old string
	New string
}

// buildReplacementPairs returns the ordered list of substitutions to apply.
// More-specific strings come before the shorter strings that would otherwise
// clobber them (e.g. "saas-starter-admin" before "saas-starter").
func buildReplacementPairs(opts Options) []Pair {
	slugUnder := strings.ReplaceAll(opts.Slug, "-", "_")
	module := fmt.Sprintf("github.com/%s/%s", opts.GithubUser, opts.Slug)

	return []Pair{
		// 1–5: slug / service / DB identifier replacements.
		// Most-specific first so shorter patterns don't fire on their prefixes.
		{Old: "saas-starter-admin", New: opts.Slug + "-admin"},
		{Old: "saas-starter", New: opts.Slug},
		{Old: "saas-api", New: opts.Slug + "-api"},
		{Old: "saas_db", New: slugUnder + "_db"},
		{Old: "saas", New: slugUnder}, // DB user in docker-compose

		// 6–8: Go module path.
		// https:// variant first so the bare "github.com/…" rule doesn't create
		// a double-prefixed URL.
		// Swagger encodes Go type names with underscores (docs/docs.go), so we
		// also replace the underscore form before the dot form fires.
		{Old: "https://github.com/edsonmubezi/myapp", New: "https://" + module},
		{Old: "github_com_edsonmubezi_myapp", New: "github_com_" + opts.GithubUser + "_" + slugUnder},
		{Old: "github.com/edsonmubezi/myapp", New: module},

		// 9–10: Dockerfile binary paths.
		{Old: "/out/myapp", New: "/out/" + opts.Slug},
		{Old: "/app/myapp", New: "/app/" + opts.Slug},

		// 11–14: remaining bare "myapp" occurrences (database.yml URLs,
		// Grafana dashboard container names, Grafana datasource database name).
		// More-specific multi-word forms first so the bare catch-all doesn't
		// fire on them prematurely.
		{Old: "myapp-api", New: opts.Slug + "-api"},
		{Old: "myapp_test", New: slugUnder + "_test"},
		{Old: "myapp_prod", New: slugUnder + "_prod"},
		{Old: "myapp", New: slugUnder + "_db"}, // catch-all: DB name in URLs and datasource configs

		// 15–17: display name.
		// VITE_ variant before the bare string so the bare rule doesn't
		// partially mangle the env-var form.
		{Old: "VITE_APP_NAME=SaaS Starter", New: "VITE_APP_NAME=" + opts.DisplayName},
		{Old: "SaaS Starter", New: opts.DisplayName},
		{Old: "SaaS Platform", New: opts.DisplayName},

		// 13–14: .env.example DB settings.
		{Old: "DB_NAME=microfinance", New: "DB_NAME=" + slugUnder + "_db"},
		{Old: "DB_USER=microfinance", New: "DB_USER=" + slugUnder},

		// 15: catch-all for any remaining GitHub username occurrences
		// (comments, URLs not matched above, etc.).
		{Old: "edsonmubezi", New: opts.GithubUser},
	}
}

// injectSecrets replaces placeholder secret values in the .env file with
// cryptographically random strings that satisfy the application's validators:
//
//   - SECRET_KEY  — exactly 32 hex characters (16 random bytes → 32 hex chars)
//   - JWT_SECRET  — 64 hex characters (32 random bytes → 64 hex chars)
//   - DB_PASSWORD — 24 hex characters (12 random bytes → 24 hex chars)
func injectSecrets(envPath string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	secretKey, err := randHex(16) // 32-char hex
	if err != nil {
		return err
	}
	jwtSecret, err := randHex(32) // 64-char hex
	if err != nil {
		return err
	}
	dbPassword, err := randHex(12) // 24-char hex
	if err != nil {
		return err
	}

	content := string(data)
	content = strings.ReplaceAll(content, "your-32-character-encryption-key", secretKey)
	content = strings.ReplaceAll(content, "your-secret-jwt-key-change-this-in-production", jwtSecret)
	content = strings.ReplaceAll(content, "change-this-password", dbPassword)

	return os.WriteFile(envPath, []byte(content), 0600)
}

// randHex returns a hex-encoded string of n random bytes (length = 2*n).
func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func printNextSteps(opts Options) {
	fmt.Printf(`
✓ Project %q created successfully!

Next steps:
  cd %s

  # Fill in secrets (DB_PASSWORD, JWT_SECRET, SMTP_PASSWORD, etc.)
  nano .env

  # Start infrastructure
  docker compose up -d postgres redis

  # Database setup
  make migrate
  make seed

  # Run the API
  make run          # or: air  (hot reload)

  # Run the frontend (new terminal)
  cd frontend && npm run dev

API:      http://localhost:8080
Frontend: http://localhost:5173
Swagger:  http://localhost:8080/swagger/index.html

Default login: admin@example.com / Admin@1234
`, opts.Slug, opts.OutputDir)
}
