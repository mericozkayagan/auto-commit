/*
Copyright © 2023 Meric Ozkayagan mericozkayagan@gmail.com
*/
package cmd

import (
	"os"

	"github.com/mericozkayagan/auto-commit/src"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "autocommit [directory]",
	Short:             "Automatically commit changes to a Git repository",
	PersistentPreRunE: src.IsRootDirectory,
	Long: `Autocommit is a CLI tool that monitors changes in a directory and automatically
commits those changes to a Git repository. It retrieves the changes in the specified
directory and generates a commit message using an AI-powered model. The generated commit
message is prefixed with the provided prefix.

The tool periodically checks for changes in the directory and commits them if any are found.`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument, which is the directory.
	RunE: src.AutoCommit,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// func init() {
// 	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
