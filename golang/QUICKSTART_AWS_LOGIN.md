# AWS Credentials for the Bedrock Provider

The `bedrock` provider uses the standard AWS SDK for Go v2 credential chain, which reads `aws login` sessions natively. No `credential_process` wrapper or extra profile is needed.

## Quick start with `aws login` (recommended)

Requires AWS CLI v2.32.0 or later (`aws --version`).

```bash
aws login --profile gudcommit --region us-east-1
aws configure get login_session --profile gudcommit   # should print your identity ARN
aws sts get-caller-identity --profile gudcommit       # should print your account and ARN
```

`aws login` opens your browser, signs you in with your AWS Console identity, caches a refresh token in `~/.aws/login/cache/`, and records `login_session` for the profile in `~/.aws/config`. The result looks like:

```ini
[profile gudcommit]
login_session = arn:aws:iam::123456789012:user/yourname
region = us-east-1
```

gudcommit and gudchangelog use the `gudcommit` profile when `AWS_PROFILE` is not set. To use a profile you already log in with, point at it instead:

```bash
export AWS_PROFILE=my-profile
```

Temporary credentials are refreshed automatically while the login session is valid (up to 12 hours, or whatever your IAM identity allows). After that, run `aws login --profile gudcommit` again.

## Other credential sources

Anything the AWS SDK supports works, for example:

- **IAM Identity Center (SSO):** `sso_session`/`sso_*` settings in the profile, then `aws sso login --profile <name>`
- **Environment variables:** `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
- **Assumed roles:** `role_arn` + `source_profile`
- **IAM roles** on EC2, ECS, or Lambda

Avoid long-lived access keys in `~/.aws/credentials` where you can; `aws login` and SSO give you short-lived credentials instead.

## Troubleshooting

**"AWS credential error" / "The provided authorization grant is invalid, expired, revoked, or malformed"**
The login session has expired. Run `aws login --profile <profile>` again.

**"failed to get shared config profile, gudcommit"**
No `gudcommit` profile exists. Run the quick start above, or set `AWS_PROFILE` to a profile you use.

**It's using a different identity than you expect**
Static keys win over `login_session`. If `~/.aws/credentials` has a section with the same profile name, or `AWS_ACCESS_KEY_ID` is set in your environment, those are used instead of your login session. Remove them, or use a different profile name.

**Wrong region / model not found**
Set `region` in the profile or `AWS_REGION`, and make sure the model in `model_id` is enabled in that region. `scripts/list-bedrock-models.sh` lists what's available.

## References

- [AWS login credentials](https://docs.aws.amazon.com/sdkref/latest/guide/feature-login-credentials.html)
- [AWS SDK credential provider chain](https://docs.aws.amazon.com/sdkref/latest/guide/standardized-credentials.html)
