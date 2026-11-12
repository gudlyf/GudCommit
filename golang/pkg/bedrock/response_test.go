package bedrock

import (
	"strings"
	"testing"
)

func TestExtractText(t *testing.T) {
	thinkingFirst := []ContentBlock{{Type: "thinking"}, {Type: "text", Text: "feat: x"}}
	if got, err := extractText("Bedrock", "end_turn", thinkingFirst); err != nil || got != "feat: x" {
		t.Fatalf("thinking-first: got %q, %v; want text block", got, err)
	}
	if _, err := extractText("Bedrock", "end_turn", []ContentBlock{{Type: "thinking"}}); err == nil {
		t.Fatal("no text block: want error")
	}
	for _, sr := range []string{"refusal", "max_tokens"} {
		_, err := extractText("Claude", sr, []ContentBlock{{Type: "text", Text: "partial"}})
		if err == nil || !strings.Contains(err.Error(), sr) {
			t.Fatalf("stop_reason %s: got %v, want error mentioning it", sr, err)
		}
	}
}
