#!/bin/bash

set -e

version="v1.24"
target="v1.24.0"
# this will be the branch we use to generate carry commits
branch="rebase-${target}"

if git show-ref --verify --quiet refs/heads/${branch}; then
  echo "branch [${branch}] already exists"
  exit 1
fi

git checkout master
git fetch upstream
git fetch openshift

git checkout -b ${branch} ${target}
git merge -s ours -m "Merge remote-tracking branch 'openshift/master' into ${branch} openshift-rebase(${version}):marker"  openshift/master
