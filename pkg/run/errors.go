package run

import "errors"

var (
	ErrNoPodsForService           = errors.New("no pods found for service")
	ErrNoPodsForDeployment        = errors.New("no pods found for deployment")
	ErrNoRunningPodsForService    = errors.New("no running pods found for service")
	ErrNoRunningPodsForDeployment = errors.New("no running pods found for deployment")
	ErrDashboardPodNotFound       = errors.New("dashboard pod not found")
)
