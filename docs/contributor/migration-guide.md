# Migrating from Kyma CLI to `modulectl`

This guide provides detailed instructions for migrating from the current Kyma CLI tool to the new `modulectl` tool.
It covers all necessary changes and deprecations to ensure a smooth transition.

## Overview

`modulectl` is the successor of the module developer-facing capabilities of Kyma CLI.
It is already tailored for the updated ModuleTemplate metadata as discussed in [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681).

## Use `modulectl`

It is available for download from the [GitHub Releases](https://github.com/kyma-project/modulectl/releases).
For an overview of the supported commands and flags, use `modulectl -h` or `modulectl <command> -h` to show the definitions.

## Deprecations and Changes

Look at the key changes between Kyma CLI and `modulectl`.

1. **New Command**

   - The new command for module creation is `modulectl create`, use `modulectl create -h` for a detailed description. 

2. **Release Channel Configuration**

   - **Deprecated**: The **.channel** field is no longer required in the config file.
   - **New Approach**: The release channel is configured separately in the [ModuleReleaseMeta CR](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md).

3. **Manifest and Default CR data**

   - **Deprecated**: The Manifest and Default CR data cannot be read from a local file.
   - **New Approach**: The Manifest and Default CR data is fetched directly from the GitHub release. See an example [here](https://github.com/kyma-project/modulectl/blob/91e01856b944fda0d5595843e040bca26416abdc/tests/e2e/create/testdata/moduleconfig/valid/with-defaultcr.yaml#L3-L4).

4. **Annotations and Labels**

   - **Deprecated**: The documentation link is not added as the `operator.kyma-project.io/doc-url` annotation.
   - **New**: The documentation link is added as **.spec.info.documentation**. The value is configured in the module config file with key **.documentation**.
   - **Deprecated**: The module version is not added as the `operator.kyma-project.io/module-version` label.
   - **New**: The module version is added as **.spec.version**.

5. **Command Flags**

   - **Deprecated**: Flag `--module-config-file` is deprecated.
   - **New**: Flag `--config-file` with shortcut `-c` (applicable for both scaffold and create commands).
   - **Deprecated**: Flag `--module-archive-version-overwrite` is deprecated.
     - There is no successor, this feature has been sunset entirely. The reason is that module versions should be immutable once built and pushed. If a version needs to be rewritten, first it should be explicitly deleted from the registry.

6. **ModuleTemplate Naming Pattern**

   - **Deprecated**: **.metadata.name** is not written as `<module-name>-<channel>`.
   - **New**: **.metadata.name** is written as `<module-name>-<version>`.

7. **Mandatory ModuleTemplates**

   - **Restriction**: At the moment, `modulectl` is not used for the creation of mandatory ModuleTemplates. Please use Kyma CLI instead until the next release.

8. **Beta and Internal Flags**

   - **Deprecated**: Beta and Internal flags are not supported for ModuleTemplates.
   - **New**: Beta and Intenral flags are configured as part of the ModuleReleaseMeta, see [ModuleReleaseMeta Configuration](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md#configuration).

## Migration Period

**Support for Both Approaches Temporarily**: 

- Both old and new approaches will be supported simultaneously by KLM during the migration period.
- After the full migration to the new approach, KLM will no longer accept ModuleTemplate with the old naming pattern.
- For testing, provide ModuleTemplate CRs in the new format accompanied by [ModuleReleaseMeta CRs](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md). KLM will attempt to use the new approach and fall back to the old approach if the new format is not found. The ModuleReleaseMeta CR can be marked as internal and beta to avoid syncing it.

## Submission Process

The general submission process may look as follows:

1. Module Team releases a new module version in the GitHub release.
2. Module Team configures the version to be released in the module config (in the new format) in the internal module-manifest repo.
3. Module Team updates the version in the related channel in ModuleReleaseMeta.
4. The submission pipeline gets triggered.
5. After certain quality gates are passed, the new version of ModuleTemplate gets provisioned into KCP first.
6. The updated ModuleReleaseMeta gets provisioned into KCP.
7. Outdated ModuleTemplate (the version not mentioned in ModuleReleaseMeta) should be removed. This step allows both outdated and new versions to coexist temporarily. KLM only handles the version defined in ModuleReleaseMeta.

## Additional Resources

- [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681)
- [`modulectl` GitHub Repository](https://github.com/kyma-project/modulectl)
