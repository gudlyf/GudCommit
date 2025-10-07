## [Unreleased]

### Added
- Go implementation of GudCommit tool
- Support for cross-platform builds
- Comprehensive test suite

### Changed
- Migrated from JavaScript to Go for better performance and portability
- Improved error handling and fallback mechanisms

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
