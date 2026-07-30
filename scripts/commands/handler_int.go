package commands

import (
	"fmt"
	"os"

	"github.com/edsonmubezi/myapp/scripts/generator"
)

func RunHandlerIntCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: testgen handler-integration --source=api/handler/<entity>_handler.go")
		os.Exit(1)
	}

	var sourcePath string
	for _, arg := range os.Args[2:] {
		if len(arg) > 9 && arg[:9] == "--source=" {
			sourcePath = arg[9:]
		}
	}

	if sourcePath == "" {
		fmt.Println("Missing --source flag")
		os.Exit(1)
	}

	err := generator.GenerateHandlerIntegrationTests(sourcePath)
	if err != nil {
		fmt.Printf("Failed to generate handler integration tests: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Handler integration tests generated successfully.")
}
