package osys

import (
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Osys
type Osys interface {
	UserHomeDir() (string, error)
	Exit(code int)
	Stdin() *os.File
	Stdout() *os.File
	Stderr() *os.File
}

type OsysClient struct{}

func New() *OsysClient {
	return &OsysClient{}
}

var _ Osys = &OsysClient{}

func (o *OsysClient) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (o *OsysClient) Exit(code int) {
	os.Exit(code)
}

func (o *OsysClient) Stdin() *os.File {
	return os.Stdin
}

func (o *OsysClient) Stdout() *os.File {
	return os.Stdout
}

func (o *OsysClient) Stderr() *os.File {
	return os.Stderr
}
