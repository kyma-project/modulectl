#!/bin/bash

if k3d registry list | grep -q "^k3d-oci.localhost\s"; then
  ./scripts/delete-test-registry.sh
fi

k3d registry create oci.localhost --port 5001
