package context

import (
	"strings"
	"testing"
	"time"

	"github.com/afeldman/fluxbrain/pkg/types"
)

func TestBuildContextFromSignalsDeterministic(t *testing.T) {
	base := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)

	signals := types.CollectedSignals{
		Status: types.ResourceStatus{
			Kind:             types.FluxResourceKindKustomization,
			Name:             "demo",
			Namespace:        "apps",
			Cluster:          "c1",
			Reason:           "ReconciliationFailed",
			Message:          "apply failed",
			SourceRepository: "github.com/org/repo",
			SourceRevision:   "main/abcdef",
			SourcePath:       "clusters/prod",
			ObservedAt:       base,
		},
		Events: []types.FluxEvent{
			{Reason: "B", Message: "second", Timestamp: base.Add(2 * time.Minute)},
			{Reason: "A", Message: "first", Timestamp: base.Add(1 * time.Minute)},
		},
		Logs: []types.LogSnippet{
			{Source: "ctrl", FromTime: base.Add(3 * time.Minute), Lines: []string{"later"}},
			{Source: "ctrl", FromTime: base.Add(2 * time.Minute), Lines: []string{"earlier"}},
		},
	}

	ctx := BuildContextFromSignals(signals)
	payload, err := MarshalErrorContext(ctx)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	expected := `{
  "source": "flux-status",
  "cluster": "c1",
  "resource": {
    "kind": "Kustomization",
    "name": "demo",
    "namespace": "apps"
  },
  "git": {
    "repository": "github.com/org/repo",
    "revision": "main/abcdef",
    "path": "clusters/prod"
  },
  "errorMsg": "apply failed",
  "reason": "ReconciliationFailed",
  "events": [
    "2024-12-24T10:01:00Z | A | first",
    "2024-12-24T10:02:00Z | B | second"
  ],
  "logSnippets": [
    "ctrl@2024-12-24T10:02:00Z -> earlier",
    "ctrl@2024-12-24T10:03:00Z -> later"
  ],
  "timestamp": "2024-12-24T10:00:00Z"
}`

	got := strings.TrimSpace(string(payload))
	if got != expected {
		t.Fatalf("unexpected JSON\nexpected:\n%s\n\ngot:\n%s", expected, got)
	}
}
