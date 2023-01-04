package fluxexec

type KeyAlgorithm string

const (
	KeyAlgorithmRSA     KeyAlgorithm = "rsa"
	KeyAlgorithmECDSA   KeyAlgorithm = "ecdsa"
	KeyAlgorithmED25519 KeyAlgorithm = "ed25519"
)

type ECDSACurve string

const (
	ECDSACurveP256 ECDSACurve = "p256"
	ECDSACurveP384 ECDSACurve = "p384"
	ECDSACurveP521 ECDSACurve = "p521"
)

type Component string

const (
	ComponentSourceController       Component = "source-controller"
	ComponentKustomizeController    Component = "kustomize-controller"
	ComponentHelmController         Component = "helm-controller"
	ComponentNotificationController Component = "notification-controller"
)

type ComponentExtra string

const (
	ComponentImageReflectorController  ComponentExtra = "image-reflector-controller"
	ComponentImageAutomationController ComponentExtra = "image-automation-controller"
)
