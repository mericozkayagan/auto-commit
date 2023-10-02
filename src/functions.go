package src

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const (
	apiEndpoint = "https://api.openai.com/v1/engines/davinci/completions"
)

func AutoCommit(cmd *cobra.Command, args []string) error {
	directory := args[0]

	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("API key not set. Please set the environment variable 'OPENAI_API_KEY'.")
	}

	// Step 1: Get the diff
	changes, err := detectChanges(directory)
	if changes == "" {
		fmt.Println("No changes detected.")
		return err
	}

	// Step 2: Add changes to Git
	gitCmd := exec.Command("git", "add", directory)
	gitCmd.Dir, _ = getGitRepositoryPath()
	_, err = gitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error in git add " + directory)
	}
	fmt.Println(gitCmd)

	// Step 3: Generate commit message using ChatGPT
	commitMessage, err := generateCommitMessage(apiKey, changes)
	if commitMessage == "" {
		return fmt.Errorf("The commit message is null")
	}

	// Step 4: Ask for user confirmation
	fmt.Printf("Are you okay with the message: %s? (yes/no): ", commitMessage)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	answer = strings.TrimSpace(answer)

	// Step 5: Commit if the user approves
	if answer != "yes" {
		fmt.Println("Commit aborted.")
		return nil
	}

	// Step 6: Commit the changes
	gitCmd = exec.Command("git", "commit", "-m", "'"+commitMessage+"'")
	gitCmd.Dir, _ = getGitRepositoryPath()
	_, err = gitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error in git commit -m  " + commitMessage)
	}
	fmt.Println(gitCmd)

	fmt.Println("Commit successful.")

	return nil
}

func detectChanges(directory string) (string, error) {
	cmd := exec.Command("git", "diff", directory)
	cmd.Dir, _ = getGitRepositoryPath()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return " ", fmt.Errorf("error in git diff " + directory)
	}

	return string(output), nil
}

func generateCommitMessage(apiKey string, changes string) (string, error) {
	// Construct a prompt for GPT to generate the commit message
	prompt := "Generate commit message with a prefix in the list [fix:, feat:, refactor:, docs:] for the following changes: " + changes

	// Generate the commit message using GPT
	commitMessage, err := getChatGPTResponseUsingPrompt(apiKey, prompt)
	if err != nil {
		return "", fmt.Errorf("Error generating commit message: %v", err)
	}

	return commitMessage, nil
}

func getChatGPTResponseUsingPrompt(apiKey string, payload string) (string, error) {
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: payload,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}

func getGitRepositoryPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error running 'git rev-parse --show-toplevel': %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
