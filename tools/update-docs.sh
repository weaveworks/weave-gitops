#!/bin/bash
# Note that GITOPS_VERSION and ALGOLIA_API_KEY environment variables must be set
# before running this script

WEAVE_GITOPS_BINARY=$1
WEAVE_GITOPS_DOC_REPO=$2

cd $WEAVE_GITOPS_DOC_REPO/docs
yarn install
# update version information
sed -i
ex - installation.mdx << EOS
/download\/
%s,download/\([^/]*\)/,download/${GITOPS_VERSION}/,
/Current Version
.,+3! ${WEAVE_GITOPS_BINARY} version
wq!
EOS
# create CLI reference
git rm -f --ignore-unmatch cli-reference.md
git rm -rf --ignore-unmatch cli-reference
mkdir -p cli-reference
cd cli-reference
git checkout _category_.json
${WEAVE_GITOPS_BINARY} docs
git add *.md
# create versioned docs
cd $WEAVE_GITOPS_DOC_REPO
npm run docusaurus docs:version $GITOPS_VERSION
