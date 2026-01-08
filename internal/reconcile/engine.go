package reconcile

import (
	"context"
	"log"

	"github.com/afeldman/fluxbrain/internal/state"
	"github.com/afeldman/fluxbrain/pkg/types"
)

// Reconciler runs the reconciliation loop once or continuously.
type Reconciler interface {
	RunOnce(ctx context.Context) error
}

// ErrorCollector abstracts FluxCD error collection.
type ErrorCollector interface {
	CollectErrors(ctx context.Context) ([]types.ErrorContext, error)
}

// Engine orchestrates error collection, deduplication, backoff, analysis, and notification.
type Engine struct {
	Collectors []ErrorCollector
	Analyzer   types.Analyzer
	Notifiers  []types.Notifier
	State      state.Store
}

// NewEngine creates a new reconciliation engine.
func NewEngine(collectors []ErrorCollector, analyzer types.Analyzer, notifiers []types.Notifier, stateStore state.Store) *Engine {
	return &Engine{
		Collectors: collectors,
		Analyzer:   analyzer,
		Notifiers:  notifiers,
		State:      stateStore,
	}
}

// RunOnce executes a single reconciliation cycle:
// 1. Collect errors from all collectors
// 2. Deduplicate via fingerprinting
// 3. Check backoff state
// 4. Analyze new/eligible errors
// 5. Notify downstream systems
// 6. Update backoff state
func (e *Engine) RunOnce(ctx context.Context) error {
	for _, collector := range e.Collectors {
		errorContexts, err := collector.CollectErrors(ctx)
		if err != nil {
			log.Printf("collector error: %v", err)
			continue
		}

		for _, ec := range errorContexts {
			fp := state.Fingerprint(ec)

			if e.State.InBackoff(fp) {
				log.Printf("skipping %s/%s (in backoff)", ec.Resource.Namespace, ec.Resource.Name)
				continue
			}

			result, err := e.Analyzer.Analyze(ctx, ec)
			if err != nil {
				log.Printf("analysis failed for %s/%s: %v", ec.Resource.Namespace, ec.Resource.Name, err)
				e.State.RegisterFailure(fp)
				continue
			}

			for _, notifier := range e.Notifiers {
				if err := notifier.Notify(ctx, ec, result); err != nil {
					log.Printf("notification failed: %v", err)
				}
			}

			e.State.RegisterSuccess(fp)
		}
	}
	return nil
}
