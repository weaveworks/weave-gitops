package testutils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/go-connections/nat"
)

// Helper function to wait for the cluster to become healthy.
func waitForCluster(url string, retries int, interval time.Duration) bool {
	var counter int

	for {
		counter++

		res, err := http.Get(url)
		if err == nil && res.StatusCode == http.StatusOK {
			return true
		}

		if counter >= retries {
			return false
		}

		time.Sleep(interval)
	}
}

// Helper function to lookup the port mapped for a container.
func getContainerPort(ports nat.PortMap, port nat.Port) (string, error) {
	for key, bindings := range ports {
		if key == port {
			for _, binding := range bindings {
				return binding.HostPort, nil
			}
		}
	}

	return "", fmt.Errorf("failed to find port: %s", port)
}
