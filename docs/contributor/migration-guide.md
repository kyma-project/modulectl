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

| Operation                           | Kyma CLI                         | `modulectl`                         |
|-------------------------------------|----------------------------------|-------------------------------------|
| Scaffold module necessary files     | `kyma alpha create scaffold ...` | `modulectl scaffold ...`            |
| Create Bundled Module(OCI artifact) | `kyma alpha create module ...`   | `modulectl create -c <config-file>` |
| Command-specific help               | `kyma alpha module <cmd> -h`     | `modulectl <cmd> -h`                |

#### 1.2.2 Flag & Behavior Differences by Command

##### 1.2.2.1 Scaffold Command Flags

| Flag (Kyma CLI v2.20.5)                                        | Flag (`modulectl scaffold`)    | Notes                                                                        |
| -------------------------------------------------------------- | ------------------------------ | ---------------------------------------------------------------------------- |
| `-d, --directory string`                                       | `-d, --directory string`       | Target directory for generated scaffold files (default `./`)                 |
| `--gen-default-cr string [=\"default-cr.yaml\"]`               | `--gen-default-cr string`      | Name of generated default CR (blank if missing; default `default-cr.yaml`)   |
| `--gen-manifest string [=\"manifest.yaml\"]`                   | `--gen-manifest string`        | Name of generated manifest file (default `manifest.yaml`)                    |
| `--gen-security-config string [=\"sec-scanners-config.yaml\"]` | `--gen-security-config string` | Name of generated security config (default `sec-scanners-config.yaml`)       |
| `--module-channel string`                                      | *Removed*                      | Channel no longer set at scaffold time                                       |
| `--module-config string [=\"scaffold-module-config.yaml\"]`    | `-c, --config-file string`     | Name of generated module config file (default `scaffold-module-config.yaml`) |
| `--module-name string`                                         | `--module-name string`         | Module name in generated config (default `kyma-project.io/module/mymodule`)  |
| `--module-version string`                                      | `--module-version string`      | Module version in generated config (default `0.0.1`)                         |
| `-o, --overwrite`                                              | `-o, --overwrite`              | Overwrite existing module config file                                        |
| `-h, --help`                                                   | `-h, --help`                   | Show help for scaffold command                                               |

##### 1.2.2.2 Create Command Flags

| Flag (Kyma CLI v2.20.5)              | Flag (`modulectl create`)         | Notes                                                         |
| ------------------------------------ | --------------------------------- | ------------------------------------------------------------- |
| `--module-config-file string`        | `-c, --config-file string`        | Path to your `module-config.yaml`                             |
| `--module-archive-path string`       | *Not supported*                   | Archive path for local module artifacts                       |
| `--module-archive-persistence`       | *Not supported*                   | Persist module archive on host filesystem                     |
| `--module-archive-version-overwrite` | `--overwrite`                     | Overwrite existing module OCI archive (testing only)          |
| `--descriptor-version string`        | *Not supported*                   | Schema version for generated descriptor                       |
| `--git-remote string`                | *Not supported*                   | Git remote name for module sources                            |
| `--insecure`                         | `--insecure`                      | Allow insecure registry connections                           |
| `--key string`                       | *Not supported*                   | Private key path for signing                                  |
| `--kubebuilder-project`              | *Not supported*                   | Indicate Kubebuilder project                                  |
| `-n, --name string`                  | `--name`                          | Override module name                                          |
| `--name-mapping string`              | *Not supported*                   | OCM component name mapping                                    |
| `--namespace string`                 | `--namespace`                     | Namespace for generated ModuleTemplate (default `kcp-system`) |
| `-o, --output string`                | `-o, --output string`             | Output path for ModuleTemplate (default `template.yaml`)      |
| `-p, --path string`                  | *Not supported*                   | Path to module contents                                       |
| `-r, --registry string`              | `-r, --registry string`           | Context URL for OCI registry                                  |
| `--registry-cred-selector string`    | `--registry-cred-selector string` | Label selector for existing `dockerconfigjson` Secret         |
| `--registry-credentials string`      | `--registry-credentials string`   | Basic auth credentials in `<user:password>` format            |
| `--dry-run`                          | `--dry-run`                       | Validate and skip pushing module descriptor                   |
| `-h, --help`                         | `-h, --help`                      | Show help for create command                                  |

## 2. Module Configuration & Metadata Changes

This section illustrates how the `module-config.yaml` looks in the **Kyma CLI** format versus the **ModuleCtl** format, with field-by-field mapping and examples.

## 2.1 Field Mapping Differences

