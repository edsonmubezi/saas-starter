package main

import (
	"fmt"
	"os"

	"github.com/edsonmubezi/myapp/scripts/commands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected a subcommand: repo | usecase | handler | handler-int")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "repo":
		commands.RunRepoCommand(os.Args[2:])
	case "usecase":
		commands.RunUsecaseCommand(os.Args[2:])
	case "handler":
		commands.RunHandlerCommand(os.Args[2:])
	case "handler-integration":
		commands.RunHandlerIntCommand()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}
