#!/bin/sh

set -ex

url=$1
close_last_issue=$4
check_before_create_issue=$5
token=$6
if [ -z $url ]; then
  repository=$2
  template_issue=$3
  is_discussion=$7

  type="issues"

  if [ "${is_discussion}" = "true" ]; then
    type="discussions"
  fi

  url=https://github.com/${repository}/${type}/${template_issue}
fi
echo ${url}

/issue-creator \
  create \
  ${url} \
  --CloseLastIssue=${close_last_issue} \
  --check-before-create-issue="${check_before_create_issue}" \
  --GithubAccessToken=${token}
