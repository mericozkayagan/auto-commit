package src

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func IsRootDirectory(cmd *cobra.Command, args []string) error {
	gitCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	stdoutStderr, err := gitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	rootDir := strings.TrimSpace(string(stdoutStderr))
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if rootDir == currentDir {
		return nil
	}

	return fmt.Errorf("please run the create commands from your repository root")
}
