package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gudlyf/GudCommit/golang/pkg/bedrock"
)

// CommitMessage represents a structured commit message
type CommitMessage struct {
	Type        string `json:"type"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

// CommitResponse represents the JSON response from Bedrock
type CommitResponse struct {
	Commits []CommitMessage `json:"commits"`
}

// checkStagedChanges checks if there are staged changes
func checkStagedChanges() error {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	err := cmd.Run()
	if err == nil {
		// Exit code 0 means no staged changes (quiet = no differences)
		return fmt.Errorf("no staged changes found. Please stage your changes first with 'git add'")
	}
	// Exit code 1 means there ARE staged changes (not quiet = differences exist)
	return nil
}

// getGitDiff retrieves the staged changes from git using git command
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return string(output), nil
}

// executeGitCommit executes the git commit with the given message
func executeGitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}
	return nil
}

// executeGitCommitEdit executes git commit with editor
func executeGitCommitEdit(message string) error {
	cmd := exec.Command("git", "commit", "-e", "-m", message)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// invokeBedrockModel invokes Bedrock using the shared package
func invokeBedrockModel(prompt, repoPath string) (string, error) {
	client, err := bedrock.NewClient()
	if err != nil {
		return "", err
	}

	// Create the prompt with the same template as before
	fullPrompt := fmt.Sprintf(`Analyze the following git diff and generate commit messages in the specified JSON format.

Repository root: %s

<git_diff>%s</git_diff>

Respond with JSON matching this exact schema:

{
  "type": "object",
  "properties": {
    "commits": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "type": {
            "type": "string",
            "enum": ["feat", "fix", "build", "chore", "ci", "docs", "style", "refactor", "perf", "test"]
          },
          "scope": {
            "type": "string",
            "description": "The full file path or component being changed (e.g., 'src/utils.js', 'terraform/module/main.tf', 'golang/cmd/gudcommit/main.go')"
          },
          "description": {
            "type": "string",
            "description": "Brief description of the change and why it was done"
          }
        },
        "required": ["type", "scope", "description"],
        "additionalProperties": false
      }
    }
  },
  "required": ["commits"],
  "additionalProperties": false
}

Rules:
- Use conventional commit format: type(scope): description
- Types: feat, fix, build, chore, ci, docs, style, refactor, perf, test
- Scope should be the full file path (e.g., 'golang/cmd/gudcommit/main.go', 'terraform/module/bedrock.tf')
- Be concise and clear
- Focus on WHAT changed and WHY
- Do not include any explanatory text outside the JSON
- Each changed file should have its own commit entry`, repoPath, prompt)

	return client.InvokeModel(fullPrompt, repoPath)
}

// parseCommitResponse parses the response from Bedrock and returns formatted commit messages
func parseCommitResponse(response string) ([]string, error) {
	// Clean the response - remove any markdown formatting or extra text
	cleanedResponse := regexp.MustCompile("```json\n?").ReplaceAllString(response, "")
	cleanedResponse = regexp.MustCompile("```\n?").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("^[^{]*").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("[^}]*$").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	var commitResp CommitResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &commitResp); err != nil {
		// Fallback: try to extract conventional commit format from the response
		fallbackCommits := extractFallbackCommits(response)
		if len(fallbackCommits) > 0 {
			return fallbackCommits, nil
		}
		return nil, fmt.Errorf("failed to parse commit response: %w", err)
	}

	if len(commitResp.Commits) == 0 {
		return nil, fmt.Errorf("no commits found in response")
	}

	var messages []string
	validTypes := map[string]bool{
		"feat": true, "fix": true, "build": true, "chore": true, "ci": true,
		"docs": true, "style": true, "refactor": true, "perf": true, "test": true,
	}

	for _, commit := range commitResp.Commits {
		if commit.Type == "" || commit.Description == "" {
			return nil, fmt.Errorf("invalid commit format: missing required fields")
		}

		if !validTypes[commit.Type] {
			return nil, fmt.Errorf("invalid commit type: %s", commit.Type)
		}

		// Format the commit message with proper scope handling
		if commit.Scope == "" || commit.Scope == "null" {
			// No scope - use simple format
			messages = append(messages, fmt.Sprintf("%s: %s", commit.Type, commit.Description))
		} else {
			// Has scope - use conventional format
			messages = append(messages, fmt.Sprintf("%s(%s): %s", commit.Type, commit.Scope, commit.Description))
		}
	}

	return messages, nil
}

