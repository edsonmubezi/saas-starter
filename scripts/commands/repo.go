package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/edsonmubezi/myapp/scripts/generator"
)

func RunRepoCommand(args []string) {
	fs := flag.NewFlagSet("repo", flag.ExitOnError)

	var sourcePath string
	var blueprintPath string

	fs.StringVar(&sourcePath, "source", "", "Path to repository interface file (required)")
	fs.StringVar(&blueprintPath, "blueprint", "", "Path to blueprint JSON file (required)")

	_ = fs.Parse(args)

	if sourcePath == "" || blueprintPath == "" {
		fmt.Println("Usage: testgen repo --source path/to/repo.go --blueprint blueprints/entity.json")
		fs.PrintDefaults()
		os.Exit(1)
	}

	err := generator.GenerateRepoTests(sourcePath, blueprintPath)
	if err != nil {
		fmt.Println("Failed to generate repository tests:", err)
		os.Exit(1)
	}

	fmt.Println("Repository integration test generated successfully.")
}
