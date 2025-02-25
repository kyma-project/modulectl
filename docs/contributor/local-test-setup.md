# Configure and Run a Local Test Setup

## Context

This tutorial shows how to setup and run local e2e tests.

## Prerequisites

Install the following tooling in the versions defined in [`versions.yaml`](../../versions.yaml):

- [Go](https://go.dev/)
- [k3d](https://k3d.io/stable/)

## Procedure

Follow the steps using scripts from the project root.

### 1. Create a Local Registry

Create a local registry.

```sh
./scripts/re-create-test-registry.sh
```

### 2. Build modulectl

```sh
./scripts/build-modulectl.sh
```

### 3. Run the CREATE Command Tests

> :bulb: Re-running the Create command requires to re-create to local registry.

```sh
./scripts/run-e2e-test.sh --cmd=create
```

### 4. Run the SCAFFOLD Command Tests

```sh
./scripts/run-e2e-test.sh --cmd=scaffold
```
