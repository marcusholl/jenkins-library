#!/bin/bash

# This script uploads an asset to an already existing release
#
# PREREQUISITES
#  - jq
# a repo on github.com with an already existing release
# PARAMETERS provided to the script in this order
#  - repo_slug e.g. marcusholl/playground
#  - release
#  asset_name
#  asset_label

DEBUG_FLAGS="-v"

REPO_SLUG=$1; shift
RELEASE=$1; shift
ASSET_NAME=$1; shift
ASSET_LABEL=$1; shift

[ -z "${REPO_SLUG}" ] && { echo "REPO_SLUG missing" > /dev/stderr; exit 1; }
[ -z "${RELEASE}" ] && { echo "RELEAE missing" > /dev/stderr; exit 1; }
[ -z "${ASSET_NAME}" ] && { echo "ASSET_NAME missing" > /dev/stderr; exit 1; }
[ -z "${ASSET_LABEL}" ] && { echo "ASSET_LABEL missing" > /dev/stderr; exit 1; }

export ASSET_URL=$(curl -netrc  https://api.github.com/repos/${REPO_SLUG}/releases | jq --raw-output ".[]| select(.tag_name == \"${RELEASE}\") | .assets |select(.[].name == \"${ASSET_NAME}\") |.[].url")

echo "[INFO] asset url: ${ASSET_URL}"

if [ ! -z "${ASSET_URL}" ];then
    echo "[INFO] Deleting already existing asset" > /dev/stderr
    curl $DEBUG_FLAGS --netrc --request DELETE "${ASSET_URL}" || exit 1
else
    echo "[INFO] There is currently no such asset." > /dev/stderr
fi

export UPLOAD_URL=$(curl -netrc  https://api.github.com/repos/${REPO_SLUG}/releases | jq --raw-output ".[] |select(.tag_name == \"v0.0.0\") |.upload_url")
UPLOAD_URL=$(echo "${UPLOAD_URL}" |sed "s/{?name,label}/?name=${ASSET_NAME}\&label=${ASSET_LABEL}/g")

curl ${DEBUG_FLAGS} -netrc --data-binary @piper --header "Content-Type: application/zip"  "${UPLOAD_URL}"