// extractFallbackCommits extracts conventional commit messages from unstructured response
func extractFallbackCommits(response string) []string {
	conventionalCommitRegex := regexp.MustCompile(`^(feat|fix|build|chore|ci|docs|style|refactor|perf|test)\([^)]+\): .+$`)
	lines := strings.Split(response, "\n")
	var commits []string

	for _, line := range lines {
		if conventionalCommitRegex.MatchString(strings.TrimSpace(line)) {
			commits = append(commits, strings.TrimSpace(line))
		}
	}

	return commits
}

// promptUser prompts the user for confirmation
func promptUser(message string) (string, error) {
	fmt.Print("Proceed with the commit? (y/n or e to Edit): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.ToLower(response)), nil
}

// run is the main function that orchestrates the commit message generation
func run() error {

	// Check for staged changes first
	if err := checkStagedChanges(); err != nil {
		fmt.Printf("❌ %v\n", err)
		return nil
	}

	// Get git diff
	diffOutput, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Check if diffOutput is empty
	if strings.TrimSpace(diffOutput) == "" {
		fmt.Println(">> No changes to commit.")
		return nil
	}

	// Get the repository root path for better context
	repoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		repoRoot = []byte(".")
	}
	repoPath := strings.TrimSpace(string(repoRoot))

	// Escape backslashes for JSON
	diffOutput = strings.ReplaceAll(diffOutput, "\\", "\\\\")

	// Generate commit message
	fmt.Println("🤖 Generating commit message...")
	completion, err := invokeBedrockModel(diffOutput, repoPath)
	if err != nil {
		return fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	if completion == "" {
		fmt.Println("Sorry. No commit message could be generated.")
		return nil
	}

	// Parse response
	commitMessages, err := parseCommitResponse(completion)
	if err != nil {
		// Fallback to raw response
		commitMessages = []string{strings.TrimSpace(completion)}
	}

	if len(commitMessages) == 0 {
		fmt.Println("Sorry. No commit message could be generated.")
		return nil
	}

	// Display the generated message(s)
	fmt.Println()
	fmt.Println("📝 Generated commit message(s):")

	// Show all commit messages
	for i, message := range commitMessages {
		if len(commitMessages) > 1 {
			fmt.Printf("\033[1m%d. %s\033[0m\n", i+1, message)
		} else {
			fmt.Printf("\033[1m%s\033[0m\n", message)
		}
	}
	fmt.Println()

	// Create a comprehensive commit message
	var mainMessage string
	if len(commitMessages) > 1 {
		// Combine all messages into one comprehensive message
		mainMessage = strings.Join(commitMessages, "\n")
		fmt.Println("📝 Combined commit message:")
		fmt.Printf("\033[1m%s\033[0m\n", mainMessage)
		fmt.Println()
	} else {
		mainMessage = commitMessages[0]
	}

	// Prompt user for confirmation
	response, err := promptUser(mainMessage)
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	switch response {
	case "y", "yes":
		// Execute git commit
		if err := executeGitCommit(mainMessage); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
		fmt.Println("✅ Commit successful!")
	case "e", "edit":
		// Execute git commit with editor
		if err := executeGitCommitEdit(mainMessage); err != nil {
			return fmt.Errorf("failed to commit with editor: %w", err)
		}
	default:
		fmt.Println("Commit canceled.")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf(">> %v", err)
	}
}
