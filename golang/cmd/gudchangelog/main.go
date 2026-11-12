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
	"github.com/gudlyf/GudCommit/golang/pkg/diffguard"
)

// stdin is shared by all prompts so buffered input is never lost between them.
var stdin = bufio.NewReader(os.Stdin)

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

// resolveCommit validates a user-supplied ref and resolves it to a commit SHA.
// Rejecting a leading "-" and passing --end-of-options prevents the ref from
// being interpreted as a git option (e.g. "--output=/path").
func resolveCommit(ref string) (string, error) {
	if ref == "" || strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("invalid target branch %q", ref)
	}
	// No shell is involved, ref cannot start with "-", and --end-of-options stops option parsing.
	out, err := exec.Command("git", "rev-parse", "--verify", "--quiet", "--end-of-options", ref+"^{commit}").Output() //nolint:gosec // G204: validated, see above
	if err != nil {
		return "", fmt.Errorf("target branch %q is not a valid commit", ref)
	}
	sha := strings.TrimSpace(string(out))
	if !commitSHA.MatchString(sha) {
		return "", fmt.Errorf("unexpected rev-parse output for %q", ref)
	}
	return sha, nil
}

var commitSHA = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// getGitDiff retrieves the diff between the target commit (a resolved SHA) and HEAD
func getGitDiff(targetCommit string) (string, error) {
	cmd := exec.Command("git", "diff", targetCommit+"..HEAD") //nolint:gosec // G204: targetCommit is a hex SHA from resolveCommit
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return string(output), nil
}

// invokeLLM invokes the configured LLM provider (Bedrock or Anthropic/Claude)
func invokeLLM(prompt, repoPath string) (string, error) {
	invoker, err := bedrock.NewInvoker()
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

	return invoker.InvokeModel(fullPrompt, repoPath)
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
			fmt.Fprintf(&result, "- %s\n", item)
		}
		result.WriteString("\n")
	}

	if len(changelog.Changed) > 0 {
		result.WriteString("### Changed\n")
		for _, item := range changelog.Changed {
			fmt.Fprintf(&result, "- %s\n", item)
		}
		result.WriteString("\n")
	}

	if len(changelog.Removed) > 0 {
		result.WriteString("### Removed\n")
		for _, item := range changelog.Removed {
			fmt.Fprintf(&result, "- %s\n", item)
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

	targetCommit, err := resolveCommit(targetBranch)
	if err != nil {
		return err
	}

	// Get git diff
	diffOutput, err := getGitDiff(targetCommit)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Check if diffOutput is empty
	if strings.TrimSpace(diffOutput) == "" {
		fmt.Println(">> No changes found between current branch and", targetBranch)
		return nil
	}

	// Check for secrets and size before the diff leaves the machine
	if ok, guardErr := diffguard.Enforce(diffOutput, stdin, os.Stdout); guardErr != nil {
		return fmt.Errorf("failed to read user input: %w", guardErr)
	} else if !ok {
		return nil
	}

	// Escape backslashes for JSON
	diffOutput = strings.ReplaceAll(diffOutput, "\\", "\\\\")

	// Generate changelog
	fmt.Println("🤖 Generating changelog...")
	completion, err := invokeLLM(diffOutput, repoPath)
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
	response, err := stdin.ReadString('\n')
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
		// CHANGELOG.md is a public, committed project file; 0644 is intended.
		if err := os.WriteFile(changelogFile, []byte(newContent), 0644); err != nil { //nolint:gosec // G306: see above
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
