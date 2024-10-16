#!/usr/bin/env bash

set -o nounset
set -o errexit
set -E
set -o pipefail

uploadFile() {
  filePath=${1}
  ghAsset=${2}

  echo "Uploading ${filePath} as ${ghAsset}"
  response=$(curl -s -o output.txt -w "%{http_code}" \
                  --request POST --data-binary @"$filePath" \
                  -H "Authorization: token $GITHUB_TOKEN" \
                  -H "Content-Type: text/yaml" \
                   "$ghAsset")
  if [[ "$response" != "201" ]]; then
    echo "Unable to upload the asset ($filePath): "
    echo "HTTP Status: $response"
    cat output.txt
    exit 1
  else
    echo "$filePath uploaded"
  fi
}

RELEASE_TAG=$1

echo "Fetching releases"
CURL_RESPONSE=$(curl -w "%{http_code}" -sL \
                -H "Accept: application/vnd.github+json" \
                -H "Authorization: Bearer $GITHUB_TOKEN"\
                "$GITHUB_URL"/releases)
JSON_RESPONSE=$(sed '$ d' <<< "${CURL_RESPONSE}")
HTTP_CODE=$(tail -n1 <<< "${CURL_RESPONSE}")
if [[ "${HTTP_CODE}" != "200" ]]; then
  echo "${CURL_RESPONSE}"
  exit 1
fi

echo "Finding release id for: ${RELEASE_TAG}"
RELEASE_ID=$(jq <<< "${JSON_RESPONSE}" --arg tag "${RELEASE_TAG}" '.[] | select(.tag_name == $ARGS.named.tag) | .id')

echo "Got '${RELEASE_ID}' release id"
if [ -z "${RELEASE_ID}" ]
then
  echo "No release with tag = ${RELEASE_TAG}"
  exit 1
fi

echo "Adding artifacts to Github release"
UPLOAD_URL="https://uploads.github.com/repos/kyma-project/modulectl/releases/${RELEASE_ID}/assets"
uploadFile "bin/modulectl-linux" "${UPLOAD_URL}?name=modulectl-linux"
uploadFile "bin/modulectl-linux-arm" "${UPLOAD_URL}?name=modulectl-linux-arm"
uploadFile "bin/modulectl-darwin" "${UPLOAD_URL}?name=modulectl-darwin"
uploadFile "bin/modulectl-darwin-arm" "${UPLOAD_URL}?name=modulectl-darwin-arm"
