// Package prompt provides interactive stdin prompts for missing CLI values.
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Values holds the answers collected from the user.
type Values struct {
	GithubUser string
	Author     string
}

// Collect prompts the user interactively for any values that are empty.
// Values already supplied (non-empty) are returned unchanged without prompting.
func Collect(githubUser, author string) (Values, error) {
	r := bufio.NewReader(os.Stdin)
	v := Values{GithubUser: githubUser, Author: author}

	if v.GithubUser == "" {
		val, err := ask(r, "GitHub username (for Go module path): ")
		if err != nil {
			return v, err
		}
		v.GithubUser = val
	}

	if v.Author == "" {
		val, err := ask(r, "Author / company name: ")
		if err != nil {
			return v, err
		}
		v.Author = val
	}

	return v, nil
}

func ask(r *bufio.Reader, question string) (string, error) {
	fmt.Print(question)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return strings.TrimSpace(line), nil
}
