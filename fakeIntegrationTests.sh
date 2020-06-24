#!/bin/bash

if [ -z "$1" ]; then
  echo "No commit hash provided." >&2
  exit 1
fi

hash="$1"; shift

message="Docu update, no integration tests needed."
if [ ! -z "$1" ]; then
  message="$1"
fi

curl -X POST \
        --fail \
        --data "{\"state\": \"success\", \"target_url\": \"https://travis-ci.org/SAP/jenkins-library\", \"description\": \"${message}\", \"context\": \"integration-tests\"}" \
        --netrc \
    "https://api.github.com/repos/SAP/jenkins-library/statuses/${hash}"


if [ -z "$2" ]; then
  echo "No github run id provided, will no perform status update of 'Go / integration-tests'." >&2
  echo "Otherwise: the github run id can be optained from the url of another github action run." >&2
  exit 0
fi

githubrunid = "$2"

curl -X POST \
        --fail \
        --data "{\"state\": \"success\", \"target_url\": \"https://github.com/SAP/jenkins-library/actions/runs/${githubrunid}\", \"description\": \"${message}\", \"context\": \"Go / integration-tests\"}" \
        --netrc \
    "https://api.github.com/repos/SAP/jenkins-library/statuses/${hash}"
