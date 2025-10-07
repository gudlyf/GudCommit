# GudCommit - Go Version

A Go implementation of the GudCommit tool for generating clean, conventional commit messages using AWS Bedrock.

## Features

- **Clean Output**: Generates conventional commit messages without extraneous text
- **AWS Bedrock Integration**: Uses Claude Sonnet for intelligent commit message generation
- **Git Integration**: Automatically detects staged changes
- **Structured Response Parsing**: Handles both JSON and conventional commit formats
- **Cross-Platform**: Builds for Linux, macOS, and Windows
- **Fast**: Compiled Go binary with minimal dependencies

## Prerequisites

- Go 1.21 or later
- AWS credentials configured (via AWS CLI, environment variables, or IAM roles)
- Git repository with staged changes
- Bedrock API key (generated via console or `scripts/auto-api-key.sh`)

## Installation

### From Source

```bash
# Clone the repository
git clone git@github.com:gudlyf/GudCommit.git
cd GudCommit/golang

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

- `github.com/aws/aws-sdk-go-v2` - AWS SDK v2
- `github.com/go-git/go-git/v5` - Git operations
- `github.com/schollz/progressbar/v3` - Progress indicators

## Configuration

### Environment Variables

- `GUD_BEDROCK_API_KEY` (required): Your Bedrock API key
- `AWS_REGION` (optional): AWS region (defaults to us-east-1)

### API Key Generation

Generate a Bedrock API key using the automated script:

```bash
# Generate API key using AWS Bedrock token generator
./scripts/auto-api-key.sh generate

# Or manually from AWS Console
# https://us-east-1.console.aws.amazon.com/bedrock/home?region=us-east-1#/api-keys
```

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

1. **AWS Authentication**: Ensure your AWS credentials are configured (via `aws configure` or environment variables)
2. **Git Repository**: Make sure you're in a git repository with staged changes
3. **SSM Parameters**: Verify that the required SSM parameters exist in your AWS account

### Debug Mode

To enable debug output, set the `DEBUG` environment variable:

```bash
DEBUG=1 ./bin/gudcommit
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
