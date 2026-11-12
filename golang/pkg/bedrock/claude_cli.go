package bedrock

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ClaudeCLIClient invokes the Claude Code CLI (claude -p) so you can use your
// existing Claude app / Claude Code login—no API key or AWS required.
type ClaudeCLIClient struct {
	claudePath string
	timeoutSec int
}

// NewClaudeCLIClient creates a client that runs the `claude` CLI with -p.
// Requires the Claude Code CLI to be installed and on PATH (or GUD_CLAUDE_CLI_PATH set).
func NewClaudeCLIClient(cfg Config) (*ClaudeCLIClient, error) {
	path := os.Getenv("GUD_CLAUDE_CLI_PATH")
	if path == "" {
		var err error
		path, err = exec.LookPath("claude")
		if err != nil {
			return nil, fmt.Errorf(`claude CLI not found. Provider "claude-cli" requires the Claude Code CLI.

Install Claude Code (includes the CLI): https://code.claude.com/
Ensure "claude" is on your PATH, or set GUD_CLAUDE_CLI_PATH to the full path of the claude binary.

Then use: "provider": "claude-cli" in ~/.gudcommit.json or GUD_PROVIDER=claude-cli`)
		}
	}
	// CLI needs time for process startup, auth, and inference; enforce a minimum
	const minTimeoutSec = 90
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 120
	}
	if timeout < minTimeoutSec {
		timeout = minTimeoutSec
	}
	return &ClaudeCLIClient{claudePath: path, timeoutSec: timeout}, nil
}

// claudeCLIArgs returns the arguments used to run the Claude CLI as a plain,
// tool-less text generator. The git diff in the prompt is untrusted input, so the
// session must not be able to run commands, edit files, or reach MCP servers.
// Note: --tools is variadic, so it is kept before other flags with explicit values.
func claudeCLIArgs() []string {
	return []string{
		"-p",
		"--tools", "",
		"--strict-mcp-config",
		"--no-session-persistence",
		"--setting-sources", "user",
		"--output-format", "text",
	}
}

// InvokeModel runs the Claude CLI in print mode, sending the prompt on stdin, and
// returns the printed response.
func (c *ClaudeCLIClient) InvokeModel(prompt, _ string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutSec)*time.Second)
	defer cancel()

	// Run from a neutral, empty directory so the CLI does not pick up the
	// repository's CLAUDE.md, .claude/settings.json, hooks, or .mcp.json.
	workDir, err := os.MkdirTemp("", "gudcommit-claude-")
	if err != nil {
		return "", fmt.Errorf("claude CLI: create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Prompt goes on stdin rather than argv: argv is visible to other local users
	// via ps and is limited by ARG_MAX, which large diffs exceed.
	// claudePath comes from the user's own PATH or GUD_CLAUDE_CLI_PATH; args are constant.
	cmd := exec.CommandContext(ctx, c.claudePath, claudeCLIArgs()...) //nolint:gosec // G204: see above
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Spinner while CLI runs
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})
	start := time.Now()
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(start)
				fmt.Fprintf(os.Stderr, "\r\033[K%s :: Awaiting response from Claude CLI ... [%ds]", spinnerChars[i%len(spinnerChars)], int(elapsed.Seconds()))
				os.Stderr.Sync()
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	err = cmd.Run()
	close(done)
	fmt.Fprint(os.Stderr, "\r\033[K")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude CLI timed out after %d seconds: %w", c.timeoutSec, err)
		}
		return "", fmt.Errorf("claude CLI failed: %w\nOutput: %s", err, strings.TrimSpace(stderr.String()))
	}

	text := strings.TrimSpace(stdout.String())
	if text == "" {
		return "", fmt.Errorf("claude CLI produced no output")
	}
	fmt.Fprint(os.Stderr, "✔ :: Response received from Claude CLI\n")
	return text, nil
}