| Kyma CLI                                       | ModuleCtl (new)          | Description / Changes                                                                     |
|------------------------------------------------|--------------------------|-------------------------------------------------------------------------------------------|
| `name`                                         | `name`                   | Module identifier (unchanged)                                                             |
| `channel`                                      | *removed*                | Release channel moved to `module-releases.yaml` (ReleaseMetadata)                         |
| `version`                                      | `version`                | Explicit version in both configs                                                          |
| `manifest`                                     | `manifest`               | Local file → must be a URL (e.g. GitHub release asset)                                    |
| `defaultCR`                                    | `defaultCR`              | Local file → URL                                                                          |
| `annotations.operator.kyma-project.io/doc-url` | `documentation`          | Moved from annotations map to top-level `documentation` key                               |
| `moduleRepo`                                   | `repository`             | Renamed to `repository`                                                                   |
| *n/a*                                          | `icons`                  | New required list: icons for UI, with `name`+`link`                                       |
| *n/a*                                          | `mandatory`              | New boolean (default `false`) to mark mandatory modules                                   |
| *n/a*                                          | `requiresDowntime`       | New boolean (default `false`) for maintenance windows                                     |
| *n/a*                                          | `security`               | Path to security scanner config                                                           |
| *n/a*                                          | `labels` / `annotations` | Pass-through for additional metadata                                                      |
| *n/a*                                          | `manager`                | Defines the module’s controller resource (name, group, version, kind, optional namespace) |
| *n/a*                                          | `associatedResources`    | List of GVKs to be cleaned up on uninstall                                                |
| *n/a*                                          | `resources`              | Additional artifacts (e.g., CRDs)                                                         |
| *n/a*                                          | `namespace`              | Target namespace for the generated `ModuleTemplate` (default `kcp-system`)                |

## 2.2 Example:

### 2.2.1 Module Config using Kyma CLI

```yaml
name: kyma-project.io/module/<module-name>
channel: <channel>
version: <version>
manifest: <module-name>-manifest.yaml
defaultCR: <module-name>-default-cr.yaml
annotations:
  operator.kyma-project.io/doc-url: https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-module-name
moduleRepo: https://github.com/kyma-project/module-manager.git
```

### 2.2.2 Module Config using ModuleCtl

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
     link: https://raw.githubusercontent.com/kyma-project/kyma/refs/heads/main/docs/assets/logo_icon.svg
```

## 2.2 Channel Mapping with ModuleReleaseMeta

The module-releases.yaml file defines how logical release channels (e.g. regular, fast, experimental, dev) map to concrete module versions.
During the submission pipeline, each entry is turned into a ModuleReleaseMeta Custom Resource, which ensures that clients subscribing to a given channel always receive the correct version of your module.

Note: Beta/Internal flags on templates  | Configured in ModuleReleaseMeta via `.spec.beta` / `.spec.internal` |


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

# 3. Module artifacts generated 
## 3.1  ModuleTemplate Comparison

Below is a side-by-side comparison of the key fields in the `ModuleTemplate` YAML generated by Kyma CLI vs. modulectl:

| Field / Path                                | Kyma CLI–Generated (`kyma alpha create module`)      | modulectl–Generated (`modulectl create`)                    |
|---------------------------------------------|------------------------------------------------------|-------------------------------------------------------------|
| **metadata.name**                           | `<module-name>-<channel>`                            | `<module-name>-<version>`                                   |
| **metadata.namespace**                      | `kcp-system`                                         | `kcp-system`                                                |
| **metadata.labels.operator.kyma-project.io/module-name** | present                                              | present                                                     |
| **metadata.annotations.operator.kyma-project.io/doc-url** | present                                              | _omitted_                                                   |
| **metadata.annotations.operator.kyma-project.io/is-cluster-scoped** | present                                              | present                                                     |
| **metadata.annotations.operator.kyma-project.io/module-version** | present                                              | _omitted_                                                   |
| **spec.channel**                            | `<channel>`                                          | _omitted_                                                   |
| **spec.moduleName**                         | _omitted_                                            | `<module-name>`                                             |
| **spec.version**                            | _omitted_                                            | `<version>`                                                 |
| **spec.mandatory**                          | `false`                                              | `false`                                                     |
| **spec.requiresDowntime**                   | _omitted_                                            | `false`                                                     |
| **spec.data**                               | embedded CR (`apiVersion`, `kind`, `metadata.name`)  | _omitted_                                                   |
| **spec.info**                               | _omitted_                                            | populated (`repository`, `documentation`, `icons`)         |
| **spec.descriptor.component.name**          | `kyma-project.io/module/<module-name>`               | `kyma-project.io/module/<module-name>`                      |
| **spec.descriptor.component.provider**      | built-by `"cli"`                                     | built-by `"modulectl"`                                      |
| **spec.descriptor.component.repositoryContexts** | OCI registry contexts                               | GitHub/OCI contexts as configured                           |
| **spec.descriptor.component.resources**     | mix of `ociRegistry` images & `localBlob` entries    | `localBlob` entry for raw-manifest                          |
| **spec.descriptor.component.sources**       | GitHub sources with scan labels                      | GitHub sources without extra scan labels                    |
| **spec.resources**                          | _omitted_                                            | list of additional artifacts (e.g. `rawManifest` link)      |

The following shows the `ModuleTemplate` YAML generated by Kyma CLI vs. modulectl, using placeholders for `<module-name>`, `<channel>`, and `<version>`.

## 3.1 Kyma CLI–Generated ModuleTemplate (channel-based)

```yaml
apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: <module-name>-<channel>
  namespace: kcp-system
  labels:
    "operator.kyma-project.io/module-name": "<module-name>"
  annotations:
    "operator.kyma-project.io/doc-url": "<documentation-url>"
    "operator.kyma-project.io/is-cluster-scoped": "false"
    "operator.kyma-project.io/module-version": "<version>"
