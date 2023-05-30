/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"gopkg.in/resty.v1"
)

var (
	prefix string
)

// rootCmd represents the base command when called without any subcommands

var rootCmd = &cobra.Command{
	Use:   "autocommit [file-path]",
	Short: "Automatically commit changes to a Git repository",
	Long: `Autocommit is a CLI tool that monitors changes in a file and automatically
commits those changes to a Git repository. It retrieves the changes in the specified
file and generates a commit message using an AI-powered model. The generated commit
message is prefixed with the provided prefix.

The tool periodically checks for changes in the file and commits them if any are found.`,
	Args: cobra.ExactArgs(1),
	Run:  autoCommit,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.auto-commit.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Commit message prefix")

}

func autoCommit(cmd *cobra.Command, args []string) {
	filePath := args[0]

	apiKey := os.Getenv("CHATGPT_API_KEY")
	if apiKey == "" {
		fmt.Println("API key not set. Please set the environment variable 'CHATGPT_API_KEY'.")
		os.Exit(1)
	}

	for {
		changes, err := getChanges(filePath)
		if err != nil {
			fmt.Println("Error getting file changes:", err)
			os.Exit(1)
		}

		if len(changes) > 0 {
			commitMessage := generateCommitMessage(prefix, changes, apiKey)
			err := executeGitCommand(filePath, commitMessage)
			if err != nil {
				fmt.Println("Error executing git command:", err)
				os.Exit(1)
			}
			fmt.Printf("Committed changes to %s: %s\n", filePath, commitMessage)
		}

		time.Sleep(1 * time.Minute)
	}
}

func getChanges(filePath string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=no", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	changes := make([]string, 0, len(lines))

	for _, line := range lines {
		if line != "" {
			changes = append(changes, line)
		}
	}

	return changes, nil
}

func generateCommitMessage(prefix string, changes []string, apiKey string) string {
	var messageBuffer bytes.Buffer

	for _, change := range changes {
		messageBuffer.WriteString("- " + change + "\n")
	}

	changeDetails := messageBuffer.String()
	commitMessage := getChatGPTResponse(prefix, changeDetails, apiKey)

	return commitMessage
}

func getChatGPTResponse(prefix string, changeDetails string, apiKey string) string {
	apiEndpoint := "https://api.openai.com/v1/engines/davinci-codex/completions"

	data := fmt.Sprintf(`{
		"prompt": "%s\n\n%s",
		"temperature": 0.7,
		"max_tokens": 64,
		"stop": "\n"
	}`, prefix, changeDetails)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey).
		SetBody(data).
		Post(apiEndpoint)

	if err != nil {
		log.Fatal("Error sending request to ChatGPT API:", err)
	}

	responseBody := resp.Body()
	commitMessage := gjson.Get(string(responseBody), "choices.0.text").String()

	return commitMessage
}

func executeGitCommand(filePath string, commitMessage string) error {
	cmd := exec.Command("git", "add", filePath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", commitMessage)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
