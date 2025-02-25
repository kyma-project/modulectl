#!/bin/bash

# Changing current directory to the root of the project
cd $(git rev-parse --show-toplevel)

while [ $# -gt 0 ]; do
  case "$1" in
    --cmd=*)
      cmd="${1#*=}"
      ;;
  esac
  shift
done

export PATH=$(pwd)/bin/:$PATH && make -C ./tests/e2e test-${cmd}-cmd
