package scaffold

import (
	"os"
	"os/exec"
)

// runCmd executes name with args in dir, streaming stdout/stderr to the
// terminal in real time so the user can follow progress.
func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGoModTidy runs `go mod tidy` in dir, regenerating go.sum for the new
// module path.
func RunGoModTidy(dir string) error {
	return runCmd(dir, "go", "mod", "tidy")
}

// RunNpmInstall runs `npm install` in dir.
func RunNpmInstall(dir string) error {
	return runCmd(dir, "npm", "install")
}

// RunGitInit initialises a new git repository in dir, stages all files, and
// creates an initial commit with the given message.
func RunGitInit(dir, msg string) error {
	if err := runCmd(dir, "git", "init"); err != nil {
		return err
	}
	if err := runCmd(dir, "git", "add", "."); err != nil {
		return err
	}
	return runCmd(dir, "git", "commit", "-m", msg)
}
