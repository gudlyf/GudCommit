package bedrock

import (
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

// InvokeModel runs `claude -p "<prompt>"` and returns the printed response.
func (c *ClaudeCLIClient) InvokeModel(prompt, _ string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.claudePath, "-p", prompt)
	cmd.Stdin = nil

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

	out, err := cmd.CombinedOutput()
	close(done)
	fmt.Fprint(os.Stderr, "\r\033[K")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude CLI timed out after %d seconds: %w", c.timeoutSec, err)
		}
		return "", fmt.Errorf("claude CLI failed: %w\nOutput: %s", err, string(out))
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", fmt.Errorf("claude CLI produced no output")
	}
	fmt.Fprint(os.Stderr, "✔ :: Response received from Claude CLI\n")
	return text, nil
}
