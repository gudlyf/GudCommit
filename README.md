# GudCommit & GudChangelog - AI-Powered Git Tools

GudCommit generates clean, conventional commit messages and GudChangelog creates changelog entries using AWS Bedrock AI models. Now with **API key authentication** - no IAM users or AWS profiles required!

## 🚀 Quick Start (Automatic API Key Method)

### Option 1: Automatic API Key Management (Recommended)
```bash
# One-time setup (if not already done)
./scripts/auto-api-key.sh setup

# Use GudCommit (automatically manages API keys)
./scripts/auto-api-key.sh run

# Use GudChangelog (automatically manages API keys)
./scripts/auto-api-key.sh run gudchangelog main
```

### Option 2: Manual API Key Method
1. Go to [AWS Bedrock Console](https://us-east-1.console.aws.amazon.com/bedrock/home?region=us-east-1#/api-keys)
2. Click "Create API Key" → "Short-term API key"
3. Set environment variable: `export GUD_BEDROCK_API_KEY="your-key"`
4. Build and use: `cd golang && make build && ./bin/gudcommit` or `./bin/gudchangelog main`

**That's it!** No AWS profiles, no IAM users, no complex setup.

## 📋 Detailed Setup

### Option 1: Automated Setup Script
```bash
# Run the interactive setup script
./scripts/auto-api-key.sh setup
```

### Option 2: Manual Setup
1. **Generate API Key**: [Bedrock Console](https://us-east-1.console.aws.amazon.com/bedrock/home?region=us-east-1#/api-keys)
2. **Set Environment Variable**:
   ```bash
   export GUD_BEDROCK_API_KEY="your-key"
   # Add to ~/.bashrc or ~/.zshrc for persistence
   ```
3. **Build and Test**:
   ```bash
   cd golang && make build
   git add . && ./bin/gudcommit
   ```

## 🔧 Configuration

### JSON Config (recommended)
You can configure model, timeout, and region without code changes using a JSON file:

Order of precedence: environment variables > `~/.gudcommit.json` > `~/.gudchangelog.json` > defaults

Example `~/.gudcommit.json`:
```json
{
  "model_id": "anthropic.claude-3-5-sonnet-20240620-v1:0",
  "timeout_seconds": 60,
  "region": "us-east-1"
}
```

Defaults if not set:
- `model_id`: `anthropic.claude-3-5-sonnet-20240620-v1:0`
- `timeout_seconds`: `60`
- `region`: `us-east-1`

Environment variable overrides:
- `GUD_BEDROCK_MODEL_ID`
- `GUD_HTTP_TIMEOUT_SECONDS`
- `AWS_REGION`

API key (still required):
- `GUD_BEDROCK_API_KEY` must be set (short‑term key from the Bedrock console)

### Automatic API Key Management
The `scripts/auto-api-key.sh` helper provides:
- **Generation**: Creates short-term API keys using AWS Bedrock token generator
- **Credential storage**: Saves to `~/.gudcommit-credentials`
- **Environment management**: Loads `GUD_BEDROCK_API_KEY` and `AWS_REGION`
- **12-hour expiration**: Keys automatically expire for security

### Manual Configuration
- `GUD_BEDROCK_API_KEY` (required): Your Bedrock API key
- `AWS_REGION` (optional): AWS region (defaults to us-east-1)

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

### Discover available models/inference profiles
List models that support direct on-demand invocation:
```bash
aws bedrock list-foundation-models --region us-east-1 \
  --query "modelSummaries[?contains(inferenceTypesSupported, 'ON_DEMAND')].[modelId,providerName]" \
  --output table
```

List inference profiles (required for many Claude 3.5+ models):
```bash
aws bedrock list-inference-profiles --region us-east-1 \
  --query 'inferenceProfileSummaries[].{arn:inferenceProfileArn,name:inferenceProfileName,model:modelSource.type}'
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

### New Implementation
- **Direct Bedrock API calls** using HTTP requests
- **API key authentication** (no AWS SDK required)
- **Default model**: `anthropic.claude-3-5-sonnet-20240620-v1:0` (configurable)
- **Conventional commit format** with structured JSON parsing

### Benefits
✅ **No IAM users required**  
✅ **No AWS profiles needed**  
✅ **No AWS SDK dependencies**  
✅ **Works anywhere with API key**  
✅ **Better performance** (direct model calls)  
✅ **Simpler setup** (just one environment variable)  

## 🔒 Security

- **Never commit API keys to git**
- **Use environment variables only**
- **Regenerate keys regularly**
- **Short-term keys are more secure**

## 🐛 Troubleshooting

### API Key Issues

#### "GUD_BEDROCK_API_KEY environment variable is not set"
```bash
# Manual method
export GUD_BEDROCK_API_KEY="your-key"

# Automatic method
./scripts/auto-api-key.sh generate
```

#### "Bedrock API error: 403"
- **API key expired**: Regenerate using `./scripts/auto-api-key.sh generate`
- **Permissions**: Check Bedrock permissions in your AWS account
- **Policy**: Ensure you have `AmazonBedrockLimitedAccess` policy attached

#### "Bedrock API error: 400"
- **Request format**: Check request format (should be automatic)
- **Region**: Verify AWS region is correct

### Automatic API Key Management Issues

#### "Python 3 not found"
```bash
# macOS
brew install python3

# Ubuntu
sudo apt install python3 python3-pip

# CentOS
sudo yum install python3 python3-pip
```

#### "AWS credentials not configured"
```bash
# Configure AWS CLI
aws configure

# Or set environment variables
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
```

#### "aws-bedrock-token-generator not found"
```bash
# Install manually
pip3 install aws-bedrock-token-generator
```

### Debug Commands
```bash
# Check API key status
./scripts/auto-api-key.sh status

# Test key generation
./scripts/auto-api-key.sh generate

# Run in debug mode
./scripts/auto-api-key.sh run --debug
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
git clone git@github.com:gudlyf/GudCommit.git
cd GudCommit
cd golang && make build
```

## 📚 Documentation

- [Conventional Commits](https://www.conventionalcommits.org/)
- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)

### Developer Documentation
- [Go Implementation Summary](golang/GO_IMPLEMENTATION_SUMMARY.md) - Architecture and development details

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

**Note**: This version uses direct Bedrock model invocation with API keys, eliminating the need for complex AWS setup while maintaining all the functionality of the original GudCommit.
