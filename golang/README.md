# GudCommit - Go Version

A Go implementation of the GudCommit tool for generating clean, conventional commit messages using Claude via **AWS Bedrock** or the **Anthropic API** (Claude app / Claude Code).

## Features

- **Clean Output**: Generates conventional commit messages without extraneous text
- **Dual provider**: **Bedrock** (AWS) or **Anthropic** (Claude API – same backend as Claude app and Claude Code)
- **Git Integration**: Automatically detects staged changes
- **Structured Response Parsing**: Handles both JSON and conventional commit formats
- **Cross-Platform**: Builds for Linux, macOS, and Windows
- **Fast**: Compiled Go binary with minimal dependencies

## Prerequisites

- Go 1.24 or later
- **For Anthropic (Claude app/Claude Code)**: `ANTHROPIC_API_KEY` and `GUD_PROVIDER=anthropic` (or `"provider": "anthropic"` in config)
- **For Bedrock**: AWS credentials (`aws login` recommended; any AWS SDK credential source works)
- Git repository with staged changes

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd gudcommit/golang

# Download dependencies
make deps

# Build binaries
make build

# Install to system (optional)
make install
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all
```

## Usage

### GudCommit (Commit Messages)

```bash
# Generate commit messages for staged changes
./bin/gudcommit

# Or if installed system-wide
gudcommit
```

### GudChangelog (Changelog Generation)

```bash
# Generate changelog comparing to a branch
./bin/gudchangelog develop

# Or if installed system-wide
gudchangelog develop
```

## Shell Integration

Add these functions to your `~/.bashrc` or `~/.zshrc`:

```bash
function gudco() {
    local commit_message="$(gudcommit)"
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

function gudcl() {
    local branch=$1
    if [[ ! "$branch" ]]; then
        echo ">> Must specify a branch to compare to as argument"
        return 1
    fi
    local changelog_message="$(gudchangelog $branch)"
    if [[ ! "$changelog_message" ]]; then
        echo ">> No message generated."
        return 1
    fi
    echo "Generated CHANGELOG.md message:"
    echo ""
    echo "\033[1m$changelog_message\033[0m"
    echo ""
    echo -n "Prepend this content to CHANGELOG.md? (y/n): "
    read confirmation
    case "$confirmation" in
        [Yy])
            local top_level="$(git rev-parse --show-toplevel)"
            echo "$changelog_message" > /tmp/gudchangelog.md
            if [[ -f "$top_level/CHANGELOG.md" ]]; then
                cat "$top_level/CHANGELOG.md" >> /tmp/gudchangelog.md
                echo "---" >> /tmp/gudchangelog.md
            fi
            mv /tmp/gudchangelog.md "$top_level/CHANGELOG.md"
            ;;
        *)
            echo "Commit canceled."
            ;;
    esac
}
```

## Development

### Project Structure

```
golang/
├── main.go              # Main gudcommit binary
├── gudchangelog.go      # Changelog generation binary
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation
└── README.md           # This file
```

### Available Make Targets

- `make build` - Build both binaries
- `make gudcommit` - Build gudcommit binary only
- `make gudchangelog` - Build gudchangelog binary only
- `make install` - Install binaries to /usr/local/bin
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make deps` - Download dependencies
- `make fmt` - Format code
- `make lint` - Lint code
- `make build-all` - Build for multiple platforms
- `make help` - Show help

### Dependencies

- `github.com/aws/aws-sdk-go-v2` - AWS SDK v2 (with SSO support)
- `github.com/aws/aws-sdk-go-v2/config` - AWS configuration loader
- `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` - Bedrock Runtime service
- `github.com/aws/smithy-go` - Smithy Go SDK primitives

## Configuration

### AWS Credentials

The tool uses the standard AWS SDK credential chain, which reads `aws login` sessions natively. It uses the `gudcommit` profile unless `AWS_PROFILE` is set.

```bash
aws login --profile gudcommit --region us-east-1   # AWS CLI v2.32.0+
gudcommit
```

Credentials refresh automatically while the login session is valid (up to 12 hours). IAM Identity Center (`aws sso login`), assumed roles, environment variables and IAM roles also work.

See [QUICKSTART_AWS_LOGIN.md](./QUICKSTART_AWS_LOGIN.md) for other credential sources and troubleshooting.

### Environment Variables

- `AWS_REGION` (optional): AWS region (defaults to us-east-1)
- `GUD_BEDROCK_MODEL_ID` (optional): Override the default model ID
- `GUD_HTTP_TIMEOUT_SECONDS` (optional): Override the default timeout (60 seconds)
- `GUD_DEBUG` (optional): Set to `1` to print config-loading diagnostics to stderr
- `GUD_MAX_DIFF_BYTES` (optional): Largest diff sent to the LLM (default 400000; 0 disables)
- `GUD_ALLOW_SECRETS` (optional): Set to `1` to send diffs flagged as containing secrets without prompting

### JSON Configuration (Optional)

Create `~/.gudcommit.json` for custom settings:

```json
{
  "model_id": "anthropic.claude-3-5-sonnet-20240620-v1:0",
  "timeout_seconds": 60,
  "region": "us-east-1"
}
```

## Error Handling

The Go version includes robust error handling:

- **AWS Authentication**: Clear error messages for expired tokens
- **Git Operations**: Handles missing repositories and branches gracefully
- **Response Parsing**: Falls back to conventional commit format if JSON parsing fails
- **Empty Diffs**: Handles cases with no staged changes

## Performance

The Go version offers several performance advantages:

- **Compiled Binary**: No runtime dependencies
- **Fast Startup**: Minimal initialization time
- **Memory Efficient**: Lower memory footprint than Node.js
- **Concurrent Processing**: Built-in goroutine support for future enhancements

## Differences from JavaScript Version

1. **Single Binary**: No need for Node.js or npm
2. **Faster Execution**: Compiled Go code runs faster than interpreted JavaScript
3. **Better Error Messages**: More descriptive error handling
4. **Cross-Platform**: Easy to build for multiple platforms
5. **Static Linking**: No external dependencies at runtime

## Troubleshooting

### Common Issues

1. **AWS Authentication**: 
   - **`aws login`**: Ensure AWS CLI v2.32.0+ and run `aws login --profile gudcommit` (sessions last up to 12 hours)
   - **Environment variables**: Ensure `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are set
   - **Legacy AWS SSO**: Ensure you're logged in: `aws sso login --profile your-profile`
   - Check that your credentials have access to AWS Bedrock
2. **Git Repository**: Make sure you're in a git repository with staged changes
3. **Region**: Verify that AWS Bedrock is available in your configured region (defaults to us-east-1)
4. **Wrong identity**: static keys in `~/.aws/credentials` or `AWS_ACCESS_KEY_ID` take precedence over `aws login`; check with `aws sts get-caller-identity --profile gudcommit`

### Debug Mode

To print config-loading diagnostics (secrets are never printed), set `GUD_DEBUG`:

```bash
GUD_DEBUG=1 ./bin/gudcommit
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Submit a pull request

## License

Same license as the main GudCommit project.
