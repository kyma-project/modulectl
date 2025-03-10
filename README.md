[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/modulectl)](https://api.reuse.software/info/github.com/kyma-project/modulectl)
# `modulectl`

## Overview
It is a command line tool which supports Module developers for Kyma. It provides a set of commands and flags which can
be used to:
- Create an empty scaffold for a new module
- Build a module and push it to a remote repository

## How to Install

From the GitHub releases page (https://github.com/kyma-project/modulectl/releases), download the binary for your operating system and architecture.
Then, move the binary to a directory in your PATH.
Or go inside the directory where the binary is located, and you can run the binary from there.
Don't forget to make the binary executable by running `chmod +x modulectl`.

### Alternative
You can build the binary from the source code.
Clone the repository and run `make build` from the root directory of the repository.
The binary will be created in the `bin` directory.
> **Note**
>
> We also provide specific Makefile targets for MacOS (darwin) & Linux operating systems, with the options to compile
> for x86 or ARM architectures.
> You can use them to build the binary for your specific operating system and architecture.

## Usage
```
modulectl <command> [flags]
```

### Available Commands
- `create` - Creates a module bundled as an OCI artifact. Detailed long explanation can be found here.
- `scaffold` - Generates necessary files required for module creation.
- `help` - Provides help about any command.
- `version` - Print the version of the `modulectl` tool.
- `completion` - Generate the autocompletion script for the specified shell.

For detailed information about the commands, you can use the `-h` or `--help` flag with the command.
For example: `modulectl create -h`.

Below are links to the detailed explanation of the commands, in case you want to know more about them without actually
running the commands:
- [Create Command](./docs/gen-docs/modulectl_create.md)
- [Scaffold Command](./docs/gen-docs/modulectl_scaffold.md)
- [Version Command](./docs/gen-docs/modulectl_version.md)

## Development

Before you start developing, a local test setup must be created.
To make life easy, we have written some scripts, which automate most of the process.
Please follow [this guide](./docs/contributor/local-test-setup.md) to know more.

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
