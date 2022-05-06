#!/bin/bash
# Copied from https://raw.githubusercontent.com/snyk-tech-services/snyk-delta/develop/snyk_delta_all_projects.sh

# Call this script as you would call snyk test | snyk-delta, minus the --all-projects and --json flags
# This is an interim fix until snyk-delta supports all projects itself (or snyk supports a --new flag)
# example: /bin/bash snyk_delta_all_projects.sh --severity=high --exclude=tests,resources -- -s config.yaml

# runs snyk test --all-projects --json $*

# this will break if there is only one project

# requires jq is installed as well

set -euo pipefail

exit_code=0

echo 'Running snyk-delta-all-projects'


for test in `snyk test --all-projects --json $* | jq -r '.[] | @base64'`; do
#    echo ${test} | base64 --decode
    echo ${test} | base64 --decode | snyk-delta -d
    project_exit_code=$?
    exit_code+=$project_exit_code
    if [ $project_exit_code -eq 2 ]
    then
        echo 'retry'
        echo ${test} | base64 --decode | snyk-delta -d
    fi
    project="$(echo ${test} | base64 --decode | jq -r '.displayTargetFile')"
    echo "project: ${project} exit code: ${project_exit_code}"
done

echo "Final Exit Code: ${exit_code}"
exit $exit_code
