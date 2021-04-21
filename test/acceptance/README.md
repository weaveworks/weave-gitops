# WEGO User Acceptance Tests

This suite contains the user acceptance tests for the Weave GitOps. To run these tests you can either use gingko runner or standard go test command .

By default test harness assumes that WEGO binary is available on `$PATH` but this can be overriden by exporting the following variable


```
export WEGO_BIN_PATH=<path/to/wego-binary>
```

# Smoke Tests

To run the **smoke tests** from the suite, run the following the command from the repo root directory.

```
ginkgo -v -tags=smoke ./test/acceptance/test/...
```
Or

```
go test -v -tags=smoke ./test/acceptance/test/...
```
# Acceptance Tests
To run the full **acceptance suite**, run the command


```
ginkgo -v -tags=acceptance ./test/acceptance/test/...
```
Or 
```
go test -v -tags=acceptance ./test/acceptance/test/...
```

# How to add new test

Smoke test can be added to `smoke_tests.go` or create a new go file with smoke as build tag.

For non smoke tests, feel free to create appropriately named go file.

This suite follows the **BDD** gherkin style specs, when adding a new test, make every effort to adhere to `Given-When-Then` semantics. 
