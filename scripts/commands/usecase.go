package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/edsonmubezi/myapp/scripts/generator"
)

func RunUsecaseCommand(args []string) {
	fs := flag.NewFlagSet("usecase", flag.ExitOnError)

	var sourcePath string

	fs.StringVar(&sourcePath, "source", "", "Path to usecase implementation file (required)")

	_ = fs.Parse(args)

	if sourcePath == "" {
		fmt.Println("Usage: testgen usecase --source path/to/usecase.go")
		fs.PrintDefaults()
		os.Exit(1)
	}

	err := generator.GenerateUsecaseTests(sourcePath)
	if err != nil {
		fmt.Println("Failed to generate usecase tests:", err)
		os.Exit(1)
	}

	fmt.Println("Usecase unit tests generated successfully.")
}
