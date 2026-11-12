package bedrock

import "testing"

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"plain", "feat(x): add y", "feat(x): add y"},
		{"keeps newline and tab", "a\n\tb", "a\n\tb"},
		{"crlf", "a\r\nb", "a\nb"},
		{"ansi color", "\x1b[31mred\x1b[0m", "[31mred[0m"},
		{"carriage return overwrite", "safe\rEVIL", "safeEVIL"},
		{"osc hyperlink", "\x1b]8;;http://x\x07link\x1b]8;;\x07", "]8;;http://xlink]8;;"},
		{"bidi override", "a\u202Eb\u2066c", "abc"},
		{"c1 csi", "a\u009bb", "ab"},
		{"unicode kept", "fix: naïve café ✓", "fix: naïve café ✓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeOutput(tt.in); got != tt.want {
				t.Errorf("SanitizeOutput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// The claude-cli provider must never expose tools to the model: the git diff in
// the prompt is untrusted and could carry prompt-injection instructions.
func TestClaudeCLIArgsDisableTools(t *testing.T) {
	args := claudeCLIArgs()
	found := false
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--tools" && args[i+1] == "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("claudeCLIArgs() = %q; want --tools \"\"", args)
	}
	for _, a := range args {
		if a == "--dangerously-skip-permissions" || a == "bypassPermissions" {
			t.Fatalf("claudeCLIArgs() contains %q", a)
		}
	}
}
