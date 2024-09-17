---
title: modulectl create
---

Creates a module bundled as an OCI artifact.

## Synopsis

Use this command to create a Kyma module, bundle it as an OCI artifact, and push it to the OCI registry optionally.

### Detailed description

This command allows you to create a Kyma module as an OCI artifact and optionally push it to the OCI registry of your choice.
For more information about Kyma modules see the [documentation](https://kyma-project.io/#/06-modules/README).

This command creates a module from an existing directory containing the module's source files.
The directory must be a valid git project that is publicly available.
The command supports just one type of directory layout for the module:
- Simple: Just a directory with a valid git configuration. All the module's sources are defined in this directory.
Such projects require providing an explicit path to the module's project directory using the "--path" flag or invoking the command from within that directory.

### Simple mode configuration

To configure the simple mode, provide the `--module-config-file` flag with a config file path.
The module config file is a YAML file used to configure the following attributes for the module:

```yaml
- name:             a string, required, the name of the module
- version:          a string, required, the version of the module
- channel:          a string, required, channel that should be used in the ModuleTemplate CR
- mandatory:        a boolean, optional, default=false, indicates whether the module is mandatory to be installed on all clusters
- manifest:         a string, required, reference to the manifest, must be a relative file name
- defaultCR:        a string, optional, reference to a YAML file containing the default CR for the module, must be a relative file name
- resourceName:     a string, optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate CR that will be created
- security:         a string, optional, name of the security scanners config file
- internal:         a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the internal flag or not
- beta:             a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the beta flag or not
- labels:           a map with string keys and values, optional, additional labels for the generated ModuleTemplate CR
- annotations:      a map with string keys and values, optional, additional annotations for the generated ModuleTemplate CR
```

The **manifest** and **defaultCR** paths are resolved against the module's directory, as configured with the "--path" flag.
The **manifest** file contains all the module's resources in a single, multi-document YAML file. These resources will be created in the Kyma cluster when the module is activated.
The **defaultCR** file contains a default custom resource for the module that will be installed along with the module.
The Default CR is additionally schema-validated against the Custom Resource Definition. The CRD used for the validation must exist in the set of the module's resources.

### Modules as OCI artifacts
Modules are built and distributed as OCI artifacts. 
This command creates a component descriptor in the configured descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI artifact.
The internal structure of the artifact conforms to the [Open Component Model](https://ocm.software/) scheme version 3.

If you configured the "--registry" flag, the created module is validated and pushed to the configured registry.
During the validation the **defaultCR** resource, if defined, is validated against a corresponding CustomResourceDefinition.
You can also trigger an on-demand **defaultCR** validation with "--validateCR=true", in case you don't push the module to the registry.

#### Name Mapping
To push the artifact into some registries, for example, the central docker.io registry, you have to change the OCM Component Name Mapping with the following flag: "--name-mapping=sha256-digest". This is necessary because the registry does not accept artifact URLs with more than two path segments, and such URLs are generated with the default name mapping: **urlPath**. In the case of the "sha256-digest" mapping, the artifact URL contains just a sha256 digest of the full Component Name and fits the path length restrictions. The downside of the "sha256-mapping" is that the module name is no longer visible in the artifact URL, as it contains the sha256 digest of the defined name.

```bash
modulectl create [--module-config-file MODULE_CONFIG_FILE | --name MODULE_NAME --version MODULE_VERSION] [--path MODULE_DIRECTORY] [--registry MODULE_REGISTRY] [flags]
```

## Examples

```bash
Examples:
Build a simple module and push it to a remote registry
		modulectl create --module-config-file=/path/to/module-config-file --registry http://localhost:5001/unsigned --insecure
```

## Flags

```bash
-c, --credentials string              Basic authentication credentials for the given repository in the <user:password> format.
    --git-remote string               Specifies the URL of the wanted GitHub repository.
-h, --help                            Provides help for the create command.
    --insecure                        Uses an insecure connection to access the registry.
    --module-config-file string       Specifies the module configuration file.
-o, --output string                   File to write the module template if the module is uploaded to a registry (default "template.yaml").
    --registry string                 Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.
    --registry-cred-selector string   Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".
```

## See also

* [modulectl](modulectl.md)	 - This is the Kyma Module Controller CLI.


