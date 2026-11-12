package parser

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
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

// GenerateRandomString generates a random alphanumeric string of the specified
// length using a cryptographically secure source.
func GenerateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
	max := big.NewInt(int64(len(charset)))
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// ParseCommitResponse parses the response from Bedrock and returns formatted commit messages
func ParseCommitResponse(response string) ([]string, error) {
	// Clean the response - remove any markdown formatting or extra text
	cleanedResponse := regexp.MustCompile("```json\n?").ReplaceAllString(response, "")
	cleanedResponse = regexp.MustCompile("```\n?").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("^[^{]*").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("[^}]*$").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	var commitResp CommitResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &commitResp); err != nil {
		// Fallback: try to extract conventional commit format from the response
		fallbackCommits := ExtractFallbackCommits(response)
		if len(fallbackCommits) > 0 {
			return fallbackCommits, nil
		}
		return nil, err
	}

	if len(commitResp.Commits) == 0 {
		return nil, nil
	}

	var messages []string
	validTypes := map[string]bool{
		"feat": true, "fix": true, "build": true, "chore": true, "ci": true,
		"docs": true, "style": true, "refactor": true, "perf": true, "test": true,
	}

	for _, commit := range commitResp.Commits {
		if commit.Type == "" || commit.Scope == "" || commit.Description == "" {
			return nil, nil
		}

		if !validTypes[commit.Type] {
			return nil, nil
		}

		messages = append(messages, fmt.Sprintf("%s(%s): %s", commit.Type, commit.Scope, commit.Description))
	}

	return messages, nil
}

// ExtractFallbackCommits extracts conventional commit messages from unstructured response
func ExtractFallbackCommits(response string) []string {
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

// ParseChangelogResponse parses the response from Bedrock and returns formatted changelog
func ParseChangelogResponse(response string) (string, error) {
	// Clean the response - remove any markdown formatting or extra text
	cleanedResponse := regexp.MustCompile("```json\n?").ReplaceAllString(response, "")
	cleanedResponse = regexp.MustCompile("```\n?").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("^[^{]*").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = regexp.MustCompile("[^}]*$").ReplaceAllString(cleanedResponse, "")
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	var changelogResp ChangelogResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &changelogResp); err != nil {
		// Fallback: try to extract changelog format from the response
		fallbackChangelog := ExtractFallbackChangelog(response)
		if fallbackChangelog != "" {
			return fallbackChangelog, nil
		}
		return "", err
	}

	var changelog strings.Builder

	if len(changelogResp.Changelog.Removed) > 0 {
		changelog.WriteString("### Removed\n\n")
		for _, item := range changelogResp.Changelog.Removed {
			fmt.Fprintf(&changelog, "- %s\n", item)
		}
		changelog.WriteString("\n")
	}

	if len(changelogResp.Changelog.Added) > 0 {
		changelog.WriteString("### Added\n\n")
		for _, item := range changelogResp.Changelog.Added {
			fmt.Fprintf(&changelog, "- %s\n", item)
		}
		changelog.WriteString("\n")
	}

	if len(changelogResp.Changelog.Changed) > 0 {
		changelog.WriteString("### Changed\n\n")
		for _, item := range changelogResp.Changelog.Changed {
			fmt.Fprintf(&changelog, "- %s\n", item)
		}
		changelog.WriteString("\n")
	}

	return strings.TrimSpace(changelog.String()), nil
}

// ExtractFallbackChangelog extracts changelog from unstructured response
func ExtractFallbackChangelog(response string) string {
	changelogRegex := regexp.MustCompile(`(### (?:Added|Changed|Removed)\n\n(?:- .+\n?)+)`)
	matches := changelogRegex.FindAllString(response, -1)
	return strings.Join(matches, "\n")
}
