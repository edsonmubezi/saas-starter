package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/edsonmubezi/myapp/scripts/generator"
)

// RunHandlerCommand handles CLI command: testgen handler --source=...
func RunHandlerCommand(args []string) {
	fs := flag.NewFlagSet("handler", flag.ExitOnError)

	var sourcePath string
	fs.StringVar(&sourcePath, "source", "", "Path to handler file (required)")

	_ = fs.Parse(args)

	if sourcePath == "" {
		fmt.Println("Usage: testgen handler --source=path/to/handler.go")
		fs.PrintDefaults()
		os.Exit(1)
	}

	err := generator.GenerateHandlerUnitTests(sourcePath)
	if err != nil {
		fmt.Println("Failed to generate handler unit tests:", err)
		os.Exit(1)
	}

	fmt.Println("Handler unit test generated successfully.")
}
