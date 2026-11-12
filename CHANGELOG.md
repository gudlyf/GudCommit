## 2.4.0

### Security
- **Config no longer leaks API keys**: removed the stderr dump of `~/.gudcommit.json` (which included `anthropic_api_key`). Diagnostics are opt-in via `GUD_DEBUG=1` and never include secrets. Warns when a key-bearing config file is group/world-readable.
- **claude-cli provider hardened against prompt injection**: runs `claude -p` with `--tools ""`, `--strict-mcp-config`, `--setting-sources user` and `--no-session-persistence`, from an empty temp directory, with the prompt on stdin (no longer visible in `ps` or limited by ARG_MAX). CLI stderr no longer leaks into commit messages.
- **Outgoing diff guard**: diffs are checked before being sent to any provider. Oversized diffs are refused (`GUD_MAX_DIFF_BYTES`, default 400 KB); sensitive files and secret-shaped values are reported by file/line and require confirmation (`GUD_ALLOW_SECRETS=1` to skip).
- **Model output sanitized**: terminal escape sequences, control characters and bidi overrides are stripped before display, commit, or writing `CHANGELOG.md`.
- **gudchangelog option injection fixed**: the target branch is validated and resolved to a commit SHA, so values like `--output=...` can't be passed to `git diff`.
- Anthropic API responses are size-limited; `parser.GenerateRandomString` uses `crypto/rand`.
- Release binaries are built with `-trimpath`, ship with `SHA256SUMS`, and carry GitHub build provenance attestations.
- CI: Go 1.27 (1.25 is out of security support), least-privilege `GITHUB_TOKEN`, SHA-pinned actions, no persisted checkout credentials, `govulncheck` gate, `gosec` linting, Dependabot, pinned GitLab images/tools and remote include.

### Removed
- `scripts/auto-api-key.sh`, `scripts/generate-api-key.py` and `scripts/requirements.txt` (deprecated Bedrock API key flow). Delete any leftover `~/.gudcommit-credentials`.
- Prebuilt binaries `golang/gudcommit` and `golang/gudchangelog` are no longer tracked in git; use release downloads or `make build`.

### Changed
- golangci-lint upgraded to v2 (config migrated); required for Go 1.27.

---

## 2.3.0

### Added
- **Multiple LLM providers**: Choose backend via `provider` in config or `GUD_PROVIDER` env:
  - **`claude-cli`** – Runs the `claude` CLI (`claude -p "..."`); uses your existing Claude app / Claude Code login; no API key or AWS. Requires [Claude Code](https://code.claude.com/) on PATH (or `GUD_CLAUDE_CLI_PATH`).
  - **`anthropic`** – Anthropic API with `ANTHROPIC_API_KEY` (or `anthropic_api_key` in config); same models as Claude app, separate billing at console.anthropic.com.
  - **`bedrock`** (default) – AWS Bedrock with `aws login` or other AWS credentials.
- Config aliases: `provider: "claude"` or `"claude_cli"` normalize to `claude-cli`; `"anthropic-api"` / `"api"` to `anthropic`.
- **claude-cli**: Minimum timeout 90 seconds so shared configs with short timeouts don't kill the CLI before it responds.

### Changed
- **Config**: New `provider` and optional `anthropic_api_key`; `LoadConfig()` is now exported for use by provider factory.
- **Invoker**: `NewInvoker()` returns the configured provider (Bedrock, Anthropic, or Claude CLI) based on config.
- README: Provider comparison table, Quick Start for Claude CLI and Anthropic API, "Claude app vs API" explanation, troubleshooting for claude-cli.

---

## 2.2.1

### Added
- Go implementation of GudCommit tool
- Support for cross-platform builds
- Comprehensive test suite
- **NEW**: Support for `aws login` credentials (AWS CLI v2.32.0+) with automatic 15-minute token rotation
- Support for multiple AWS credential providers via `credential_process`, environment variables, shared credentials, legacy SSO, and IAM roles
- **Automatic profile defaulting**: gudcommit/gudchangelog now default to using `gudcommit` profile (no need to set `AWS_PROFILE`)
- **Helpful error messages**: Automatic display of setup instructions when credentials fail, including:
  - How to configure `~/.aws/config` with `credential_process`
  - Step-by-step `aws login` setup guide
  - Alternative credential methods
  - Troubleshooting for common errors (IMDS, credential_process failures, etc.)
- Quick start guide: `golang/QUICKSTART_AWS_LOGIN.md`
- Comprehensive AWS login documentation: `golang/AWS_SSO_SUPPORT.md`

### Changed
- Migrated from JavaScript to Go for better performance and portability
- Improved error handling and fallback mechanisms
- **BREAKING**: Migrated from custom API key authentication to AWS SDK v2 with SigV4 signing
- **DEPRECATED**: `GUD_BEDROCK_API_KEY` environment variable and API key generation scripts are no longer needed
- Replaced `GUD_BEDROCK_API_KEY` with standard AWS credential chain
- Updated to use AWS SDK Go v2 (v1.30.5)
- Enhanced credential error messages with helpful troubleshooting guidance
- Added `credential_process` support to enable new `aws login` method (workaround for Go SDK v2 limitation)
- Updated main README with new `aws login` instructions and deprecated API key methods
- Client now defaults to `gudcommit` profile if `AWS_PROFILE` not set
- Added debug logging for configuration file loading to help troubleshoot config issues
- Added helper script `scripts/list-bedrock-models.sh` to discover available models and inference profiles
- Added documentation about using inference profile ARNs for newer Claude models

### Removed
- **DEPRECATED**: Old API key generation scripts (`scripts/auto-api-key.sh`, `scripts/generate-api-key.py`) are no longer needed

---

### Added
- Added `ora` package for terminal spinner/progress indicator
- Added `spinner` parameter to `invokeBedrockAgent` function to display spinner while waiting for Bedrock response
- Added error handling with spinner to indicate errors while waiting for Bedrock response

### Changed
- Updated AWS provider version in Terraform configuration from 5.55.0 to 5.57.0
- Added `aws_profile` variable in Terraform module to specify AWS profile for Bedrock
- Updated `prepare_gudcommit` and `prepare_gudchangelog` resources to use `aws_profile` variable when preparing agents

The changes were made to improve the user experience by providing visual feedback while waiting for Bedrock's response, and to enhance error handling. Additionally, the AWS provider version was updated, and a new variable was added to the Terraform module to allow specifying the AWS profile for Bedrock.

---

## Added

- Support for the `hashicorp/null` provider version 3.2.2.
- New `null_resource` blocks `prepare_gudcommit` and `prepare_gudchangelog` to execute AWS CLI commands for preparing the respective agents after their creation.

## Changed

- Renamed output variables in `dev/outputs.tf` and `module/outputs.tf` to include prefixes `gudcommit_` and `gudchangelog_` for better clarity and organization.
- Updated dependencies for the new `null_resource` blocks to ensure they run after the agent resources are created.

The changes were made to improve the organization and maintainability of the Terraform configuration. The addition of the `null_resource` blocks allows for executing additional commands after the agent resources are created, potentially for further setup or configuration. The renaming of output variables provides better clarity and separation between the different agents in the project.
