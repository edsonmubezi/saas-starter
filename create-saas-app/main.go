package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/edsonmubezi/saas-starter/create-saas-app/internal/prompt"
	"github.com/edsonmubezi/saas-starter/create-saas-app/internal/scaffold"
	"github.com/spf13/cobra"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func main() {
	root := &cobra.Command{
		Use:   "create-saas-app",
		Short: "Scaffold a new SaaS project from the saas-starter template",
		Long: `create-saas-app clones the saas-starter template, renames every
template string to match your project, runs go mod tidy and npm install,
and initialises a fresh git repository — all in one command.`,
	}

	root.AddCommand(newInitCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newInitCmd() *cobra.Command {
	var (
		displayName string
		githubUser  string
		author      string
		templateURL string
		outputDir   string
	)

	cmd := &cobra.Command{
		Use:   "init <slug>",
		Short: "Initialise a new project from the saas-starter template",
		Long: `Clone the saas-starter template and rename all template strings
to match the given project slug.

The slug must be lowercase, may contain hyphens in the middle, and must
start and end with a letter or digit (e.g. "todo-app", "my-crm").`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			if !slugRe.MatchString(slug) {
				return fmt.Errorf(
					"invalid slug %q: must match ^[a-z0-9][a-z0-9-]*[a-z0-9]$\n"+
						"  examples: todo-app, my-crm, invoicer",
					slug,
				)
			}

			if displayName == "" {
				displayName = toDisplayName(slug)
			}

			collected, err := prompt.Collect(githubUser, author)
			if err != nil {
				return fmt.Errorf("collecting inputs: %w", err)
			}
			if githubUser == "" {
				githubUser = collected.GithubUser
			}
			if author == "" {
				author = collected.Author
			}

			if outputDir == "" {
				outputDir = filepath.Join(".", slug)
			}

			if templateURL == "" {
				templateURL = "https://github.com/edsonmubezi/saas-starter"
			}

			return scaffold.Scaffold(scaffold.Options{
				Slug:        slug,
				DisplayName: displayName,
				GithubUser:  githubUser,
				Author:      author,
				TemplateURL: templateURL,
				OutputDir:   outputDir,
			})
		},
	}

	cmd.Flags().StringVarP(&displayName, "display-name", "d", "",
		"Human-readable project name (default: title-cased slug)")
	cmd.Flags().StringVarP(&githubUser, "github-user", "g", "",
		"GitHub username for the Go module path")
	cmd.Flags().StringVarP(&author, "author", "a", "",
		"Author or company name")
	cmd.Flags().StringVar(&templateURL, "template", "",
		"Template repository URL (default: https://github.com/edsonmubezi/saas-starter)")
	cmd.Flags().StringVar(&outputDir, "output-dir", "",
		"Output directory (default: ./<slug>)")

	return cmd
}

// toDisplayName converts a hyphenated slug into a title-cased display name.
// "todo-app" → "Todo App"
func toDisplayName(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
