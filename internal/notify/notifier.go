package notify

import "github.com/afeldman/fluxbrain/pkg/types"

type Notifier interface {
	Notify(ctx types.ErrorContext, result types.AnalysisResult) error
}
