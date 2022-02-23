GO111MODULE=on go get github.com/freshautomations/stoml

latestFluxVersion="$(curl -s --request GET --url "https://api.github.com/repos/fluxcd/flux2/releases?per_page=1" | jq . | jq '.[0] | .tag_name' | jq -r | sed -e 's/v//')"

currentFluxVersion="$(stoml tools/dependencies.toml flux.version)"

echo "Latest flux version $latestFluxVersion"
echo "Current flux version $currentFluxVersion"

if [[ "$latestFluxVersion" == "$currentFluxVersion" ]]; then
  echo "No changes needed"
  exit 0
fi

newContent=$(cat tools/dependencies.toml | tr '\n' '\r' | sed -e "/\[flux\]\rversion\=\"[0-9\.]*\"/{s//\[flux\]\rversion=\"$latestFluxVersion\"/;:a" -e '$!N;$!ba' -e '}' | tr '\r' '\n')

echo "$newContent" > tools/dependencies.toml

git --no-pager diff

PR_BODY="**What changed?**
Upgrade flux version from ${currentFluxVersion} to ${latestFluxVersion}

**Why?**
To be up to date with latest flux2 releases

**How did you test it?**
Test suites on CI

**Release notes**

**Documentation Changes**"

echo "--BODY--"
echo "$PR_BODY"

# This adds proper break lines so they can be passed in properly to the Github action creating the
PR_BODY="${PR_BODY//'%'/'%25'}"
PR_BODY="${PR_BODY//$'\n'/'%0A'}"
PR_BODY="${PR_BODY//$'\r'/'%0D'}"

PR_TITLE="Upgrade flux version from ${currentFluxVersion} to ${latestFluxVersion}"

echo "--TITLE--"
echo "$PR_TITLE"
