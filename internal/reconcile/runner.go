package reconcile

import (
	"context"
	"log"
	"time"
)

// Runner manages continuous reconciliation with a ticker.
type Runner struct {
	Engine   Reconciler
	Interval time.Duration
}

// NewRunner creates a new ticker-based runner.
func NewRunner(engine Reconciler, interval time.Duration) *Runner {
	return &Runner{
		Engine:   engine,
		Interval: interval,
	}
}

// Start begins the reconciliation loop. Blocks until context is canceled.
func (r *Runner) Start(ctx context.Context) error {
	log.Printf("starting reconciliation loop with interval %s", r.Interval)

	// Run once immediately
	if err := r.Engine.RunOnce(ctx); err != nil {
		log.Printf("initial reconciliation failed: %v", err)
	}

	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("reconciliation loop stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := r.Engine.RunOnce(ctx); err != nil {
				log.Printf("reconciliation cycle failed: %v", err)
			}
		}
	}
}