spec:
  channel: <channel>
  mandatory: false
  data:
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: <ModuleKind>
    metadata:
      name: default
  descriptor:
    component:
      componentReferences: []
      labels:
      - name: <scan-label-name>
        value: <scan-label-value>
        version: <label-version>
      name: kyma-project.io/module/<module-name>
      provider: '{"name":"kyma-project.io","labels":[{"name":"kyma-project.io/built-by","value":"cli","version":"v1"}]}'
      repositoryContexts:
      - baseUrl: <oci-registry-url>
        componentNameMapping: urlPath
        type: OCIRegistry
      resources:
      - access:
          imageReference: <oci-image>:<version>
          type: ociRegistry
        labels:
        - name: <image-scan-label>
          value: <image-scan-value>
          version: v1
        name: <component-name>
        relation: external
        type: ociImage
        version: <version>
      - access:
          localReference: <sha256-digest>
          mediaType: application/octet-stream
          referenceName: raw-manifest
          type: localBlob
        digest:
          hashAlgorithm: SHA-256
          normalisationAlgorithm: genericBlobDigest/v1
          value: <sha256-digest>
        name: raw-manifest
        relation: local
        type: yaml
        version: <version>
      sources:
      - access:
          commit: <git-commit>
          repoUrl: <repo-url>
          type: gitHub
        labels:
        - name: git.kyma-project.io/ref
          value: HEAD
          version: v1
        name: module-sources
        type: Github
        version: <version>
      version: <version>
    meta:
      schemaVersion: v2
```

## 3.2 modulectl–Generated ModuleTemplate (version-based)

```yaml
apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: <module-name>-<version>
  namespace: kcp-system
  labels:
    "operator.kyma-project.io/module-name": "<module-name>"
  annotations:
    "operator.kyma-project.io/is-cluster-scoped": "false"
spec:
  moduleName: <module-name>
  version: <version>
  mandatory: false
  requiresDowntime: false
  info:
    repository: <repo-url>
    documentation: <documentation-url>
    icons:
    - name: <icon-name>
      link: <icon-url>
  descriptor:
    component:
      componentReferences: []
      name: kyma-project.io/module/<module-name>
      provider: '{"name":"kyma-project.io","labels":[{"name":"kyma-project.io/built-by","value":"modulectl","version":"v1"}]}'
      repositoryContexts:
      - baseUrl: <oci-registry-url>
        componentNameMapping: urlPath
        type: OCIRegistry
      resources:
      - access:
          localReference: <sha256-digest>
          mediaType: application/x-tar
          referenceName: raw-manifest
          type: localBlob
        digest:
          hashAlgorithm: SHA-256
          normalisationAlgorithm: genericBlobDigest/v1
          value: <sha256-digest>
        name: raw-manifest
        relation: local
        type: directory
        version: <version>
      sources:
      - access:
          commit: <git-commit>
          repoUrl: <repo-url>
          type: gitHub
        labels:
        - name: git.kyma-project.io/ref
          value: HEAD
          version: v1
        name: module-sources
        type: Github
        version: <version>
      version: <version>
    meta:
      schemaVersion: v2
  resources:
  - name: rawManifest
    link: <manifest-url>
```



## Additional Resources

- [`modulectl` GitHub Repository](https://github.com/kyma-project/modulectl)
- [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681)
- [ModuleTemplate Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/03-moduletemplate.md)
- [ModuleReleaseMeta Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md)
