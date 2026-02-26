# AWS Login Credentials Support for gudcommit/gudchangelog

## Overview

The Go implementation now supports the new **`aws login`** credentials method that stores credentials in `~/.aws/login/cache` with automatic rotation. This update migrates from custom API key authentication to using the official AWS SDK for Go v2.

## What Changed

### Previous Implementation
- Used custom API key authentication via `GUD_BEDROCK_API_KEY` environment variable
- Made direct HTTP requests to Bedrock with Bearer token authentication
- Did not support AWS login or standard AWS credential providers

### New Implementation
- Uses AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2`)
- Supports multiple credential providers in the standard AWS credential chain:
  - **NEW: `aws login` credentials** from `~/.aws/login/cache` (15-minute auto-refresh, up to 12 hours)
  - `credential_process` (enables `aws login` support via CLI delegation)
  - Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`)
  - Shared credentials file (`~/.aws/credentials`)
  - Legacy AWS SSO (`aws sso login`)
  - IAM roles
  - EC2 instance metadata
- Automatically handles SigV4 request signing
- Includes automatic retry logic for transient errors

## Important Note: Go SDK v2 and `aws login`

**AWS SDK for Go v2 does not yet natively support the `login_session` provider.** However, it works via the `credential_process` workaround, where the SDK delegates to the AWS CLI to fetch credentials. The CLI manages the `~/.aws/login/cache` and auto-refresh.

## AWS SDK Dependencies

The following packages were added to support AWS SSO:

```
github.com/aws/aws-sdk-go-v2 v1.30.5
github.com/aws/aws-sdk-go-v2/config v1.27.33
github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.15.0
github.com/aws/smithy-go v1.20.4
```

Plus additional dependencies for SSO support:
- `github.com/aws/aws-sdk-go-v2/credentials`
- `github.com/aws/aws-sdk-go-v2/service/sso`
- `github.com/aws/aws-sdk-go-v2/service/ssooidc`
- `github.com/aws/aws-sdk-go-v2/service/sts`

## Usage

### Using the NEW `aws login` Method (Recommended)

The new `aws login` command (AWS CLI v2.32.0+) provides short-lived, auto-rotating credentials from your AWS Console login.

#### Method 1: Direct Profile (Simple Setup)

1. Log in with the new `aws login` command:
```bash
aws login --profile my-profile
```

This opens your browser to authenticate with AWS Console credentials (IAM user, root, or federated). After successful login:
- Credentials are cached in `~/.aws/login/cache`
- Your `~/.aws/config` is updated with `login_session`
- Credentials expire every 15 minutes but auto-refresh (up to 12 hours max)

2. Set your AWS profile and run:
```bash
export AWS_PROFILE=my-profile
gudcommit
```

**However**: Since Go SDK v2 doesn't natively read `login_session` yet, you need to set up `credential_process` as shown in Method 2.

#### Method 2: Using `credential_process` (Recommended for Go SDK v2)

This is the workaround that makes `aws login` work with Go SDK v2:

1. First, log in:
```bash
aws login --profile my-login-profile
```

2. Edit your `~/.aws/config` to add a profile that uses `credential_process`:
```ini
[profile my-login-profile]
login_session = arn:aws:iam::123456789012:user/username
region = us-east-1

[profile my-app-profile]
credential_process = aws configure export-credentials --profile my-login-profile --format process
region = us-east-1
```

3. Use the app profile:
```bash
export AWS_PROFILE=my-app-profile
gudcommit
```

The SDK delegates to the AWS CLI via `credential_process`, and the CLI manages the login cache and auto-refresh.

### Using Legacy AWS SSO (aws sso login)

If your organization uses AWS SSO (not the new `aws login`):

1. Configure your AWS SSO profile in `~/.aws/config`:
```ini
[profile my-sso-profile]
sso_start_url = https://my-domain.awsapps.com/start
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = MyRole
region = us-east-1
```

2. Log in:
```bash
aws sso login --profile my-sso-profile
```

3. Set profile and run:
```bash
export AWS_PROFILE=my-sso-profile
gudcommit
```

### Using Other Credential Methods

#### Environment Variables
```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_SESSION_TOKEN=your-token  # if using temporary credentials
export AWS_REGION=us-east-1
gudcommit
```

#### Shared Credentials File
Configure `~/.aws/credentials`:
```ini
[default]
aws_access_key_id = your-key
aws_secret_access_key = your-secret
```

Then run:
```bash
gudcommit
```

#### IAM Role (on EC2/ECS/Lambda)
No configuration needed - credentials are automatically retrieved from instance metadata.

## Error Handling

If there are credential issues, you'll receive a helpful error message:

```
AWS credential error: ...

Please ensure you are logged in with 'aws sso login' or have valid AWS credentials configured
```

## Prerequisites

To use the new `aws login` method:
- AWS CLI version 2.32.0 or later
- Run `aws --version` to check your version
- Upgrade if needed: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

## Migration Notes

### Removed
- `GUD_BEDROCK_API_KEY` environment variable is no longer used
- Custom Bearer token authentication removed

### Maintained
- `AWS_REGION` environment variable still works
- Configuration files (`~/.gudcommit.json`, `~/.gudchangelog.json`) still work
- `GUD_BEDROCK_MODEL_ID` and `GUD_HTTP_TIMEOUT_SECONDS` environment variables still work

### Go SDK v2 Limitation
- Go SDK v2 does not yet natively support `login_session` from `~/.aws/login/cache`
- **Workaround**: Use `credential_process` in your AWS config (see Method 2 above)
- Alternative: Use environment variables or other credential methods

## Building

```bash
cd golang
go mod tidy
go build ./cmd/gudcommit
go build ./cmd/gudchangelog
```

## Testing

```bash
cd golang
go test ./...
```

## Credential Precedence

The AWS SDK uses the following credential provider chain (in order):

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`)
2. `credential_process` (enables `aws login` support) - **Recommended for new `aws login`**
3. Shared credentials file (`~/.aws/credentials`)
4. SSO credentials (legacy `aws sso login`)
5. Web identity token from STS
6. ECS/Fargate container credentials
7. EC2 instance metadata

This means if you have both environment variables set AND use `aws login`, the environment variables will take precedence.

## How `aws login` Credentials Work

1. **Initial Login**: `aws login` opens browser for Console authentication
2. **Cache**: Credentials stored in `~/.aws/login/cache/` (can override with `AWS_LOGIN_CACHE_DIRECTORY`)
3. **Short-lived**: Credentials expire every 15 minutes
4. **Auto-refresh**: AWS CLI (via `credential_process`) automatically refreshes them
5. **Max session**: Session lasts up to 12 hours, then you must run `aws login` again
6. **Profile config**: `login_session = arn:aws:iam::account:user/name` added to `~/.aws/config`

## Troubleshooting `aws login`

### Error: "Unable to load AWS SDK config"
- Ensure AWS CLI version 2.32.0 or later
- Run `aws login --profile your-profile` first
- Set `AWS_PROFILE=your-profile`
- Or use `credential_process` workaround (Method 2)

### Error: "credential_process failed"
- Verify AWS CLI is in your PATH
- Test: `aws configure export-credentials --profile your-profile --format process`
- Check that your login session hasn't expired (max 12 hours)

### Credentials not refreshing
- The `credential_process` method handles refresh automatically via AWS CLI
- If using profile directly with `login_session`, remember Go SDK v2 doesn't support it yet
- Use Method 2 (credential_process) for auto-refresh

### Session expired after 12 hours
- Run `aws login --profile your-profile` again
- The maximum session duration is set by your IAM identity
