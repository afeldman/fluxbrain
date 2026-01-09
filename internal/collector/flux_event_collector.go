package collector

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// K8sEvent is a lightweight representation of a Kubernetes Event.
type K8sEvent struct {
	InvolvedKind string
	Name         string
	Namespace    string
	Reason       string
	Message      string
	Type         string
	Timestamp    time.Time
	Source       string
}

// EventLister lists Kubernetes events in a namespace.
type EventLister interface {
	ListEvents(ctx context.Context, namespace string) ([]K8sEvent, error)
}

// FluxEventCollector emits ErrorContext objects for failed Flux resources based on Events.
type FluxEventCollector struct {
	Cluster   string
	Namespace string
	Lister    EventLister
}

// NewFluxEventCollector constructs a collector for a specific cluster/namespace.
func NewFluxEventCollector(cluster, namespace string, lister EventLister) *FluxEventCollector {
	return &FluxEventCollector{
		Cluster:   cluster,
		Namespace: namespace,
		Lister:    lister,
	}
}

// CollectFailedKustomizations finds FluxCD Kustomization failures based solely on Events.
func (c *FluxEventCollector) CollectFailedKustomizations(ctx context.Context) ([]types.ErrorContext, error) {
	if c.Lister == nil {
		return nil, errors.New("event lister is not configured")
	}

	events, err := c.Lister.ListEvents(ctx, c.Namespace)
	if err != nil {
		return nil, err
	}

	bucket := map[string]types.ErrorContext{}
	for _, ev := range events {
		if !strings.EqualFold(ev.Type, "Warning") {
			continue
		}
		if !strings.EqualFold(ev.InvolvedKind, string(types.FluxResourceKindKustomization)) {
			continue
		}
		if !isReconciliationFailure(ev.Reason, ev.Message) {
			continue
		}

		key := fmt.Sprintf("%s/%s", ev.Namespace, ev.Name)
		ctxEvents := bucket[key]
		if ctxEvents.Resource.Name == "" {
			ctxEvents = types.ErrorContext{
				Source:  "flux-event",
				Cluster: c.Cluster,
				Resource: types.ResourceRef{
					Kind:      types.FluxResourceKindKustomization,
					Name:      ev.Name,
					Namespace: ev.Namespace,
				},
				Timestamp: ev.Timestamp,
				Reason:    ev.Reason,
			}
		}
		ctxEvents.ErrorMsg = ev.Message
		ctxEvents.Events = append(ctxEvents.Events, formatEvent(ev))
		bucket[key] = ctxEvents
	}

	// deterministic order
	keys := make([]string, 0, len(bucket))
	for k := range bucket {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]types.ErrorContext, 0, len(keys))
	for _, k := range keys {
		ctxEvents := bucket[k]
		sort.Strings(ctxEvents.Events)
		out = append(out, ctxEvents)
	}
	return out, nil
}

func isReconciliationFailure(reason, message string) bool {
	reasonLower := strings.ToLower(reason)
	msgLower := strings.ToLower(message)
	if strings.Contains(reasonLower, "reconciliationfailed") || strings.Contains(reasonLower, "ready") {
		return true
	}
	failureHints := []string{"reconciliation failed", "apply failed", "health check failed", "dependency not ready"}
	for _, hint := range failureHints {
		if strings.Contains(msgLower, hint) {
			return true
		}
	}
	return false
}

func formatEvent(ev K8sEvent) string {
	return fmt.Sprintf("[%s] %s: %s", ev.Timestamp.UTC().Format(time.RFC3339), ev.Reason, ev.Message)
}

// KubernetesEventLister is a placeholder for a real implementation using client-go.
// KubernetesEventLister nutzt client-go, um Events aus einem echten Cluster zu listen.
type KubernetesEventLister struct {
	Client Interface // Interface ist ein Platzhalter für client-go Event-Client
}

// Interface ist ein Platzhalter für das client-go Event-API-Interface
type Interface interface {
	List(ctx context.Context, namespace string) ([]K8sEvent, error)
}

// ListEvents ruft Events aus dem Cluster ab (client-go Integration erforderlich)
func (k KubernetesEventLister) ListEvents(ctx context.Context, namespace string) ([]K8sEvent, error) {
	if k.Client == nil {
		return nil, errors.New("client-go Event-Client nicht konfiguriert")
	}
	return k.Client.List(ctx, namespace)
}
