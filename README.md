[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/modulectl)](https://api.reuse.software/info/github.com/kyma-project/modulectl)
# modulectl

## Overview
modulectl is a command line tool that supports developers of Kyma modules. It provides a set of commands and flags to:
* Create an empty scaffold for a new module
* Build a module and push it to a remote repository

## Installation

1. Download the binary for your operating system and architecture from the [GitHub releases page](https://github.com/kyma-project/modulectl/releases).
2. Move the binary to a directory in your PATH or navigate to the directory where the binary is located.
3. Make the binary executable by running `chmod +x modulectl`.

### Alternative
You can build the binary from the source code.

Clone the repository and run `make build` from the root directory of the repository.

The binary is created in the `bin` directory.

> [!NOTE]
>
> You can use Makefile targets for MacOS (darwin) or Linux, with the option to compile for x86 or ARM architectures, to build the binary for your specific operating system and architecture.

## Usage
```
modulectl <command> [flags]
```

### Available Commands
- `create` - Creates a module bundled as an OCI artifact. See [modulectl create](./docs/gen-docs/modulectl_create.md).
- `scaffold` - Generates necessary files required for module creation. See [modulectl scaffold](./docs/gen-docs/modulectl_scaffold.md)
- `help` - Provides help with any command.
- `version` - Prints the current version of the modulectl tool. See [modulectl version](./docs/gen-docs/modulectl_version.md).
- `completion` - Generates the autocompletion script for the specified shell.

For detailed information about the commands, you can use the `-h` or `--help` flag with the command. For example: `modulectl create -h`.

## Development

Before you start developing, create a local test setup. For more information, see [Configure and Run a Local Test Setup](./docs/contributor/local-test-setup.md).

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
