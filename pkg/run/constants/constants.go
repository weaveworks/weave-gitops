package constants

const (
	RunDevBucketName        = "run-dev-bucket"
	RunDevKsName            = "run-dev-ks"
	RunDevHelmName          = "run-dev-helm"
	GitOpsRunNamespace      = "gitops-run"
	RunDevBucketCredentials = RunDevBucketName + "-credentials"
	RunDevKsDecryption      = RunDevKsName + "-decryption"
)

const (
	PausedAnnotation = "loft.sh/paused"
)
