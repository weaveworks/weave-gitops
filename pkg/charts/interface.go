package charts

import (
	"context"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
)

// ChartClient implementations interact with Helm repositories.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . ChartClient
type ChartClient interface {
	UpdateCache(ctx context.Context) error
	// FileFromChart gets the bytes for a named file from the chart.
	FileFromChart(ctx context.Context, c *ChartReference, filename string) ([]byte, error)
	SetRepository(repo *sourcev1beta1.HelmRepository)
}
