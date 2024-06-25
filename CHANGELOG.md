## Added

- Support for the `hashicorp/null` provider version 3.2.2.
- New `null_resource` blocks `prepare_gudcommit` and `prepare_gudchangelog` to execute AWS CLI commands for preparing the respective agents after their creation.

## Changed

- Renamed output variables in `dev/outputs.tf` and `module/outputs.tf` to include prefixes `gudcommit_` and `gudchangelog_` for better clarity and organization.
- Updated dependencies for the new `null_resource` blocks to ensure they run after the agent resources are created.

---
