# GudCommit & GudChangelog - AI-Powered Git Tools

GudCommit generates clean, conventional commit messages and GudChangelog creates changelog entries using Claude. You can use **the Claude CLI** (your existing Claude app / Claude Code login), **AWS Bedrock**, or the **Anthropic API**.

## Choosing your LLM provider

| Provider | Best for | Setup |
|----------|----------|--------|
| **`claude-cli`** | Claude app / Claude Code users; no API key or AWS | Install [Claude Code](https://code.claude.com/) (includes `claude` CLI), then set `"provider": "claude-cli"` |
| **`anthropic`** | No AWS; you have an API key from console.anthropic.com | Set `ANTHROPIC_API_KEY` and `"provider": "anthropic"` |
| **`bedrock`** (default) | AWS accounts, `aws login` | `aws login --profile gudcommit` (or any AWS credentials) |

- **`claude-cli`** – Runs the `claude` CLI (`claude -p "..."`) so it uses **your existing Claude app or Claude Code login**. No API key and no separate console.anthropic.com account.
- **`anthropic`** – Calls the Anthropic API; requires an API key from [console.anthropic.com](https://console.anthropic.com/) and separate pay-as-you-go billing.
- **`bedrock`** – Uses your AWS account and supports the new `aws login` method with automatic credential rotation.

### Claude app / Claude Code vs API (why console.anthropic.com looks like a different account)

**Claude app and Claude Code** sign you in with your usual login (e.g. Google, Apple, or email). That's your **Claude.ai** account (Pro/Max/etc.).  

**API keys** come from **[console.anthropic.com](https://console.anthropic.com/)**. That's a separate product with its own sign-in and **pay-as-you-go billing** (prepaid credits). A Pro/Max subscription does **not** include API access; API usage is billed separately.

If you only use the Claude app or Claude Code, you have three options:

1. **Use the Claude CLI (recommended)** – Set `"provider": "claude-cli"`. GudCommit runs `claude -p "..."` under the hood, so it uses the same login as your Claude app / Claude Code. No API key, no console.anthropic.com. Requires the [Claude Code](https://code.claude.com/) app (which installs the `claude` command).
2. **Use Bedrock** – If you have AWS, use the `bedrock` provider and `aws login` (or other AWS credentials).
3. **Use the Anthropic API** – Sign up at [console.anthropic.com](https://console.anthropic.com/), add billing and prepaid credits, then create an API key and set `"provider": "anthropic"`.

## 🚀 Quick Start (Claude CLI – use your Claude app / Claude Code login)

If you already use the Claude app or Claude Code:

1. **Install Claude Code** (includes the `claude` CLI): [code.claude.com](https://code.claude.com/)
2. **Log in** in the app so `claude` is authenticated (e.g. run `claude` once in a terminal and complete login).
3. **Configure and run**:
```bash
# In ~/.gudcommit.json: { "provider": "claude-cli" }
# Or:
export GUD_PROVIDER=claude-cli
cd golang && make build
git add . && ./bin/gudcommit
```

No API key and no AWS needed. The tool runs `claude -p "..."` and captures the response.

## 🚀 Quick Start (Claude / Anthropic API)

If you have an API key from [console.anthropic.com](https://console.anthropic.com/) (separate from your Claude app login):

```bash
export ANTHROPIC_API_KEY=your-key   # or add to ~/.gudcommit.json
export GUD_PROVIDER=anthropic        # or "provider": "anthropic" in config
cd golang && make build
git add . && ./bin/gudcommit
```

No AWS configuration needed.

## 🚀 Quick Start (AWS Bedrock with aws login)

### Prerequisites
- AWS CLI v2.32.0 or later: `aws --version`
- Go 1.24 or later (to build from source)

### Setup

1. **Log in** (creates the `gudcommit` profile, which the tools use by default):
```bash
aws login --profile gudcommit --region us-east-1
aws sts get-caller-identity --profile gudcommit   # confirm it works
```

2. **Build and use**:
```bash
cd golang && make build
git add .
./bin/gudcommit
```

To use a profile you already log in with, set `AWS_PROFILE=<profile>` instead. Credentials refresh automatically while your login session is valid (up to 12 hours).

## 📋 Detailed Setup

See [golang/QUICKSTART_AWS_LOGIN.md](golang/QUICKSTART_AWS_LOGIN.md) for other credential sources (SSO, assumed roles, env vars) and troubleshooting.

### Alternative: Environment Variables

If you prefer not to use `aws login`, you can use traditional credentials:

```bash
export AWS_ACCESS_KEY_ID=your-key-id
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
cd golang && make build
./bin/gudcommit
```

## 🔧 Configuration

### JSON Config (recommended)

Use `~/.gudcommit.json` or `~/.gudchangelog.json`. Precedence: **env vars > ~/.gudcommit.json > ~/.gudchangelog.json > defaults**.

**Example – use Claude (Anthropic API):**
```json
{
  "provider": "anthropic",
  "model_id": "claude-opus-5-5",
  "timeout_seconds": 60
}
```
Set `ANTHROPIC_API_KEY` in your environment (recommended) or add `"anthropic_api_key": "sk-ant-..."` to the config.

**Example – use Claude CLI (your app/Code login):**
```json
{
  "provider": "claude-cli",
  "timeout_seconds": 120
}
```

**Example – use AWS Bedrock:**
```json
{
  "provider": "bedrock",
  "model_id": "us.anthropic.claude-opus-5-5",
  "timeout_seconds": 60,
  "region": "us-east-1"
}
```

**Config options:**

| Option | claude-cli | Bedrock | Anthropic | Default |
|--------|------------|---------|-----------|---------|
| `provider` | `"claude-cli"` | `"bedrock"` | `"anthropic"` | `"bedrock"` |
| `model_id` | (ignored) | Bedrock model ID or inference ARN | Anthropic model name | see above |
| `timeout_seconds` | CLI timeout | Request timeout | Request timeout | `60` (claude-cli: 120) |
| `region` | (ignored) | AWS region | (ignored) | `us-east-1` |
| `anthropic_api_key` | (ignored) | (ignored) | API key (optional if env set) | — |

For `claude-cli`, ensure the `claude` binary is on your PATH, or set `GUD_CLAUDE_CLI_PATH` to its full path.

**Environment overrides:**
- `GUD_PROVIDER` – `bedrock`, `anthropic`, or `claude-cli`
- `GUD_CLAUDE_CLI_PATH` – full path to `claude` binary (claude-cli only)
- `ANTHROPIC_API_KEY` or `GUD_ANTHROPIC_API_KEY` – for Anthropic provider
- `GUD_BEDROCK_MODEL_ID` / `GUD_ANTHROPIC_MODEL` – model ID/name
- `GUD_HTTP_TIMEOUT_SECONDS` – timeout in seconds
- `AWS_REGION`, `AWS_PROFILE` (Bedrock only; profile defaults to `gudcommit`)
- `GUD_DEBUG=1` – print config-loading diagnostics to stderr (secrets are never printed)
- `GUD_MAX_DIFF_BYTES` – largest diff that will be sent to the LLM (default `400000`; `0` disables the limit)
- `GUD_ALLOW_SECRETS=1` – send a diff even if it looks like it contains secrets, without asking (for scripts)

If you put `anthropic_api_key` in `~/.gudcommit.json`, restrict it with `chmod 600 ~/.gudcommit.json`; the tool warns when the file is group- or world-readable.

### AWS Credentials (Bedrock only)

When `provider` is `bedrock`, gudcommit uses the standard AWS SDK credential chain with the `gudcommit` profile (or `AWS_PROFILE`):
- `aws login` (recommended): short-lived credentials that refresh automatically; read natively by the SDK
- IAM Identity Center (SSO): `aws sso login`
- Assumed roles, environment variables, `~/.aws/credentials`, and IAM roles on EC2/ECS/Lambda

**Note**: API keys are **no longer needed**. The old `GUD_BEDROCK_API_KEY` method is deprecated.

## 📖 Usage

### GudCommit (Commit Messages)
```bash
# Stage your changes
git add .

# Generate commit message
./golang/bin/gudcommit
```

### GudChangelog (Changelog Entries)
```bash
# Generate changelog for changes between branches
./golang/bin/gudchangelog main

# Or compare with any other branch
./golang/bin/gudchangelog develop
```

### Discover available models/inference profiles (Bedrock only)

**Important**: Newer Claude models (like Claude 4.5 Sonnet and Claude 3.5 Sonnet v2) require using **inference profile ARNs** instead of direct model IDs.

Use the helper script to find available models and profiles:
```bash
./scripts/list-bedrock-models.sh
```

Or manually:

List models that support direct on-demand invocation:
```bash
aws bedrock list-foundation-models --region us-east-1 \
  --query "modelSummaries[?contains(inferenceTypesSupported, 'ON_DEMAND')].[modelId,providerName]" \
  --output table
```

List inference profiles (required for newer Claude models):
```bash
aws bedrock list-inference-profiles --region us-east-1 \
  --query 'inferenceProfileSummaries[].{Name:inferenceProfileName,ARN:inferenceProfileArn,Models:models[0].modelId}' \
  --output table
```

**Using newer models**: For models that require inference profiles, use the ARN as your `model_id`:
```json
{
  "model_id": "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-sonnet-4-5-v1:0",
  "timeout_seconds": 60,
  "region": "us-east-1"
}
```

### Shell Aliases (Optional)
Add to your `~/.bashrc` or `~/.zshrc`:
```bash
function gudco() {
    local commit_message="$($HOME/golang/bin/gudcommit)"
    echo "Generated commit message:"
    echo ""
    echo "\033[1m$commit_message\033[0m"
    echo ""
    echo -n "Proceed with the commit? (y/n or e to Edit): "
    read confirmation
    case "$confirmation" in
        [Yy])
            git commit -m "$commit_message"
            ;;
        [Ee])
            git commit -e -m "$commit_message"
            ;;
        *)
            echo "Commit canceled."
            ;;
    esac
}
```

## 🏗️ Architecture

### Current Implementation (Go)
- **Three providers**: **claude-cli** (Claude Code CLI), **Bedrock** (AWS), **Anthropic** (Claude API)
- **claude-cli**: Runs `claude -p "..."`; uses your Claude app / Claude Code login; no API key or AWS
- **Bedrock**: AWS SDK Go v2, SigV4, native `aws login` support, default model `us.anthropic.claude-opus-5-5` (inference profile)
- **Anthropic**: HTTPS to `api.anthropic.com`, `ANTHROPIC_API_KEY`, default model `claude-opus-5-5`
- **Conventional commit format** with structured JSON parsing
- **Config**: `~/.gudcommit.json` / env vars; `provider` chooses backend

### Benefits
✅ **Use Claude app/Claude Code** – set `provider: claude-cli` and use your existing login (no API key)  
✅ **Or use AWS Bedrock** – `aws login` and automatic credential rotation  
✅ **Or use Anthropic API** – `provider: anthropic` and `ANTHROPIC_API_KEY`  
✅ **Single config** – switch providers without changing code  
✅ **Cross-platform** (Linux, macOS, Windows)  

## 🔒 Security

- ✅ **No long-lived credentials** - Use `aws login` for short-lived tokens
- ✅ **Automatic rotation** - Credentials refresh every 15 minutes
- ✅ **Standard AWS security** - Uses IAM policies and SigV4 signing
- ✅ **Secure by default** - No API keys to accidentally commit
- ⚠️ **Never commit AWS credentials** to git

**What leaves your machine:** your staged diff (or, for `gudchangelog`, the branch diff) is sent to the configured provider. Before sending, both tools check it:
- Diffs larger than `GUD_MAX_DIFF_BYTES` (default 400 KB) are refused. Unstage lockfiles or generated files, or raise the limit.
- Sensitive files (`.env`, `*.pem`, `*.key`, SSH keys, `.npmrc`, Terraform state, ...) and secret-shaped values (AWS, Anthropic, OpenAI, GitHub, GitLab, Slack, Google and Stripe keys, private key blocks) are listed by file and line, never by value, and you're asked whether to send anyway (default: no). Removed lines are checked too, since they are sent as well.

**Prompt injection:** a diff can contain text written to manipulate the model.
- The `claude-cli` provider runs `claude -p` with all tools disabled, MCP servers off and project settings ignored, from an empty temp directory, with the prompt on stdin. Injected text can't make it run commands or edit files.
- Model output is stripped of terminal escape sequences and bidi control characters before it is displayed, committed, or written to `CHANGELOG.md`. Always read the proposed message before confirming.

**Verifying release binaries:** releases built after this change include `SHA256SUMS` and GitHub build provenance (earlier releases don't).
```bash
sha256sum -c SHA256SUMS --ignore-missing          # macOS: shasum -a 256 -c SHA256SUMS --ignore-missing
gh attestation verify gudcommit-darwin-arm64 --repo gudlyf/GudCommit
```

## 🐛 Troubleshooting

### AWS Credential Issues

If credentials fail, the error message names the profile in use and how to fix it.

#### "AWS credential error" or "authorization grant is invalid, expired, revoked, or malformed"

Your `aws login` session expired (max 12 hours). Log in again:
```bash
aws login --profile gudcommit
```

#### "unable to load AWS configuration" / "failed to get shared config profile, gudcommit"

The `gudcommit` profile doesn't exist. Create it with `aws login --profile gudcommit --region us-east-1`, or set `AWS_PROFILE` to a profile you use.

#### Wrong identity or account

Static keys take precedence over `aws login`. If `~/.aws/credentials` has a section with the same profile name, or `AWS_ACCESS_KEY_ID` is set, those are used instead. Check with `aws sts get-caller-identity --profile gudcommit`.

### Claude CLI issues (`provider: claude-cli`)

#### "claude CLI not found"

- Install [Claude Code](https://code.claude.com/) (the desktop app includes the `claude` command).
- Ensure `claude` is on your PATH (e.g. run `claude --help` in a terminal).
- Or set the full path: `export GUD_CLAUDE_CLI_PATH=/path/to/claude`

#### "claude CLI failed" or no output

- Log in once: run `claude` in a terminal and complete any sign-in flow.
- If the prompt is very long, consider using `bedrock` or `anthropic` instead (CLI may hit argument-length limits on some systems).

### Alternative: Environment Variables

**Bedrock:**
```bash
export AWS_ACCESS_KEY_ID=your-key-id
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
gudcommit
```

**Claude CLI (use app/Code login):**
```bash
export GUD_PROVIDER=claude-cli
gudcommit
```

**Anthropic API:**
```bash
export ANTHROPIC_API_KEY=your-key
export GUD_PROVIDER=anthropic
gudcommit
```

### Debug Commands
```bash
# Which AWS identity will be used
aws sts get-caller-identity --profile gudcommit

# Is the profile set up for aws login?
aws configure get login_session --profile gudcommit

# Check AWS CLI version (needs 2.32.0+ for aws login)
aws --version
```

## 📦 Releases & Downloads

### Pre-built Binaries
Download the latest release binaries for your platform:

- **Linux (amd64, arm64)**: `gudcommit-linux-*`, `gudchangelog-linux-*`
- **macOS (amd64, arm64)**: `gudcommit-darwin-*`, `gudchangelog-darwin-*`
- **Windows (amd64, arm64)**: `gudcommit-windows-*.exe`, `gudchangelog-windows-*.exe`

### Installation
```bash
# Download the appropriate binary for your platform
chmod +x gudcommit-<platform>-<arch>
chmod +x gudchangelog-<platform>-<arch>

# Move to your PATH (optional)
sudo mv gudcommit-<platform>-<arch> /usr/local/bin/gudcommit
sudo mv gudchangelog-<platform>-<arch> /usr/local/bin/gudchangelog
```

### From Source
```bash
# Clone and build
git clone <repository-url>
cd gudcommit
cd golang && make build
```

## 📚 Documentation

- **[AWS Credentials Guide](golang/QUICKSTART_AWS_LOGIN.md)** - `aws login` setup, other credential sources, troubleshooting
- **[Go Implementation](golang/GO_IMPLEMENTATION_SUMMARY.md)** - Architecture details
- [Conventional Commits](https://www.conventionalcommits.org/)
- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [AWS Login Credentials](https://docs.aws.amazon.com/sdkref/latest/guide/feature-login-credentials.html)

## 🧪 Development & Testing

### Running Tests
```bash
# Run all tests
cd golang && go test ./...

# Run tests with coverage
cd golang && go test -cover ./...

# Run benchmarks
cd golang && go test -bench=. ./...

# Run linting
cd golang && golangci-lint run
```

### Build Commands
```bash
# Build for current platform
cd golang && make build

# Build for all platforms
cd golang && make build-all

# Clean build artifacts
cd golang && make clean

# Run tests
cd golang && make test
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `cd golang && make test`
5. Run linting: `cd golang && golangci-lint run`
6. Submit a pull request

---

## 🔄 Migration from API Keys

**OLD METHOD (DEPRECATED)**: Using `GUD_BEDROCK_API_KEY` environment variable  
**NEW METHOD**: Using `aws login` with automatic credential rotation

The old API key scripts (`scripts/auto-api-key.sh`, `scripts/generate-api-key.py`) have been removed. The tool now uses standard AWS credentials with automatic rotation. If you used them, delete the leftover key file: `rm ~/.gudcommit-credentials`.

See [QUICKSTART_AWS_LOGIN.md](golang/QUICKSTART_AWS_LOGIN.md) to set up `aws login`.
