package diffguard

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// Fake credentials are assembled at runtime so this file does not itself
// contain anything that looks like a secret to scanners (including this one).
var (
	fakeAWSKey       = "AKIA" + "Q3VZ7K2M4N6P8R0T"
	fakeAnthropicKey = "sk-" + "ant-api03-" + "q8Zr2LmN4vB6xC1dF3gH5jK7"
	fakeGitHubToken  = "ghp" + "_" + "a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8"
	fakePrivateKey   = "-----BEGIN " + "RSA PRIVATE KEY-----"
)

func diffFor(path string, lines ...string) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + path + " b/" + path + "\n")
	b.WriteString("index 1111111..2222222 100644\n")
	b.WriteString("--- a/" + path + "\n")
	b.WriteString("+++ b/" + path + "\n")
	b.WriteString("@@ -10,3 +20,4 @@ func x() {\n")
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	return b.String()
}

func TestScanFindsSecretsWithLineNumbers(t *testing.T) {
	diff := diffFor("app/config.go",
		" context line",                     // old 10, new 20
		"-old := \""+fakeGitHubToken+"\"",   // old 11 (removed)
		"+key := \""+fakeAWSKey+"\"",        // new 21
		"+ok := true",                       // new 22
		"+anth := \""+fakeAnthropicKey+"\"", // new 23
		" context",
	)
	got := Scan(diff)
	want := []Finding{
		{File: "app/config.go", Line: 11, Removed: true, Rule: "GitHub token"},
		{File: "app/config.go", Line: 21, Rule: "AWS access key ID"},
		{File: "app/config.go", Line: 23, Rule: "Anthropic API key"},
	}
	if len(got) != len(want) {
		t.Fatalf("Scan() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("finding %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestScanPrivateKey(t *testing.T) {
	got := Scan(diffFor("deploy/key.txt", "+"+fakePrivateKey))
	if len(got) != 1 || got[0].Rule != "private key" {
		t.Fatalf("Scan() = %v, want one private key finding", got)
	}
}

func TestScanIgnoresPlaceholders(t *testing.T) {
	diff := diffFor("README.md",
		"+export ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx",
		"+aws_access_key_id = AKIAIOSFODNN7EXAMPLE",
		"+plain text line",
	)
	if got := Scan(diff); len(got) != 0 {
		t.Fatalf("Scan() = %v, want no findings", got)
	}
}

func TestSensitiveFiles(t *testing.T) {
	flagged := []string{".env", "config/.env.production", "certs/server.pem", "tls.key",
		"home/.aws/credentials", "infra/terraform.tfstate", "id_ed25519", ".npmrc"}
	for _, p := range flagged {
		if sensitiveFile(p) == "" {
			t.Errorf("sensitiveFile(%q) = \"\", want flagged", p)
		}
	}
	clean := []string{".env.example", ".env.sample", "main.go", "keyboard.go", "docs/keys.md", "go.sum"}
	for _, p := range clean {
		if why := sensitiveFile(p); why != "" {
			t.Errorf("sensitiveFile(%q) = %q, want not flagged", p, why)
		}
	}
	got := Scan(diffFor(".env", "+DEBUG=1"))
	if len(got) != 1 || got[0].Line != 0 || got[0].File != ".env" {
		t.Fatalf("Scan(.env diff) = %v, want one file-level finding", got)
	}
}

func TestFindingStringOmitsValue(t *testing.T) {
	for _, f := range Scan(diffFor("a.go", "+k := \""+fakeAWSKey+"\"")) {
		if strings.Contains(f.String(), fakeAWSKey) {
			t.Fatalf("Finding.String() leaks the secret: %s", f)
		}
	}
}

func TestEnforce(t *testing.T) {
	secretDiff := diffFor("a.go", "+k := \""+fakeAWSKey+"\"")
	cleanDiff := diffFor("a.go", "+x := 1")

	run := func(diff, input string) (bool, string) {
		var out bytes.Buffer
		ok, err := Enforce(diff, bufio.NewReader(strings.NewReader(input)), &out)
		if err != nil {
			t.Fatalf("Enforce error: %v", err)
		}
		return ok, out.String()
	}

	t.Setenv("GUD_ALLOW_SECRETS", "")
	t.Setenv("GUD_MAX_DIFF_BYTES", "")

	if ok, out := run(cleanDiff, ""); !ok || out != "" {
		t.Errorf("clean diff: ok=%v out=%q, want ok and silent", ok, out)
	}
	if ok, out := run(secretDiff, "y\n"); !ok || strings.Contains(out, fakeAWSKey) {
		t.Errorf("secret diff + y: ok=%v, leaked=%v", ok, strings.Contains(out, fakeAWSKey))
	}
	if ok, _ := run(secretDiff, "\n"); ok {
		t.Error("secret diff + enter: want refused (default no)")
	}
	if ok, _ := run(secretDiff, ""); ok {
		t.Error("secret diff + EOF: want refused")
	}

	t.Setenv("GUD_ALLOW_SECRETS", "1")
	if ok, _ := run(secretDiff, ""); !ok {
		t.Error("GUD_ALLOW_SECRETS=1: want allowed without prompt")
	}

	t.Setenv("GUD_MAX_DIFF_BYTES", "50")
	if ok, out := run(cleanDiff, "y\n"); ok || !strings.Contains(out, "limit") {
		t.Errorf("oversized diff: ok=%v out=%q, want refused", ok, out)
	}
	t.Setenv("GUD_MAX_DIFF_BYTES", "0")
	if ok, _ := run(cleanDiff, ""); !ok {
		t.Error("GUD_MAX_DIFF_BYTES=0: want limit disabled")
	}
}
