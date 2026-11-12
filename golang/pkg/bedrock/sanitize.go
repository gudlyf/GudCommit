package bedrock

import (
	"strings"
	"unicode"
)

// SanitizeOutput strips characters from model output that could manipulate a
// terminal or hide content in commit messages and changelogs: ANSI/OSC escape
// sequences and other C0/C1 control characters (newline and tab are kept), and
// Unicode bidirectional override/isolate characters ("Trojan Source").
// Model output is derived from untrusted input (the git diff), so it is treated
// as untrusted too.
func SanitizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case unicode.IsControl(r):
			// drops ESC (0x1b), CR, BEL, DEL, C1 controls, etc.
		case isBidiControl(r):
		case r == unicode.ReplacementChar:
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isBidiControl(r rune) bool {
	return (r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069) || r == 0x200E || r == 0x200F || r == 0x061C
}

// sanitizingInvoker wraps an Invoker and sanitizes every response.
type sanitizingInvoker struct {
	inner Invoker
}

func (s sanitizingInvoker) InvokeModel(prompt, repoPath string) (string, error) {
	out, err := s.inner.InvokeModel(prompt, repoPath)
	return SanitizeOutput(out), err
}
