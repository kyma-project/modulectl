---
title: modulectl create
---

Creates a module bundled as an OCI artifact.

## Synopsis

Use this command to create a Kyma module, bundle it as an OCI artifact, and push it to the OCI registry optionally.

### Detailed description

This command allows you to create a Kyma module as an OCI artifact and optionally push it to the OCI registry of your choice.
For more information about Kyma modules see the [documentation](https://kyma-project.io/#/06-modules/README).

### Configuration

Provide the `--config-file` flag with a config file path.
The module config file is a YAML file used to configure the following attributes for the module:

```yaml
- name:             a string, required, the name of the module
- version:          a string, required, the version of the module
- channel:          a string, required, channel that should be used in the ModuleTemplate CR
- manifest:         a string, required, reference to the manifest, must be a relative file name or URL

- defaultCR:        a string, optional, reference to a YAML file containing the default CR for the module, must be a relative file name or URL
- mandatory:        a boolean, optional, default=false, indicates whether the module is mandatory to be installed on all clusters
- resourceName:     a string, optional, default={NAME}-{CHANNEL}, the name for the ModuleTemplate CR that will be created
- internal:         a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the internal flag or not
- beta:             a boolean, optional, default=false, determines whether the ModuleTemplate CR should have the beta flag or not
- security:         a string, optional, name of the security scanners config file
- labels:           a map with string keys and values, optional, additional labels for the generated ModuleTemplate CR
- annotations:      a map with string keys and values, optional, additional annotations for the generated ModuleTemplate CR
```

The **manifest** file contains all the module's resources in a single, multi-document YAML file. These resources will be created in the Kyma cluster when the module is activated.
The **defaultCR** file contains a default custom resource for the module that is installed along with the module. It is additionally schema-validated against the Custom Resource Definition.
The CRD used for the validation must exist in the set of the module's resources.

### Modules as OCI artifacts
Modules are built and distributed as OCI artifacts. 
This command creates a component descriptor in the configured descriptor path (./mod as a default) and packages all the contents on the provided path as an OCI artifact.
The internal structure of the artifact conforms to the [Open Component Model](https://ocm.software/) scheme version 3.

If you configured the "--registry" flag, the created module is validated and pushed to the configured registry.


```bash
modulectl create [--config-file MODULE_CONFIG_FILE] [--registry MODULE_REGISTRY] [flags]
```

## Examples

```bash
Build a simple module and push it to a remote registry
		modulectl create --module-config-file=/path/to/module-config-file --registry http://localhost:5001/unsigned --insecure
```

## Flags

```bash
-c, --config-file string              Specifies the path to the module configuration file.
    --git-remote string               Specifies the URL of the module's GitHub repository. 
-h, --help                            Provides help for the create command.
    --insecure                        Uses an insecure connection to access the registry.
-o, --output string                   Path to write the ModuleTemplate file to, if the module is uploaded to a registry (default "template.yaml").
-r, --registry string                 Context URL of the repository. The repository URL will be automatically added to the repository contexts in the module descriptor.
    --registry-cred-selector string   Label selector to identify an externally created Secret of type "kubernetes.io/dockerconfigjson". It allows the image to be accessed in private image registries. It can be used when you push your module to a registry with authenticated access. For example, "label1=value1,label2=value2".
    --registry-credentials string     Basic authentication credentials for the given repository in the <user:password> format.
```

## See also

* [modulectl](modulectl.md)	 - This is the Kyma Module Controller CLI.


