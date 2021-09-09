#!/bin/sh -l

repository=$1
template_issue=$2
close_last_issue=$3
check_before_create_issue=$4
token=$5
is_discussion=$6

type="issues"

if [ ${is_discussion} = "true" ]; then
  type="discussions"
fi

url=https://github.com/${repository}/${type}/${template_issue}
echo ${url}

issue-creator \
  create \
  ${url} \
  --CloseLastIssue=${close_last_issue} \
  --check-before-create-issue="${check_before_create_issue}" \
  --GithubAccessToken=${token}
