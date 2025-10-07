package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseCommitResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected []string
		hasError bool
	}{
		{
			name: "Valid JSON response",
			response: `{
				"commits": [
					{
						"type": "feat",
						"scope": "auth",
						"description": "Add user authentication"
					},
					{
						"type": "fix",
						"scope": "api",
						"description": "Fix rate limiting bug"
					}
				]
			}`,
			expected: []string{
				"feat(auth): Add user authentication",
				"fix(api): Fix rate limiting bug",
			},
			hasError: false,
		},
		{
			name: "Fallback conventional commits",
			response: `feat(auth): Add user authentication
fix(api): Fix rate limiting bug
docs(readme): Update installation instructions`,
			expected: []string{
				"feat(auth): Add user authentication",
				"fix(api): Fix rate limiting bug",
				"docs(readme): Update installation instructions",
			},
			hasError: false,
		},
		{
			name:     "Empty response",
			response: "",
			expected: nil,
			hasError: true,
		},
		{
			name: "Invalid JSON with fallback",
			response: `Some text before
feat(auth): Add user authentication
Some text after`,
			expected: []string{
				"feat(auth): Add user authentication",
			},
			hasError: false,
		},
		{
			name:     "JSON with markdown formatting",
			response: "```json\n{\n\t\"commits\": [\n\t\t{\n\t\t\t\"type\": \"feat\",\n\t\t\t\"scope\": \"ui\",\n\t\t\t\"description\": \"Add dark mode toggle\"\n\t\t}\n\t]\n}\n```",
			expected: []string{
				"feat(ui): Add dark mode toggle",
			},
			hasError: false,
		},
		{
			name: "Invalid commit types",
			response: `{
				"commits": [
					{
						"type": "invalid",
						"scope": "test",
						"description": "This should fail"
					}
				]
			}`,
			expected: nil,
			hasError: false,
		},
		{
			name: "Missing required fields",
			response: `{
				"commits": [
					{
						"type": "feat",
						"scope": "",
						"description": "Missing scope"
					}
				]
			}`,
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCommitResponse(tt.response)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d commits, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestParseChangelogResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
		hasError bool
	}{
		{
			name: "Valid JSON response",
			response: `{
				"changelog": {
					"added": ["New feature A", "New feature B"],
					"changed": ["Updated API", "Improved performance"],
					"removed": ["Deprecated feature"]
				}
			}`,
			expected: `### Removed

- Deprecated feature

### Added

- New feature A
- New feature B

### Changed

- Updated API
- Improved performance`,
			hasError: false,
		},
		{
			name: "Fallback markdown",
			response: `### Added

- New feature A
- New feature B

### Changed

- Updated API`,
			expected: `### Added

- New feature A
- New feature B

### Changed

- Updated API`,
			hasError: false,
		},
		{
			name:     "Empty response",
			response: "",
			expected: "",
			hasError: true,
		},
		{
			name:     "JSON with markdown formatting",
			response: "```json\n{\n\t\"changelog\": {\n\t\t\"added\": [\"Feature X\"],\n\t\t\"changed\": [\"Performance improvement\"]\n\t}\n}\n```",
			expected: `### Added

- Feature X

### Changed

- Performance improvement`,
			hasError: false,
		},
		{
			name: "Empty changelog sections",
			response: `{
				"changelog": {
					"added": [],
					"changed": [],
					"removed": []
				}
			}`,
			expected: "",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseChangelogResponse(tt.response)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestExtractFallbackCommits(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected []string
	}{
		{
			name: "Multiple conventional commits",
			response: `Some text before
feat(auth): Add user authentication
fix(api): Fix rate limiting bug
docs(readme): Update installation instructions
Some text after`,
			expected: []string{
				"feat(auth): Add user authentication",
				"fix(api): Fix rate limiting bug",
				"docs(readme): Update installation instructions",
			},
		},
		{
			name:     "No conventional commits",
			response: `Just some regular text without conventional commits`,
			expected: []string{},
		},
		{
			name:     "Empty response",
			response: "",
			expected: []string{},
		},
		{
			name: "Mixed valid and invalid commits",
			response: `feat(auth): Valid commit
invalid: Invalid commit
fix(api): Another valid commit
not a commit at all`,
			expected: []string{
				"feat(auth): Valid commit",
				"fix(api): Another valid commit",
			},
		},
		{
			name: "All commit types",
			response: `feat(auth): New feature
fix(api): Bug fix
build(deps): Build change
chore(ci): Chore task
ci(workflow): CI change
docs(readme): Documentation
style(code): Code style
refactor(utils): Refactoring
perf(api): Performance
test(unit): Test addition`,
			expected: []string{
				"feat(auth): New feature",
				"fix(api): Bug fix",
				"build(deps): Build change",
				"chore(ci): Chore task",
				"ci(workflow): CI change",
				"docs(readme): Documentation",
				"style(code): Code style",
				"refactor(utils): Refactoring",
				"perf(api): Performance",
				"test(unit): Test addition",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFallbackCommits(tt.response)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d commits, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestExtractFallbackChangelog(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name: "Valid changelog sections",
			response: `### Added

- New feature A
- New feature B

### Changed

- Updated API
- Improved performance`,
			expected: `### Added

- New feature A
- New feature B

### Changed

- Updated API
- Improved performance`,
		},
		{
			name: "Multiple sections",
			response: `### Removed

- Deprecated feature

### Added

- New feature

### Changed

- Updated feature`,
			expected: `### Removed

- Deprecated feature

### Added

- New feature

### Changed

- Updated feature`,
		},
		{
			name:     "No changelog sections",
			response: `Just some regular text without changelog sections`,
			expected: "",
		},
		{
			name:     "Empty response",
			response: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFallbackChangelog(tt.response)

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test different lengths
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			result := GenerateRandomString(length)

			if len(result) != length {
				t.Errorf("Expected length %d, got %d", length, len(result))
			}

			// Check that all characters are from the expected charset
			charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
			for _, char := range result {
				if !strings.ContainsRune(charset, char) {
					t.Errorf("Character %c not in expected charset", char)
				}
			}
		})
	}

	// Test that multiple calls produce different results (very high probability)
	t.Run("Uniqueness", func(t *testing.T) {
		results := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result := GenerateRandomString(16)
			if results[result] {
				t.Errorf("Generated duplicate string: %s", result)
			}
			results[result] = true
		}
	})
}

func TestCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		commit   CommitMessage
		expected string
	}{
		{
			name: "Standard commit",
			commit: CommitMessage{
				Type:        "feat",
				Scope:       "auth",
				Description: "Add user authentication",
			},
			expected: "feat(auth): Add user authentication",
		},
		{
			name: "Fix commit",
			commit: CommitMessage{
				Type:        "fix",
				Scope:       "api",
				Description: "Fix rate limiting bug",
			},
			expected: "fix(api): Fix rate limiting bug",
		},
		{
			name: "Documentation commit",
			commit: CommitMessage{
				Type:        "docs",
				Scope:       "readme",
				Description: "Update installation instructions",
			},
			expected: "docs(readme): Update installation instructions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmt.Sprintf("%s(%s): %s", tt.commit.Type, tt.commit.Scope, tt.commit.Description)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestChangelogEntry(t *testing.T) {
	entry := ChangelogEntry{
		Added:   []string{"Feature A", "Feature B"},
		Changed: []string{"API Update", "Performance"},
		Removed: []string{"Deprecated Feature"},
	}

	if len(entry.Added) != 2 {
		t.Errorf("Expected 2 added items, got %d", len(entry.Added))
	}

	if len(entry.Changed) != 2 {
		t.Errorf("Expected 2 changed items, got %d", len(entry.Changed))
	}

	if len(entry.Removed) != 1 {
		t.Errorf("Expected 1 removed item, got %d", len(entry.Removed))
	}
}

// Benchmark tests
func BenchmarkGenerateRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRandomString(32)
	}
}

func BenchmarkParseCommitResponse(b *testing.B) {
	response := `{
		"commits": [
			{
				"type": "feat",
				"scope": "auth",
				"description": "Add user authentication"
			}
		]
	}`

	for i := 0; i < b.N; i++ {
		ParseCommitResponse(response)
	}
}

func BenchmarkExtractFallbackCommits(b *testing.B) {
	response := `feat(auth): Add user authentication
fix(api): Fix rate limiting bug
docs(readme): Update installation instructions`

	for i := 0; i < b.N; i++ {
		ExtractFallbackCommits(response)
	}
}
