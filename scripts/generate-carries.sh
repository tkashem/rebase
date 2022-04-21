#!/bin/bash

set -ex

target="v1.24.0-rc.0"
saveto="carry-commits-${target}.log"
# this is used as the base for carry commits.
base="v1.23.0"

# this will be the branch we use to generate carry commits
branch="generate-carry-commits-${target}"

git checkout master
git fetch upstream
git fetch openshift

if git show-ref --verify --quiet refs/heads/${branch}; then
  git branch -D ${branch}
fi

git checkout -b ${branch} ${target}

git merge -s ours -m "Merge remote-tracking branch 'openshift/master' into ${branch}" openshift/master

git log $(git merge-base openshift/master ${base})..openshift/master --ancestry-path --reverse --no-merges \
--pretty='tformat:%x09%h%x09%x09%x09%s%x09https://github.com/openshift/kubernetes/commit/%h?w=1' | \
grep -E $'\t''UPSTREAM: .*'$'\t' | \
sed -E 's~UPSTREAM: ([0-9]+)(:.*)~UPSTREAM: \1\2\thttps://github.com/kubernetes/kubernetes/pull/\1~' > ${saveto}
