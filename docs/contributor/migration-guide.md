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

This section illustrates how the commands from Kyma CLI are mapped to `modulectl` format.

| Operation                           | Kyma CLI                         | `modulectl`                         |
|-------------------------------------|----------------------------------|-------------------------------------|
| Scaffold module necessary files     | `kyma alpha create scaffold ...` | `modulectl scaffold ...`            |
| Create Bundled Module(OCI artifact) | `kyma alpha create module ...`   | `modulectl create -c <config-file>` |
| Command-specific help               | `kyma alpha module <cmd> -h`     | `modulectl <cmd> -h`                |

#### 1.2.2 Flag Mapping

This section illustrates how the command flags from Kyma CLI are mapped to `modulectl` format.

##### 1.2.2.1 Scaffold Flag Mapping

| Flag (Kyma CLI v2.20.5)                                        | Flag (`modulectl`)                         | Notes                                                                        |
| -------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------- |
| `-d, --directory string`                                       | `-d, --directory string`                   | Target directory for generated scaffold files (default `./`)                 |
| `--gen-default-cr string`                                      | `--gen-default-cr string`                  | Name of generated default CR (default `default-cr.yaml`)                     |
| `--gen-manifest string`                                        | `--gen-manifest string`                    | Name of generated manifest file (default `manifest.yaml`)                    |
| `--gen-security-config string`                                 | `--gen-security-config string`             | Name of generated security config (default `sec-scanners-config.yaml`)       |
| `--module-channel string`                                      | **Removed**                                | Channel no longer set at scaffold time                                       |
| `--module-config string`                                       | **Renamed** `-c, --config-file string`     | Name of generated module config file (default `scaffold-module-config.yaml`) |
| `--module-name string`                                         | `--module-name string`                     | Module name in generated config (default `kyma-project.io/module/mymodule`)  |
| `--module-version string`                                      | `--module-version string`                  | Module version in generated config (default `0.0.1`)                         |
| `-o, --overwrite`                                              | `-o, --overwrite`                          | Overwrite existing module config file                                        |
| `-h, --help`                                                   | `-h, --help`                               | Show help for scaffold command                                               |

##### 1.2.2.2 Create Flag Mapping

| Flag (Kyma CLI v2.20.5)              | Flag (`modulectl`)                            | Notes                                                         |
| ------------------------------------ | --------------------------------------------- | ------------------------------------------------------------- |
| `--module-config-file string`        | **Renamed** `-c, --config-file string`        | Path to your `module-config.yaml`                             |
| `--module-archive-path string`       | **Removed**                                   | Archive path for local module artifacts                       |
| `--module-archive-persistence`       | **Removed**                                   | Persist module archive on host filesystem                     |
| `--module-archive-version-overwrite` | **Renamed** `--overwrite`                     | Overwrite existing module OCI archive (**for testing only**)  |
| `--descriptor-version string`        | **Removed**                                   | Schema version for generated descriptor                       |
| `--git-remote string`                | **Removed**                                   | Git remote name for module sources                            |
| `--insecure`                         | `--insecure`                                  | Allow insecure registry connections                           |
| `--key string`                       | **Removed**                                   | Private key path for signing                                  |
| `--kubebuilder-project`              | **Removed**                                   | Indicate Kubebuilder project                                  |
| `-n, --name string`                  | `--name`                                      | Override module name                                          |
| `--name-mapping string`              | **Removed**                                   | OCM component name mapping                                    |
| `--namespace string`                 | `--namespace`                                 | Namespace for generated ModuleTemplate (default `kcp-system`) |
| `-o, --output string`                | `-o, --output string`                         | Output path for ModuleTemplate (default `template.yaml`)      |
| `-p, --path string`                  | **Removed**                                   | Path to module contents                                       |
| `-r, --registry string`              | `-r, --registry string`                       | Context URL for OCI registry                                  |
| `--registry-cred-selector string`    | `--registry-cred-selector string`             | Label selector for existing `dockerconfigjson` Secret         |
| `--registry-credentials string`      | `--registry-credentials string`               | Basic auth credentials in `<user:password>` format            |
| `--dry-run`                          | `--dry-run`                                   | Validate and skip pushing module descriptor                   |
| `-h, --help`                         | `-h, --help`                                  | Show help for create command                                  |

