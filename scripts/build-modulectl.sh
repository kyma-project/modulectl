#!/bin/bash

# Changing current directory to the root of the project
cd $(git rev-parse --show-toplevel)

make build-darwin-arm
cp ./bin/modulectl-darwin-arm ./bin/modulectl
