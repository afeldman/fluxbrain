package context

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// Builder orchestrates collection and normalization of signals into an ErrorContext.
type Builder struct {
	Collector types.Collector
}

// NewBuilder returns a Builder using the provided collector.
func NewBuilder(c types.Collector) *Builder {
	return &Builder{Collector: c}
}

// Build constructs an ErrorContext for the selected resource.
func (b *Builder) Build(ctx context.Context, selector types.ResourceSelector) (types.ErrorContext, error) {
	signals, err := b.Collector.Collect(ctx, selector)
	if err != nil {
		return types.ErrorContext{}, err
	}
	return BuildContextFromSignals(signals), nil
}

// BuildContextFromSignals converts raw signals into an LLM-friendly ErrorContext.
func BuildContextFromSignals(signals types.CollectedSignals) types.ErrorContext {
	events := make([]string, 0, len(signals.Events))
	sort.Slice(signals.Events, func(i, j int) bool {
		return signals.Events[i].Timestamp.Before(signals.Events[j].Timestamp)
	})
	for _, ev := range signals.Events {
		events = append(events, formatEvent(ev))
	}

	logSnippets := make([]string, 0, len(signals.Logs))
	sort.Slice(signals.Logs, func(i, j int) bool {
		if signals.Logs[i].Source == signals.Logs[j].Source {
			return signals.Logs[i].FromTime.Before(signals.Logs[j].FromTime)
		}
		return signals.Logs[i].Source < signals.Logs[j].Source
	})
	for _, l := range signals.Logs {
		logSnippets = append(logSnippets, formatLog(l))
	}

	status := signals.Status
	git := types.GitContext{
		Repository: status.SourceRepository,
		Revision:   status.SourceRevision,
		Path:       status.SourcePath,
	}

	return types.ErrorContext{
		Source:  "flux-status",
		Cluster: status.Cluster,
		Resource: types.ResourceRef{
			Kind:      status.Kind,
			Name:      status.Name,
			Namespace: status.Namespace,
		},
		Git:         git,
		ErrorMsg:    status.Message,
		Reason:      status.Reason,
		Events:      events,
		LogSnippets: logSnippets,
		Timestamp:   signalTimestamp(status.ObservedAt),
	}
}

// MarshalErrorContext renders the context as deterministic JSON.
func MarshalErrorContext(ec types.ErrorContext) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ec); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func formatEvent(ev types.FluxEvent) string {
	ts := ev.Timestamp.UTC().Format(time.RFC3339)
	return ts + " | " + ev.Reason + " | " + ev.Message
}

func formatLog(sn types.LogSnippet) string {
	ts := sn.FromTime.UTC().Format(time.RFC3339)
	return sn.Source + "@" + ts + " -> " + joinLines(sn.Lines)
}

func joinLines(lines []string) string {
	switch len(lines) {
	case 0:
		return ""
	case 1:
		return lines[0]
	default:
		return lines[0] + " | " + lines[1]
	}
}

func signalTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts
}
