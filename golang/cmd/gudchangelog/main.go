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

// ChangelogEntry represents a changelog section
type ChangelogEntry struct {
	Added   []string `json:"added"`
	Changed []string `json:"changed"`
	Removed []string `json:"removed"`
}

// ChangelogResponse represents the JSON response from Bedrock
type ChangelogResponse struct {
	Changelog ChangelogEntry `json:"changelog"`
}

// getGitDiff retrieves the diff between current branch and target branch
func getGitDiff(targetBranch string) (string, error) {
	cmd := exec.Command("git", "diff", targetBranch+"..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return string(output), nil
}

// invokeBedrockModel invokes Bedrock directly using API key authentication
func invokeBedrockModel(prompt, repoPath string) (string, error) {
	client, err := bedrock.NewClient()
	if err != nil {
		return "", err
	}

	// Create the prompt with the same template as before
	fullPrompt := fmt.Sprintf(`Analyze the following git diff and generate changelog entries in the specified JSON format.

Repository root: %s

<git_diff>%s</git_diff>

Respond with JSON matching this exact schema:

{
  "type": "object",
  "properties": {
    "changelog": {
      "type": "object",
      "properties": {
        "added": {
          "type": "array",
          "items": {"type": "string"},
          "description": "List of new features or additions"
        },
        "changed": {
          "type": "array",
          "items": {"type": "string"},
          "description": "List of changes to existing functionality"
        },
        "removed": {
          "type": "array",
          "items": {"type": "string"},
          "description": "List of removed features or functionality"
        }
      },
      "required": ["added", "changed", "removed"],
      "additionalProperties": false
    }
  },
  "required": ["changelog"],
  "additionalProperties": false
}

Rules:
- Follow Keep a Changelog format (http://keepachangelog.com/)
- Be concise and clear
- Focus on WHAT changed and WHY
- Do not include any explanatory text outside the JSON
- Each category should contain meaningful entries`, repoPath, prompt)

	return client.InvokeModel(fullPrompt, repoPath)
}

// parseChangelogResponse parses the response from Bedrock and returns formatted changelog entries
func parseChangelogResponse(response string) (*ChangelogEntry, error) {
	// Clean the response - remove any markdown formatting or extra text
	cleanedResponse := regexp.MustCompile("```json\n?").ReplaceAllString(response, "")
	cleanedResponse = regexp.MustCompile("```\n?").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("^[^{]*").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("[^}]*$").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	var changelogResp ChangelogResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &changelogResp); err != nil {
		return nil, fmt.Errorf("failed to parse changelog response: %w", err)
	}

	return &changelogResp.Changelog, nil
}

// formatChangelog formats the changelog entry for display
func formatChangelog(changelog *ChangelogEntry) string {
	var result strings.Builder

	result.WriteString("## [Unreleased]\n\n")

	if len(changelog.Added) > 0 {
		result.WriteString("### Added\n")
		for _, item := range changelog.Added {
			result.WriteString(fmt.Sprintf("- %s\n", item))
		}
		result.WriteString("\n")
	}

	if len(changelog.Changed) > 0 {
		result.WriteString("### Changed\n")
		for _, item := range changelog.Changed {
			result.WriteString(fmt.Sprintf("- %s\n", item))
		}
		result.WriteString("\n")
	}

	if len(changelog.Removed) > 0 {
		result.WriteString("### Removed\n")
		for _, item := range changelog.Removed {
			result.WriteString(fmt.Sprintf("- %s\n", item))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// run is the main function that orchestrates the changelog generation
func run() error {

	// Check command line arguments
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: gudchangelog <target-branch>")
	}

	targetBranch := os.Args[1]

	// Get repository root path for better context
	repoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		repoRoot = []byte(".")
	}
	repoPath := strings.TrimSpace(string(repoRoot))

	// Get git diff
	diffOutput, err := getGitDiff(targetBranch)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Check if diffOutput is empty
	if strings.TrimSpace(diffOutput) == "" {
		fmt.Println(">> No changes found between current branch and", targetBranch)
		return nil
	}

	// Escape backslashes for JSON
	diffOutput = strings.ReplaceAll(diffOutput, "\\", "\\\\")

	// Generate changelog
	fmt.Println("🤖 Generating changelog...")
	completion, err := invokeBedrockModel(diffOutput, repoPath)
	if err != nil {
		return fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	if completion == "" {
		fmt.Println("Sorry. No changelog could be generated.")
		return nil
	}

	// Parse response
	changelog, err := parseChangelogResponse(completion)
	if err != nil {
		fmt.Printf("Warning: Failed to parse structured response: %v\n", err)
		// Fallback to raw response
		fmt.Println("Raw response:")
		fmt.Println(completion)
		return nil
	}

	// Format and display the changelog
	formattedChangelog := formatChangelog(changelog)

	fmt.Println()
	fmt.Println("📝 Generated changelog:")
	fmt.Println("========================")
	fmt.Println(formattedChangelog)

	// Ask if user wants to prepend to CHANGELOG.md
	fmt.Print("Prepend this content to CHANGELOG.md? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		// Write to CHANGELOG.md
		changelogFile := "CHANGELOG.md"

		// Read existing content if file exists
		var existingContent string
		if content, err := os.ReadFile(changelogFile); err == nil {
			existingContent = string(content)
		}

		// Create new content with generated changelog at the top
		newContent := formattedChangelog
		if existingContent != "" {
			newContent += "\n---\n\n" + existingContent
		}

		// Write to file
		if err := os.WriteFile(changelogFile, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
		}

		fmt.Printf("✅ Changelog written to %s\n", changelogFile)
	} else {
		fmt.Println("Changelog generation completed.")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf(">> %v", err)
	}
}
