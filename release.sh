#!/bin/sh

set -eu

if [ $# -ne 1 ]; then
  echo "Requrire tag. e.g. v0.3.0"
  exit 1
fi

tag=$1
current_commit=$(git rev-parse HEAD)

# Actions が利用できるDocker Imageの制限に対応するため。 https://stackoverflow.com/questions/76403845/when-accessing-github-marketplace-actions-i-am-seeing-getting-error-should-be-e 
echo "tag: $tag"
echo "current_commit: $current_commit"

# Create working branch
git checkout master
git pull origin master
git checkout -b releaser/$current_commit

# Create Dockerfile
echo "FROM ghcr.io/rerost/issue-creator:${tag}" > Dockerfile.actions
git add Dockerfile.actions
git commit -m "Generate docker file"

# Create Tag
git tag -a $tag -m "$tag"
git push origin $tag
echo "RELASED"

# Clean
git checkout master
git branch -D releaser/$current_commit
echo "CLEANED"