## 2. Module Configuration (`module-config.yaml`) Differences

This section illustrates how the `module-config.yaml` looks in the Kyma CLI format versus the `modulectl` format, with field-by-field mapping and examples.

## 2.1 Field Mapping Differences

| Kyma CLI                                       | ModuleCtl (new)                  | Description / Changes                                                                     |
|------------------------------------------------|--------------------------------- | ------------------------------------------------------------------------------------------|
| `name`                                         | `name`                           | Name of the module                                                                                                                                                          |
| `channel`                                      | **Removed**                      | Version to channel mapping moved to [ModuleReleaseMetadata](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md) |
| `version`                                      | `version`                        | Version of the module                                                                                                                                                       |
| `manifest`                                     | `manifest`                       | Manifest of the module. Previously local file → now must be a URL (e.g. GitHub release asset)                                                                               |
| `defaultCR`                                    | `defaultCR`                      | Default Module CR of the module. Previously local file → now must be a URL (e.g. GitHub release asset)                                                                      |
| `annotations.operator.kyma-project.io/doc-url` | `documentation`                  | Link to the module documentation. Moved from annotations map to top-level `documentation` key                                                                               |
| `moduleRepo`                                   | `repository`                     | Link to the repository of the module                                                                                                                                        |
| *n/a*                                          | **New** `icons`                  | List of module icons for the UI with `name`+`link`                                                                                                                          |
| `mandatory`                                    | `mandatory`                      | Marks the module as mandatory (default `false`)                                                                                                                             |
| *n/a*                                          |  **New** `requiresDowntime`      | Marks the module to require a maintenance window to update from previoius version (default `false`)                                                                         |
| `security`                                     | `security`                       | Path to security scanner config. Must be a local path and will be be resolved against the checked out version on Gitub                                                      |
| `labels` / `annotations`                       | `labels` / `annotations`         | Labels and annotations to put into the ModuleTemplate                                                                                                                       |
| *n/a*                                          | **New** `manager`                | Controller resource of the module (GVK, name, optional namespace)                                                                                                           |
| *n/a*                                          | **New** `associatedResources`    | List of GVKs to be cleaned up on uninstall                                                                                                                                  |
| *n/a*                                          | **New** `resources`              | Additional artifacts for the UI (e.g., CRDs)                                                                                                                                |
| `namespace`                                    | `namespace`                      | Target namespace for the generated ModuleTemplate (default `kcp-system`)                                                                                                    |
| `moduleRepoTag`                                | `moduleRepoTag`                  | Indicator for the pipeline to checkout the provided tag (default tag to checkout is the `version`)                                                                          |
| `beta`                                         | **Removed**                      | Marks the module as beta. Moved to [ModuleReleaseMeta](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md)                                                                                                                                |
| `internal`                                     | **Removed**                      | Marks the module as internal. Moved to [ModuleReleaseMeta](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md)                                                                                                                    |


## 2.2 Example:

### 2.2.1 Module Config using Kyma CLI

```yaml
# modules/<module-name>/<channel>/module-config.yaml
name: kyma-project.io/module/<module-name>
channel: <channel>
version: <version>
manifest: <module-name>-manifest.yaml
defaultCR: <module-name>-default-cr.yaml
annotations:
  operator.kyma-project.io/doc-url: https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-module-name
moduleRepo: https://github.com/kyma-project/module-manager.git
```

### 2.2.2 Module Config using `modulectl`

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

# 3. ModuleTemplate Differences 

## 3.1  ModuleTemplate Comparison

Below is a side-by-side comparison of the key fields in the `ModuleTemplate` YAML generated by Kyma CLI vs. modulectl:

