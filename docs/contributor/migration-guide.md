# Migrating from Kyma CLI to `modulectl`

This guide provides detailed instructions for migrating from the current Kyma CLI tool to the new `modulectl` tool.
It covers all necessary changes and deprecations to ensure a smooth transition.

## Overview

`modulectl` is the successor of the module developer-facing capabilities of Kyma CLI.
It is already tailored for the updated ModuleTemplate metadata as discussed in [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681).

## 1. Tooling & Workflow Changes

This section focuses on the `modulectl` CLI itself and the related submission, deployment, and migration workflows.

### 1.1 Use `modulectl`
Modulectl is available for download from the [GitHub Releases](https://github.com/kyma-project/modulectl/releases).
For an overview of the supported commands and flags, use `modulectl -h` or `modulectl <command> -h` to show the definitions.

```bash
modulectl -h                # general help
modulectl create -h         # help for 'create'
modulectl scaffold -h       # help for 'scaffold'
```

### 1.2 Command & Flag Differences

#### 1.2.1 Command Mapping

| Operation                | Kyma CLI                       | `modulectl`                         |
|--------------------------|--------------------------------|-------------------------------------|
| Scaffold module template | N/A                            | `modulectl scaffold ...`            |
| Create ModuleTemplate CR | `kyma alpha create module ...` | `modulectl create -c <config-file>` |
| List installed modules   | `kyma alpha module list ...`   | N/A                                 |
| Command-specific help    | `kyma alpha module <cmd> -h`   | `modulectl <cmd> -h`                |

#### 1.2.2 Flag & Behavior Differences

| Feature                       | Deprecated in Kyma CLI                            | New in `modulectl`                                             |
| ----------------------------- | ------------------------------------------------- | -------------------------------------------------------------- |
| Config-file flag              | `--module-config-file`                            | `--config-file`, shortcut `-c` (all commands)                  |
| Archive overwrite             | `--module-archive-version-overwrite`              | **Removed** (module versions are immutable)                    |
| ModuleTemplate naming pattern | `.metadata.name = <module-name>-<channel>`        | `.metadata.name = <module-name>-<version>`                     |
| Version label                 | `operator.kyma-project.io/module-version` label   | Populated in `.spec.version`                                   |
| Documentation annotation      | `operator.kyma-project.io/doc-url` annotation     | Defined in `.spec.info.documentation` via `documentation:` key |
| Manifest & DefaultCR source   | Local file references only                        | Fetched directly from GitHub release URLs                      |
| Release channel in config     | Required `.channel` field in `module-config.yaml` | **Removed**; channel mapping managed by ModuleReleaseMeta CR   |

### 1.3 Submission Process & Migration Period

#### Module Submission Process

1. **Publish Artifacts**: Release module version assets (manifest, defaultCR) on GitHub.
2. **Submit Version Config**: Add/update `/modules/<module>/<version>/module-config.yaml`.
3. **Update Channel Mapping**: Edit `/modules/<module>/module-releases.yaml` to map channels to versions.
4. **Trigger Pipeline**: On PR merge, the submission pipeline:

   * Validates schema, FQDN, version uniqueness
   * Builds via `modulectl`, pushes OCI image
   * Generates ModuleTemplate and ModuleReleaseMeta in `/kyma/kyma-modules`
5. **ArgoCD Sync**: ArgoCD deploys ModuleTemplates and ModuleReleaseMeta to KCP landscapes.
6. **Cleanup**: Remove obsolete channel-based configs when stable.

#### Migration Period & Coexistence

* Both channel-based (old) and version-based (new) metadata are supported during the transition.
* KLM reads version-based templates first; falls back to channel-based if missing.
* After full migration, legacy `<module>-<channel>` templates are rejected.

---
## 2. Module Configuration & Metadata Changes

This section covers the structure and content of `module-config.yaml` and `module-releases.yaml` files under the new version-based layout.

### 2.1 Directory Structure

* **Old Layout**: `/modules/<module-name>/<channel>/module-config.yaml`
* **New Layout**: `/modules/<module-name>/<version>/module-config.yaml`

### 2.2 `module-config.yaml` Schema Comparison Example

This comparison shows a generic module configuration. Replace `<module-name>`, `<channel>`, and `<version>` with your module’s actual values.

##### Old Format

```yaml
name: kyma-project.io/module/<module-name>
channel: <channel>
version: <version>
manifest: <module-name>-manifest.yaml
defaultCR: <module-name>-default-cr.yaml
annotations:
  operator.kyma-project.io/doc-url: https://help.sap.com/.../<module-name>-module
moduleRepo: https://github.com/kyma-project/<module-name>.git
```

##### New Format

```yaml
# modules/<module-name>/<version>/module-config.yaml
name: kyma-project.io/module/<module-name>
repository: https://github.com/kyma-project/<module-manager-name>.git
version: 1.34.0
manifest: https://github.com/kyma-project/<module-manager>/releases/download/1.34.0/<module-manager-name>.yaml
defaultCR: https://github.com/kyma-project/<module-manager>/releases/download/1.34.0/<module-name-default-cr>.yaml
security: sec-scanners-config.yaml
manager:
   name: <module-manager-name>
   namespace: kyma-system
   group: apps
   version: v1
   kind: Deployment
associatedResources:
   - group: operator.kyma-project.io
     kind: <ModuleName>
     version: v1alpha1
   - group: operator.kyma-project.io
     kind: LogParser
     version: v1alpha1
   - group: operator.kyma-project.io
     kind: LogPipeline
     version: v1alpha1
   - group: operator.kyma-project.io
     kind: MetricPipeline
     version: v1alpha1
   - group: operator.kyma-project.io
     kind: TracePipeline
     version: v1alpha1
documentation: https://help.sap.com/docs/btp/sap-business-technology-platform/<kyma-module-name>
icons:
   - name: module-icon
      # TODO: provide <module-name> icon
     link: https://raw.githubusercontent.com/kyma-project/kyma/refs/heads/main/docs/assets/logo_icon.svg
```

### 2.3 Channel Mapping with ModuleReleaseMeta Example

```yaml
# modules/<module-name>/module-releases.yaml
channels:
  - channel: regular
    version: 1.34.0
  - channel: fast
    version: 1.34.0
  - channel: experimental
    version: 1.34.0-experimental
  - channel: dev
    version: 1.35.0-rc1
```

On merge, the pipeline:

* Validates no downgrades and version existence
* Generates ModuleReleaseMeta CRs per landscape
* Updates landscape-specific kustomizations to reference only active versions

### 2.4 Metadata Deprecations & New Practices

| Deprecated Feature                | Replacement / New Location                                                    |
| --------------------------------- | ----------------------------------------------------------------------------- |
| `.channel` field in module config | Moved to ModuleReleaseMeta CR (`module-releases.yaml`)                        |
| `mandatory` on ModuleTemplate     | Set `mandatory: true` in module config; reconciler picks highest version      |
| Beta/Internal flags on templates  | Configured in ModuleReleaseMeta via `.spec.info.beta` / `.spec.info.internal` |

---

## Additional Resources

- [`modulectl` GitHub Repository](https://github.com/kyma-project/modulectl)
- [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681)
- [ModuleTemplate Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/03-moduletemplate.md)
- [ModuleReleaseMeta Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md)
