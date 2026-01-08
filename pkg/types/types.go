package types

import (
	"context"
	"time"
)

// FluxResourceKind represents supported FluxCD kinds.
type FluxResourceKind string

const (
	FluxResourceKindKustomization FluxResourceKind = "Kustomization"
	FluxResourceKindHelmRelease   FluxResourceKind = "HelmRelease"
	FluxResourceKindGitRepository FluxResourceKind = "GitRepository"
)

// ResourceRef identifies a Flux resource.
type ResourceRef struct {
	Kind      FluxResourceKind `json:"kind"`
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
}

// GitContext captures the Git origin of a resource.
type GitContext struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Path       string `json:"path"`
}

// ErrorContext is the LLM-optimized context handed to analyzers.
type ErrorContext struct {
	Source      string      `json:"source"`
	Cluster     string      `json:"cluster"`
	Resource    ResourceRef `json:"resource"`
	Git         GitContext  `json:"git"`
	ErrorMsg    string      `json:"errorMsg"`
	Reason      string      `json:"reason"`
	Events      []string    `json:"events"`
	LogSnippets []string    `json:"logSnippets"`
	Timestamp   time.Time   `json:"timestamp"`
}

// AnalysisResult is the normalized analysis output used downstream.
type AnalysisResult struct {
	Summary         string   `json:"summary"`
	RootCause       string   `json:"rootCause"`
	Recommendations []string `json:"recommendations"`
	RetrySafe       bool     `json:"retrySafe"`
	Confidence      float64  `json:"confidence"`
	Severity        string   `json:"severity,omitempty"`
}

// Analyzer performs root-cause analysis.
type Analyzer interface {
	Analyze(ctx context.Context, ec ErrorContext) (AnalysisResult, error)
}

// Notifier delivers analysis results to downstream systems.
type Notifier interface {
	Notify(ctx context.Context, ec ErrorContext, result AnalysisResult) error
}

// Collector gathers raw signals for a given resource.
type Collector interface {
	Collect(ctx context.Context, selector ResourceSelector) (CollectedSignals, error)
}

// ResourceSelector narrows which resources to process.
type ResourceSelector struct {
	Kind      FluxResourceKind
	Namespace string
	Name      string
	Cluster   string
}

// FluxEvent describes a Kubernetes event related to a Flux resource.
type FluxEvent struct {
	Kind      FluxResourceKind `json:"kind"`
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	Type      string           `json:"type"`
	Reason    string           `json:"reason"`
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
}

// LogSnippet contains relevant controller log lines for correlation.
type LogSnippet struct {
	Source   string    `json:"source"`
	Lines    []string  `json:"lines"`
	FromTime time.Time `json:"fromTime"`
}

// ResourceStatus captures the latest observed status of a Flux resource.
type ResourceStatus struct {
	Kind                FluxResourceKind `json:"kind"`
	Name                string           `json:"name"`
	Namespace           string           `json:"namespace"`
	Cluster             string           `json:"cluster"`
	Ready               bool             `json:"ready"`
	Status              string           `json:"status"`
	Reason              string           `json:"reason"`
	Message             string           `json:"message"`
	LastHandledRevision string           `json:"lastHandledRevision"`
	SourceRepository    string           `json:"sourceRepository"`
	SourcePath          string           `json:"sourcePath"`
	SourceRevision      string           `json:"sourceRevision"`
	ReconciliationID    string           `json:"reconciliationId"`
	ObservedAt          time.Time        `json:"observedAt"`
}

// CollectedSignals bundles the raw data before context building.
type CollectedSignals struct {
	Status ResourceStatus
	Events []FluxEvent
	Logs   []LogSnippet
}
