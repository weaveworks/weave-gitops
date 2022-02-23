

# Cache system of gitproviders pkg

# How it works

This cache system works through Cassettes. A Cassette is a yaml file that stores all the interactions made towards the github api where that same Cassette was involved. This file is generated the first time the tests using the same cassette are run. If the file exists, that cassette will be used to emulate the previously recorded responses.

## How to implement a Cassette recorder

### Set up environment variables

- `GITHUB_TOKEN` is used for authentication.
- `GITHUB_ORG` is used to specify the Github organization we want to test with.
- `GITHUB_USER` is used to specify the Github account we want to test with.

Note: Only GITHUB_TOKEN is a must, the other can be set as needed.

### Set up a recorder

To initialize the recorder, use the function `getTestClientWithCassette`:
```go
client, recorder, err := getTestClientWithCassette("CASSETTE_NAME", "provider name")
```

- `client` is the object from `fluxcd/go-git-providers` that was injected with the recorder and makes the api calls.
- `recorder` is the object from `dnaeon/go-vcr` that helps to save the yaml file which is triggered by using `recorder.Done()`.
- `CASSETTE_NAME` name of the Cassette.

### Complete implementation example using ginkgo

```go
var _ = Describe("create github repo", func() {
	accounts := getAccounts()

    var gitProvider defaultGitProvider
	var recorder *recorder.Recorder
	var err error
	BeforeEach(func() {
	    var client gitprovider.Client
		client, recorder, err = getTestClientWithCassette("CASSETTE_ID", "provider name")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			provider: client,
		}
	})

	It("should create a repository successfully", func() {
		err = gitProvider.CreateRepository(repoName, accounts.GithubOrgName, true)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	    // delete org repo
		ctx := context.Background()
		orgRepoRef := NewOrgRepositoryRef(github.DefaultDomain, accounts.GithubOrgName, repoName)
		org, err := client.OrgRepositories().Get(ctx, orgRepoRef)
		Expect(err).NotTo(HaveOccurred())
		err = org.Delete(ctx)
		Expect(err).NotTo(HaveOccurred())
		// save Cassette in pkg/gitproviders/cache/CASSETTE_ID.yaml
		err = recorder.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

})
```
## Troubleshooting

- If you face the error `Requested interaction not found` it means there is an api call that
was not saved in the Cassette when it was created.
- If you added more interactions to an existing test you will need to regenerate the Cassette yaml. To do so remove yaml file and rerun the test. Otherwise, you will get the previous error.
