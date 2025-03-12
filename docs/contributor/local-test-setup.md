# Configure and Run a Local Test Setup

## Context

This tutorial shows how to set up and run local e2e tests.

## Prerequisites

Install the following tooling in the versions defined in [`versions.yaml`](https://github.com/kyma-project/lifecycle-manager/blob/main/versions.yaml):

- [Go](https://go.dev/)
- [k3d](https://k3d.io/stable/)

## Procedure

Follow the steps using scripts from the project root.

1. Create a local registry.

   ```sh
   ./scripts/re-create-test-registry.sh
   ```

2. Build modulectl.

   ```sh
   ./scripts/build-modulectl.sh
   ```

3. Run the `create` command tests.

   > :bulb: Re-running the `create` command requires to re-create the local registry.

   ```sh
   ./scripts/run-e2e-test.sh --cmd=create
   ```

4. Run the `scaffold` command tests.

   ```sh
   ./scripts/run-e2e-test.sh --cmd=scaffold
   ```