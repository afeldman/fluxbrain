package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// SlackNotifier posts alerts to a Slack incoming webhook.
type SlackNotifier struct {
	WebhookURL string
	ChannelID  string
}

func (s SlackNotifier) Channel() string { return "slack" }

// Notify posts a structured message to Slack.
func (s SlackNotifier) Notify(ctx context.Context, ec types.ErrorContext, result types.AnalysisResult) error {
	if s.WebhookURL == "" {
		return fmt.Errorf("slack webhook is empty")
	}

	text := fmt.Sprintf("*Fluxbrain Alert*\n*Cluster:* %s\n*Resource:* %s/%s (%s)\n*Reason:* %s\n*Summary:* %s\n*Recommendation:* %s\n*Retry safe:* %t\n*Revision:* %s",
		ec.Cluster, ec.Resource.Namespace, ec.Resource.Name, ec.Resource.Kind, result.RootCause, result.Summary, join(result.Recommendations), result.RetrySafe, ec.Git.Revision)

	payload := map[string]interface{}{
		"text": text,
	}
	if s.ChannelID != "" {
		payload["channel"] = s.ChannelID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

func join(list []string) string {
	if len(list) == 0 {
		return "n/a"
	}
	if len(list) == 1 {
		return list[0]
	}
	return list[0]
}
