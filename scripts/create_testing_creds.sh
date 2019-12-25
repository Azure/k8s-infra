#!/usr/bin/env bash

# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

set -o errexit
set -o nounset
set -o pipefail

# Enable tracing in this script off by setting the TRACE variable in your
# environment to any value:
#
# $ TRACE=1 test.sh
TRACE=${TRACE:-""}
if [[ -n "${TRACE}" ]]; then
  set -x
fi

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

json=$(az ad sp create-for-rbac -o json)
client_id=$(echo "$json" | jq -r '.appId')
client_secret=$(echo "$json" | jq -r '.password')
tenant_id=$(echo "$json" | jq -r '.tenant')
subscription_id=$(az account show -o json | jq -r ".id")

echo -e "export AZURE_TENANT_ID=$tenant_id\nexport AZURE_CLIENT_ID=$client_id\nexport AZURE_CLIENT_SECRET=$client_secret\nexport AZURE_SUBSCRIPTION_ID=$subscription_id" > "$REPO_ROOT"/.env