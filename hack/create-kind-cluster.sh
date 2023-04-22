#!/usr/bin/env bash

set -e
set -o errexit
set -o nounset
set -o pipefail

script_name=$0
script_full_path=$(dirname "$0")

kind create cluster --config=${script_full_path}/kind-config.yaml