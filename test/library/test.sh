#!/bin/bash
set -e
set -x

git clone https://github.com/weaveworks-gitops-test/wego-library-test.git test/library/wego-library-test
cd test/library/wego-library-test
yarn install --frozen-lockfile
yarn build
go mod tidy
go build main.go
