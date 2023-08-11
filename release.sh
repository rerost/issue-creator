#!/bin/sh

set -eu

if [ $# -ne 1 ]; then
  echo "Requrire tag. e.g. v0.3.0"
  exit 1
fi

tag=$1

# Actions が利用できるDocker Imageの制限に対応するため。 https://stackoverflow.com/questions/76403845/when-accessing-github-marketplace-actions-i-am-seeing-getting-error-should-be-e 
echo "tag: $tag"
git checkout master
git pull origin master
# Only mac
sed -i '' "s/TAG/$tag/g" Dockerfile.actions
git add Dockerfile.actions
git commit -m "Generate docker file"
git push origin master
git tag -a $tag -m "$tag"
git push origin $tag
