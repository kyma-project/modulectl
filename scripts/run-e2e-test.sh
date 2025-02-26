#!/bin/bash

# Changing current directory to the root of the project
cd $(git rev-parse --show-toplevel)

while [ $# -gt 0 ]; do
  case "$1" in
    --cmd=*)
      CMD="${1#*=}"
      ;;
  esac
  shift
done

# Check for mandatory arguments
if [[ -z "$CMD" ]]; then
  echo "[$(basename $0)] Missing required arguments"
  echo "Usage: $(basename $0) --cmd=<create|scaffold>"
  exit 1
fi



export PATH=$(pwd)/bin/:$PATH && make -C ./tests/e2e test-${CMD}-cmd
