package bedrock

import "fmt"

// Default request size. Current Claude models think before answering, and
// thinking counts toward max_tokens, so leave room beyond the reply itself.
const defaultMaxTokens = 16000

// extractText returns the first non-empty text block of a Messages API
// response. Blocks are matched by type, not position: responses from current
// models can begin with (empty) thinking blocks.
func extractText(provider, stopReason string, blocks []ContentBlock) (string, error) {
	switch stopReason {
	case "refusal":
		return "", fmt.Errorf("%s declined to process this diff (stop_reason: refusal); try a smaller or different set of changes", provider)
	case "max_tokens":
		return "", fmt.Errorf("%s response was cut off at max_tokens; try staging fewer changes", provider)
	}
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			return b.Text, nil
		}
	}
	return "", fmt.Errorf("%s returned no text content (stop_reason: %q)", provider, stopReason)
}
