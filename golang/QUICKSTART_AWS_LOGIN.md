# Quick Start: Using `aws login` with gudcommit/gudchangelog

This guide shows you how to set up the NEW `aws login` method with your gudcommit/gudchangelog tools.

## Prerequisites

Check your AWS CLI version (must be v2.32.0 or later):
```bash
aws --version
```

If you need to upgrade: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

## Setup (5 minutes)

### Step 1: Log in with AWS Console Credentials

```bash
aws login --profile my-aws
```

This opens your browser to authenticate. After logging in, credentials are cached in `~/.aws/login/cache`.

### Step 2: Configure credential_process

Edit `~/.aws/config` and add:

```ini
[profile my-aws]
login_session = arn:aws:iam::123456789012:user/yourname
region = us-east-1

[profile gudcommit]
credential_process = aws configure export-credentials --profile my-aws --format process
region = us-east-1
```

**Why?** AWS SDK for Go v2 doesn't yet natively read `login_session`, so we use `credential_process` to delegate to the AWS CLI, which does support it.

### Step 3: Test It

```bash
aws sts get-caller-identity --profile gudcommit
```

If this works, you're ready!

### Step 4: Use gudcommit

```bash
cd /path/to/your/git/repo
git add .
gudcommit  # Automatically uses 'gudcommit' profile by default!
```

**Note**: gudcommit automatically defaults to the `gudcommit` profile, so you don't need to set `AWS_PROFILE` unless you want to use a different profile.

## How It Works

1. **Login**: `aws login` gets your Console credentials
2. **Cache**: Credentials stored in `~/.aws/login/cache/`
3. **Short-lived**: Credentials expire every 15 minutes
4. **Auto-refresh**: AWS CLI (via `credential_process`) automatically refreshes them
5. **Max session**: Up to 12 hours, then run `aws login` again

## Troubleshooting

### "Unable to load AWS SDK config"
- Run `aws login --profile my-aws` first
- Check `~/.aws/config` has the `credential_process` line
- Set `AWS_PROFILE=gudcommit`

### "credential_process failed"
Test the command directly:
```bash
aws configure export-credentials --profile my-aws --format process
```

If this fails:
- Your login session may have expired (max 12 hours)
- Run `aws login --profile my-aws` again

### "No credentials found"
gudcommit automatically uses the `gudcommit` profile by default. Make sure `~/.aws/config` has a `[profile gudcommit]` section with `credential_process` configured.

Alternatively, you can override the profile:
```bash
export AWS_PROFILE=my-other-profile
```

Or use environment variables directly:
```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_REGION=us-east-1
```

## Alternative: Environment Variables

If you prefer not to use `aws login`, you can use traditional credentials:

```bash
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
export AWS_SESSION_TOKEN=your-token  # if temporary
export AWS_REGION=us-east-1
gudcommit
```

## Benefits of `aws login`

- ✅ No long-lived access keys to manage
- ✅ Automatic credential rotation every 15 minutes
- ✅ Uses your Console login (IAM user, root, or federated)
- ✅ More secure than static credentials
- ✅ Same credentials work across AWS CLI and SDK tools

## More Information

For complete details, see:
- [AWS_SSO_SUPPORT.md](./AWS_SSO_SUPPORT.md) - Full documentation
- [README.md](./README.md) - General gudcommit documentation
- [AWS Login Docs](https://docs.aws.amazon.com/sdkref/latest/guide/feature-login-credentials.html)
