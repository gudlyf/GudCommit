## 2.0.0

### Added
- Go implementation of GudCommit and GudChangelog tools
- CI/CD pipeline with GitHub Actions for building, testing, and releasing
- Cross-platform build support for Linux, macOS, and Windows
- API key authentication for Bedrock model invocation
- JSON configuration file support for customizing model and timeout settings
- Automatic API key management script for improved user experience
- Comprehensive test suite for Go implementation
- GitLab CI configuration for alternative CI/CD setup

### Changed
- Migrated from JavaScript to Go for improved performance and portability
- Replaced AWS SDK and Bedrock agents with direct HTTP requests to Bedrock API
- Updated project structure to support Go modules and packages
- Improved error handling and fallback mechanisms in Go implementation
- Refactored README with detailed setup and usage instructions for Go version
- Updated CHANGELOG format to be more detailed and follow Keep a Changelog conventions
- Simplified authentication process by using API keys instead of AWS credentials

### Removed
- JavaScript implementation of GudCommit and GudChangelog
- Terraform configurations for Bedrock agents
- AWS SDK dependencies and related configuration
- Node.js package management files (package.json, package-lock.json)
- GNU General Public License (replaced with new licensing, not specified in diff)

---

## [Unreleased]

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

## [Unreleased]

## Added

- Support for the `hashicorp/null` provider version 3.2.2.
- New `null_resource` blocks `prepare_gudcommit` and `prepare_gudchangelog` to execute AWS CLI commands for preparing the respective agents after their creation.

## Changed

- Renamed output variables in `dev/outputs.tf` and `module/outputs.tf` to include prefixes `gudcommit_` and `gudchangelog_` for better clarity and organization.
- Updated dependencies for the new `null_resource` blocks to ensure they run after the agent resources are created.

The changes were made to improve the organization and maintainability of the Terraform configuration. The addition of the `null_resource` blocks allows for executing additional commands after the agent resources are created, potentially for further setup or configuration. The renaming of output variables provides better clarity and separation between the different agents in the project.
