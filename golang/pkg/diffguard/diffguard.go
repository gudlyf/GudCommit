// Package diffguard inspects a git diff before it is sent to an LLM provider.
// It flags likely secrets and sensitive files, and enforces a size limit, so
// that credentials are not shipped to a third party by accident.
package diffguard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// DefaultMaxBytes is the default diff size limit (~100K tokens).
const DefaultMaxBytes = 400_000

// Finding is a single suspected secret or sensitive file in a diff.
// It never carries the matched value, so it is safe to print.
type Finding struct {
	File    string
	Line    int  // 0 when the finding is about the file itself
	Removed bool // true when the match is on a removed ("-") line
	Rule    string
}

func (f Finding) String() string {
	if f.Line == 0 {
		return fmt.Sprintf("%s — %s", f.File, f.Rule)
	}
	side := "added"
	if f.Removed {
		side = "removed"
	}
	return fmt.Sprintf("%s:%d (%s line) — %s", f.File, f.Line, side, f.Rule)
}

type rule struct {
	name string
	re   *regexp.Regexp
}

// Order matters: the first matching rule wins for a given line.
var contentRules = []rule{
	{"private key", regexp.MustCompile(`-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY(?: BLOCK)?-----`)},
	{"AWS access key ID", regexp.MustCompile(`\b(?:AKIA|ASIA|ABIA|ACCA)[0-9A-Z]{16}\b`)},
	{"AWS secret access key", regexp.MustCompile(`(?i)aws_?secret_?access_?key["'\s:=]+[A-Za-z0-9/+]{40}\b`)},
	{"Amazon Bedrock API key", regexp.MustCompile(`\b(?:ABSK[A-Za-z0-9+/=]{40,}|bedrock-api-key-[A-Za-z0-9+/=._-]{20,})`)},
	{"Anthropic API key", regexp.MustCompile(`\bsk-ant-[A-Za-z0-9_-]{20,}`)},
	{"OpenAI API key", regexp.MustCompile(`\bsk-(?:proj-|svcacct-|admin-)?[A-Za-z0-9_-]{32,}`)},
	{"GitHub token", regexp.MustCompile(`\b(?:(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36,}|github_pat_[A-Za-z0-9_]{22,})\b`)},
	{"GitLab token", regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{20,}`)},
	{"Slack token", regexp.MustCompile(`\bxox[abposr]-[A-Za-z0-9-]{10,}`)},
	{"Google API key", regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	{"Stripe live key", regexp.MustCompile(`\b[rs]k_live_[0-9A-Za-z]{20,}\b`)},
}

// placeholder matches obvious documentation/sample values that should not be flagged.
var placeholder = regexp.MustCompile(`(?i)example|x{6,}|placeholder|dummy|your[-_]?(?:key|token|secret)|redacted|\.\.\.`)

var sensitiveNames = map[string]bool{
	".env": true, ".netrc": true, ".npmrc": true, ".pypirc": true, ".pgpass": true,
	".git-credentials": true, ".htpasswd": true,
	"id_rsa": true, "id_dsa": true, "id_ecdsa": true, "id_ed25519": true,
	"terraform.tfstate": true, "terraform.tfstate.backup": true,
}

var sensitiveExts = []string{".pem", ".key", ".p12", ".pfx", ".jks", ".keystore", ".kdbx", ".tfstate", ".ppk"}

var safeEnvSuffixes = []string{".example", ".sample", ".template", ".dist", ".defaults"}

// sensitiveFile reports why a path looks like it holds secrets, or "".
func sensitiveFile(p string) string {
	base := strings.ToLower(path.Base(p))
	lp := strings.ToLower(p)
	if sensitiveNames[base] {
		return "sensitive file name"
	}
	if strings.HasSuffix(lp, ".aws/credentials") || strings.HasSuffix(lp, ".docker/config.json") {
		return "sensitive file name"
	}
	if strings.HasPrefix(base, ".env.") {
		for _, s := range safeEnvSuffixes {
			if strings.HasSuffix(base, s) {
				return ""
			}
		}
		return "sensitive file name"
	}
	for _, ext := range sensitiveExts {
		if strings.HasSuffix(base, ext) {
			return "sensitive file type (" + ext + ")"
		}
	}
	return ""
}

var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// Scan returns suspected secrets and sensitive files in a unified git diff.
func Scan(diff string) []Finding {
	var findings []Finding
	var file string
	var oldLine, newLine int
	inHunk := false

	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			file = parseDiffGitPath(line)
			inHunk = false
			if why := sensitiveFile(file); why != "" {
				findings = append(findings, Finding{File: file, Rule: why})
			}
			continue
		case !inHunk && (strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ")):
			continue
		}
		if m := hunkHeader.FindStringSubmatch(line); m != nil {
			oldLine, _ = strconv.Atoi(m[1])
			newLine, _ = strconv.Atoi(m[2])
			inHunk = true
			continue
		}
		if !inHunk || line == "" {
			continue
		}

		var lineNo int
		removed := false
		switch line[0] {
		case '+':
			lineNo = newLine
			newLine++
		case '-':
			lineNo = oldLine
			oldLine++
			removed = true
		case ' ':
			oldLine++
			newLine++
			continue
		default: // "\ No newline at end of file", etc.
			continue
		}

		content := line[1:]
		for _, r := range contentRules {
			if m := r.re.FindString(content); m != "" && !placeholder.MatchString(m) {
				findings = append(findings, Finding{File: file, Line: lineNo, Removed: removed, Rule: r.name})
				break
			}
		}
	}
	return findings
}

// parseDiffGitPath extracts the post-image path from a "diff --git a/X b/Y" header.
func parseDiffGitPath(header string) string {
	rest := strings.TrimPrefix(header, "diff --git ")
	if i := strings.LastIndex(rest, " b/"); i >= 0 {
		return strings.Trim(rest[i+3:], `"`)
	}
	if i := strings.LastIndex(rest, ` "b/`); i >= 0 {
		return strings.Trim(rest[i+4:], `"`)
	}
	return rest
}

