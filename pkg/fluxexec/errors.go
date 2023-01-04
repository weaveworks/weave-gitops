package fluxexec

import "fmt"

type ErrNoSuitableBinary struct {
	err error
}

func (e *ErrNoSuitableBinary) Error() string {
	return fmt.Sprintf("no suitable flux binary could be found: %s", e.err.Error())
}

func (e *ErrNoSuitableBinary) Unwrap() error {
	return e.err
}
