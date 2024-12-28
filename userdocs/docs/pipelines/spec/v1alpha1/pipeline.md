---
title: Pipeline

---
import TierLabel from "../../../_components/TierLabel";

# Pipeline ~ENTERPRISE~

The Pipeline API defines a resource for continuous delivery pipelines.

An example of a fully defined pipeline that creates pull requests for application promotions is shown below.

```yaml
apiVersion: pipelines.weave.works/v1alpha1
kind: Pipeline
metadata:
  name: podinfo-02
  namespace: flux-system
spec:
  appRef:
    apiVersion: helm.toolkit.fluxcd.io/v2beta1
    kind: HelmRelease
    name: podinfo
  environments:
    - name: dev
      targets:
        - namespace: podinfo-02-dev
          clusterRef:
            kind: GitopsCluster
            name: dev
            namespace: flux-system
    - name: test
      targets:
        - namespace: podinfo-02-qa
          clusterRef:
            kind: GitopsCluster
            name: dev
            namespace: flux-system
        - namespace: podinfo-02-perf
          clusterRef:
            kind: GitopsCluster
            name: dev
            namespace: flux-system
    - name: prod
      targets:
        - namespace: podinfo-02-prod
          clusterRef:
            kind: GitopsCluster
            name: prod
            namespace: flux-system
  promotion:
    strategy:
      pull-request:
        type: github
        url: https://github.com/my-org/my-app-repo
		baseBranch: main
        secretRef:
          name: github-credentials
```

## Specification

The documentation for version `v1alpha1`  of a `Pipeline` resource is found next.

### Pipeline


```go
// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PipelineSpec `json:"spec,omitempty"`
	// +kubebuilder:default={"observedGeneration":-1}
	Status PipelineStatus `json:"status,omitempty"`
}

type PipelineSpec struct {
	// Environments is a list of environments to which the pipeline's application is supposed to be deployed.
	// +required
	Environments []Environment `json:"environments"`
	// AppRef denotes the name and type of the application that's governed by the pipeline.
	// +required
	AppRef LocalAppReference `json:"appRef"`
	// Promotion defines details about how promotions are carried out between the environments
	// of this pipeline.
	// +optional
	Promotion *Promotion `json:"promotion,omitempty"`
}

type Environment struct {
	// Name defines the name of this environment. This is commonly something such as "dev" or "prod".
	// +required
	Name string `json:"name"`
	// Targets is a list of targets that are part of this environment. Each environment should have
	// at least one target.
	// +required
	Targets []Target `json:"targets"`
	// Promotion defines details about how the promotion is done on this environment.
	// +optional
	Promotion *Promotion `json:"promotion,omitempty"`
}

type Target struct {
	// Namespace denotes the namespace of this target on the referenced cluster. This is where
	// the app pointed to by the environment's `appRef` is searched.
	// +required
	Namespace string `json:"namespace"`
	// ClusterRef points to the cluster that's targeted by this target. If this field is not set, then the target is assumed
	// to point to a Namespace on the cluster that the Pipeline resources resides on (i.e. a local target).
	// +optional
	ClusterRef *CrossNamespaceClusterReference `json:"clusterRef,omitempty"`
}

// Promotion define promotion configuration for the pipeline.
type Promotion struct {
	// Manual option to allow promotion between to require manual approval before proceeding.
	// +optional
	Manual bool `json:"manual,omitempty"`
	// Strategy defines which strategy the promotion should use.
	Strategy Strategy `json:"strategy"`
}

// Strategy defines all the available promotion strategies. All of the fields in here are mutually exclusive, i.e. you can only select one
// promotion strategy per Pipeline. Failure to do so will result in undefined behaviour.
type Strategy struct {
	// PullRequest defines a promotion through a Pull Request.
	// +optional
	PullRequest *PullRequestPromotion `json:"pull-request,omitempty"`
	// Notification defines a promotion where an event is emitted through Flux's notification-controller each time an app is to be promoted.
	// +optional
	Notification *NotificationPromotion `json:"notification,omitempty"`
	// SecrefRef reference the secret that contains a 'hmac-key' field with HMAC key used to authenticate webhook calls.
	// +optional
	SecretRef *meta.LocalObjectReference `json:"secretRef,omitempty"`
}
type GitProviderType string

const (
	Github          GitProviderType = "github"
	Gitlab          GitProviderType = "gitlab"
	BitBucketServer GitProviderType = "bitbucket-server"
)

type PullRequestPromotion struct {
	// Indicates the git provider type to manage pull requests.
	// +required
	// +kubebuilder:validation:Enum=github;gitlab;bitbucket-server
	Type GitProviderType `json:"type"`
	// The git repository HTTPS URL used to patch the manifests for promotion.
	// +required
	URL string `json:"url"`
	// The branch to checkout after cloning. Note: This is just the base
	// branch that will eventually receive the PR changes upon merge and does
	// not denote the branch used to create a PR from. The latter is generated
	// automatically and cannot be provided.
	// +required
	BaseBranch string `json:"baseBranch"`
	// SecretRef specifies the Secret containing authentication credentials for
	// the git repository and for the Git provider API.
	// For HTTPS repositories the Secret must contain 'username' and 'password'
	// fields.
	// For Git Provider API to manage pull requests, it must contain a 'token' field.
	// +required
	SecretRef meta.LocalObjectReference `json:"secretRef"`
}

type NotificationPromotion struct{}

```

### References

```go
// LocalAppReference is used together with a Target to find a single instance of an application on a certain cluster.
type LocalAppReference struct {
	// API version of the referent.
	// +required
	APIVersion string `json:"apiVersion"`

	// Kind of the referent.
	// +required
	Kind string `json:"kind"`

	// Name of the referent.
	// +required
	Name string `json:"name"`
}

// CrossNamespaceClusterReference contains enough information to let you locate the
// typed Kubernetes resource object at cluster level.
type CrossNamespaceClusterReference struct {
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent.
	// +required
	Kind string `json:"kind"`

	// Name of the referent.
	// +required
	Name string `json:"name"`

	// Namespace of the referent, defaults to the namespace of the Kubernetes resource object that contains the reference.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}
```

### Status

```go
type PipelineStatus struct {
	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the Pipeline.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

#### Condition Reasons
```go
// Reasons are provided as utility, and are not part of the declarative API.
const (
	// TargetClusterNotFoundReason signals a failure to locate a cluster resource on the management cluster.
	TargetClusterNotFoundReason string = "TargetClusterNotFound"
	// TargetClusterNotReadyReason signals that a cluster pointed to by a Pipeline is not ready.
	TargetClusterNotReadyReason string = "TargetClusterNotReady"
	// ReconciliationSucceededReason signals that a Pipeline has been successfully reconciled.
	ReconciliationSucceededReason string = "ReconciliationSucceeded"
)
```