// MaxBytes returns the configured diff size limit: GUD_MAX_DIFF_BYTES if set
// (0 disables the limit), otherwise DefaultMaxBytes.
func MaxBytes() int {
	if v := strings.TrimSpace(os.Getenv("GUD_MAX_DIFF_BYTES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return DefaultMaxBytes
}

func allowSecrets() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GUD_ALLOW_SECRETS"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// Enforce checks the diff before it leaves the machine. It returns true when
// the diff may be sent. Oversized diffs are refused outright. When suspected
// secrets are found, the findings are printed (without their values) and the
// user is asked to confirm, defaulting to no; GUD_ALLOW_SECRETS=1 skips the
// prompt for non-interactive use.
func Enforce(diff string, in *bufio.Reader, out io.Writer) (bool, error) {
	if limit := MaxBytes(); limit > 0 && len(diff) > limit {
		fmt.Fprintf(out, "❌ Diff is %d bytes, over the %d-byte limit for sending to the LLM.\n", len(diff), limit)
		fmt.Fprintln(out, "   Stage fewer files (lockfiles and generated files are common culprits),")
		fmt.Fprintln(out, "   or raise the limit with GUD_MAX_DIFF_BYTES (0 disables it).")
		return false, nil
	}

	findings := Scan(diff)
	if len(findings) == 0 {
		return true, nil
	}

	fmt.Fprintln(out, "⚠️  The diff appears to contain secrets or sensitive files:")
	for _, f := range findings {
		fmt.Fprintf(out, "   - %s\n", f)
	}
	fmt.Fprintln(out, "   This diff would be sent to your configured LLM provider.")

	if allowSecrets() {
		fmt.Fprintln(out, "   GUD_ALLOW_SECRETS is set; sending anyway.")
		return true, nil
	}

	fmt.Fprint(out, "Send it anyway? (y/N): ")
	answer, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	}
	fmt.Fprintln(out, "Aborted. Unstage the flagged files or remove the secrets, then try again.")
	return false, nil
}