| Field / Path                                                        | Kyma CLI–Generated (`kyma alpha create module`)      | modulectl–Generated (`modulectl create`)                    |
|---------------------------------------------------------------------|------------------------------------------------------|-------------------------------------------------------------|
| **metadata.name**                                                   | Present `<module-name>-<channel>`                    | **Changed** `<module-name>-<version>`                       |
| **metadata.namespace**                                              | Present                                              | Unchanged                                                   |
| **metadata.labels.operator.kyma-project.io/module-name**            | Present                                              | Unchanged                                                   |
| **metadata.annotations.operator.kyma-project.io/doc-url**           | Present                                              | **Removed**                                                 |
| **metadata.annotations.operator.kyma-project.io/is-cluster-scoped** | Present                                              | Unchanged                                                   |
| **metadata.annotations.operator.kyma-project.io/module-version**    | Present                                              | **Removed**                                                 |
| **spec.channel**                                                    | Present                                              | **Removed**                                                 |
| **spec.moduleName**                                                 | *n/a*                                                | **New**                                                     |
| **spec.version**                                                    | *n/a*                                                | **New**                                                     |
| **spec.mandatory**                                                  | Present                                              | Unchanged                                                   |
| **spec.requiresDowntime**                                           | *n/a*                                                | **New**                                                     |
| **spec.manager.name**                                               | *n/a*                                                | **New**                                                     |
| **spec.manager.namespace**                                          | *n/a*                                                | **New**                                                     |
| **spec.manager.group**                                              | *n/a*                                                | **New**                                                     |
| **spec.manager.version**                                            | *n/a*                                                | **New**                                                     |
| **spec.data**                                                       | Present                                              | Unchanged                                                   |
| **spec.info**                                                       | *n/a*                                                | **New**                                                     |
| **spec.descriptor.component**                                       | Present                                              | Unchanged                                                   |
| **spec.resources**                                                  | *n/a*                                                | **New**                                                     |

The following shows the `ModuleTemplate` YAML generated for TemplateOperator by Kyma CLI vs. `modulectl`.

## 3.1 Kyma CLI–Generated ModuleTemplate (channel-based)

```yaml
apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: template-operator-regular
  namespace: kcp-system
  labels:
    "operator.kyma-project.io/module-name": "template-operator"
  annotations:
    "operator.kyma-project.io/module-version": "0.0.1"
spec:
  mandatory: false
  channel: regular
  data:
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: Sample
    metadata:
      name: sample-yaml
    spec:
      resourceFilePath: "./module-data/yaml"
  descriptor:
    component:
      # ... component descriptor
    meta:
      schemaVersion: v2
```

## 3.2 `modulectl`–Generated ModuleTemplate (version-based)

```yaml
apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  annotations:
    operator.kyma-project.io/is-cluster-scoped: "false"
    operator.kyma-project.io/module-version: 0.0.1
  creationTimestamp: "2024-11-14T13:33:08Z"
  generation: 3
  labels:
    operator.kyma-project.io/module-name: template-operator
  name: template-operator-0.0.1
  namespace: kcp-system
  resourceVersion: "4369227537"
  uid: b00c8fcc-aaa2-43ce-bdab-ba82365e11b7
spec:
  associatedResources:
  - group: operator.kyma-project.io
    kind: Managed
    version: v1alpha1
  data:
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: Sample
    metadata:
      name: sample-yaml
    spec:
      resourceFilePath: ./module-data/yaml
  descriptor:
    component:
      # ... component descriptor
    meta:
      schemaVersion: v2
  info:
    documentation: https://github.com/kyma-project/template-operator/blob/main/README.md
    icons:
    - link: https://raw.githubusercontent.com/kyma-project/kyma/refs/heads/main/docs/assets/logo_icon.svg
      name: module-icon
    repository: https://github.com/kyma-project/template-operator
  manager:
    group: apps
    kind: Deployment
    name: template-operator-controller-manager
    namespace: template-operator-system
    version: v1
  mandatory: false
  moduleName: template-operator
  resources:
  - link: https://github.com/kyma-project/template-operator/releases/download/0.0.1/template-operator.yaml
    name: rawManifest
  version: 0.0.1
```



## Additional Resources

- [`modulectl` GitHub Repository](https://github.com/kyma-project/modulectl)
- [ADR: Iteratively moving forward with module requirements and aligning responsibilities](https://github.com/kyma-project/lifecycle-manager/issues/1681)
- [ModuleTemplate Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/03-moduletemplate.md)
- [ModuleReleaseMeta Custom Resource](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/contributor/resources/05-modulereleasemeta.md)
