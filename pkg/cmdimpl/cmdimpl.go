package cmdimpl

import "fmt"

type DeploymentType string
type SourceType string
type ConfigType string

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
