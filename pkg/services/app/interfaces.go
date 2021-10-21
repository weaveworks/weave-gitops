package app

type Automation interface {
	ToAutomationYAML() []byte
}

type automation struct {
	yaml []byte
}

func (a automation) ToAutomationYAML() []byte {
	return a.yaml
}

type Source interface {
	ToSourceYAML() []byte
}

type source struct {
	yaml []byte
}

func (s source) ToSourceYAML() []byte {
	return s.yaml
}

type AppManifest interface {
	ToAppYAML() []byte
}

type appManifest struct {
	yaml []byte
}

func (a appManifest) ToAppYAML() []byte {
	return a.yaml
}
