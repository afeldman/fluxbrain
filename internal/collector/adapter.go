package collector

import (
	"context"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// FluxErrorCollector adapts FluxEventCollector to the ErrorCollector interface.
type FluxErrorCollector struct {
	*FluxEventCollector
}

// NewFluxErrorCollector wraps a FluxEventCollector.
func NewFluxErrorCollector(cluster, namespace string, lister EventLister) *FluxErrorCollector {
	return &FluxErrorCollector{
		FluxEventCollector: NewFluxEventCollector(cluster, namespace, lister),
	}
}

// CollectErrors implements the ErrorCollector interface.
func (c *FluxErrorCollector) CollectErrors(ctx context.Context) ([]types.ErrorContext, error) {
	return c.CollectFailedKustomizations(ctx)
}
