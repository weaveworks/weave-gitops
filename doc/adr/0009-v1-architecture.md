# 9. v1 Architecture

Date: 2021-10-07

## Status

Accepted

## Context

As we begin to implement more features and integrate with more projects, it will be crucial for us to agree to a desired architecture for Weave GitOps Core. This document seeks to lay out the aspirational architecture for a conceptual "v1" of the project.

The goals for this proposed architecture are as follows.

It should:

- Separate data from business logic
- Be maintainable (and therefore testable)
- Be composable
- Be easy to understand
- Be a repeatable pattern

Here are some repositories that provided context and inspiration for our decision:

- https://github.com/gitops-tools/image-updater
- https://github.com/gitops-tools/tekton-ci
- https://github.com/weaveworks/reignite

## Decision

We have decided to go with an approach that we are calling "Hexagonal Lite". Much of the proposal consists of using the underlaying pricipals of Hexagonal Architecture without trying to exactly match the canonical nomeclature. This approach also borrows heavily from "Domain Driven Design", but again does not seek to match all of the canonical naming conventions of DDD.

Here are the proposed parts of the codebase:

### Models

- The `models` package contains data structures for the different WeGO domain entities. These are generally `structs` that do not contain business logic.
- Model objects may have helper methods for common situations, such as converting between types, accessing a computed property, etc.
- Models are typically thought of as "immutable", meaning that methods that mutates a model should return a new copy of that model (pass by value).

### Services

- Services represent the business logic of our program
- Services typically consume models and interact with adapters to execute operations
- Services should seek to abstract the implementation details of Adapters
- Services will often consume other services
- Services are generally named based on what they do: `Applier`, `Creator`, `Fetcher`

### Adapters

- Adapters are the "machinery" that is used to interact with the outside world
- Adapters can be incoming or outgoing
- Models do not use Adapters directly
- An example incoming adapter would be the HTTP API server
- An example outgoing adapter would be a Kubernetes client or [go-git-providers](https://github.com/fluxcd/go-git-providers) client.

### Example Implementation (psuedo-code)

Here is an example of what an "add application" operation might look like in a psuedo-code implementation. Extra comments have been added for the purposes of this doc:

```golang
// ./services/application/add.go
package application

// ...

// The Adder service interface
type Adder interface {
	Add(app models.Application, cl models.Cluster, params AddParams) error
}

// A constructor to return a new Adder service
func NewAdder(gs gitrepo.AppComitter, prs pullrequest.Manager, cs cluster.Applier, dks deploykey.Manager) Adder {
	return addService{
		gs:  gs,
		prs: prs,
		cs:  cs,
		dks: dks,
	}
}

// A struct that implements the Adder interface
type addService struct {
	gs  gitrepo.AppComitter
	prs pullrequest.Manager
	cs  cluster.Applier
	dks deploykey.Manager
}

// A params struct is recommended when a function signature would otherwise require more than 4 arguments
type AddParams struct {
	AutoMerge bool
	Token     string
}

// Notice that we accept models here
func (a addService) Add(app models.Application, cl models.Cluster, params AddParams) error {
	// destRepo is an instance of a models.GitRepository
	destRepo := models.NewGitRepoFromURL(app.ConfigRepoURL)

	dk, err := a.dks.Fetch(cl, app)
	if err != nil {
		return err
	}

	if err := a.gs.CommitApplication(destRepo, dk, app); err != nil {
		return err
	}

	pr, err := a.prs.Create(destRepo, params.Token, "main")
	if err != nil {
		return err
	}

	if params.AutoMerge {
		if err := a.prs.Merge(pr, params.Token); err != nil {
			return err
		}
	}

	if err := a.cs.ApplyApplication(cl, app); err != nil {
		return err
	}

	return nil
}

```

### Weave GitOps Repository Directory Structure

Go packages are coupled with the file system in such a way that it is neccessary to propose an example directory structure. Note that this is an example only, and does not seek to enumerate all files or reflect the actual packages that will exist. Here is an example of what the repo might look like:

```
├── cmd
│   ├── gitops
│   │   ├── add
│   │   │   └── main.go
│   │   ├── get
│   │   │   └── main.go
│   │   └── get-commits
│   │       └── main.go
│   └── gitops-server
│       └── main.go
├── go.mod
├── go.sum
├── pkg
│   ├── adapters
│   │   └── server
│   │       └── handlers.go
│   ├── models
│   │   ├── application.go
│   │   ├── cluster.go
│   │   ├── commit.go
│   │   ├── deploykey.go
│   │   ├── gitrepo.go
│   │   └── pullrequest.go
│   └── services
│       ├── application
│       │   ├── add.go
│       │   └── get.go
│       ├── cluster
│       │   └── cluster.go
│       ├── commit
│       │   └── commit.go
│       ├── deploykey
│       │   └── deploykey.go
│       ├── gitrepo
│       │   └── gitrepo.go
│       └── pullrequest
│           └── pullrequest.go
└── README.md

```

### FAQ

Q: **When does a function belong on the model and when does it belong in a services?**
A: Models reciever methods should be focused on performing common data manipulation operations on their parent object. An example of a good use of model methods would be a method that converts the model object into another type. An example of a bad use of model methods would be a method that retrieves something from a cluster or external HTTP API server.

Q: **Can two models satisfy the same interface**?
A: Yes. For example, if we wanted both `HelmReleases` and `Kustomizations` to implement an `Automation` interface, like so:
```golang
type Automation interface {
  ToAutomationYaml()
}
```
Then a function can specify the `Automation` interface as an argument and receive either `model` object.

Q: **Should model functions be on pointers (func (f \*foo) vs (f foo)) so that they can manipulate the model?**
A: Typically, we would want the reciever to be passed by value. If we are manipulating data of a Model instance, that may be an indication that we should use a service or return another type instead.

Q: **Can service methods accept other services as arguments?**
A: You betcha

Q: **How do I instantiate a Model? Should I use a constructor or fill out a `struct`**?
A: Constructors are preferable, but if a constructor requires too many arguments (more than 4), or a "params" `struct`, consider filling out a `struct` and using a `Validate` method instead.

Q: **Do Models always have a corresponding Service (and vice versa)?**:
A: No, not necessarily

## Consequences

Accepting this proposal would mean that regular contributors to this repository agree on this approach for feature development going forward. In addition, it also means that existing code be refactored to match the ideas and constraints laid out in this proposal. This refactor will be done incrementally, and therefore must be accounted for when estimating work.

Acceptance of this proposal also implies that code reviewers decide that a Pull Request deviates significantly from this proposed architecture and subsequently request changes to the Pull Request.

This document deliberately omits opinions about testing standards as those remain unchanged.  We expect automated tests to be present wherever applicable.
