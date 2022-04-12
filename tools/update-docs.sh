# Note that GITOPS_VERSION and ALGOLIA_API_KEY environment variables must be set
# before running this script

WEAVE_GITOPS_BINARY=$1
WEAVE_GITOPS_DOC_REPO=$2

cd $WEAVE_GITOPS_DOC_REPO/docs
yarn install
# update version information
ex - installation.mdx << EOS
/download\/
%s,download/\(.*\)/,download/${GITOPS_VERSION}/,
/Current Version
.,+4! ${WEAVE_GITOPS_BINARY} version
+5d
wq!
EOS
# create CLI reference
git rm -f --ignore-unmatch cli-reference.md
git rm -rf --ignore-unmatch cli-reference
mkdir -p cli-reference
cd cli-reference
ex - _category_.json << EOS
1i
{
  "label": "CLI Reference",
  "position": 3
}
.
wq!
EOS
${WEAVE_GITOPS_BINARY} docs
git add *.md
# create versioned docs
cd $WEAVE_GITOPS_DOC_REPO
version_number=$(cut -f2 -d'v' <<< $GITOPS_VERSION)
npm run docusaurus docs:version $version_number
