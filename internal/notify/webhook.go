package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// WebhookNotifier pushes the full context and result to an arbitrary HTTP endpoint.
type WebhookNotifier struct {
	URL string
}

func (w WebhookNotifier) Channel() string { return "webhook" }

func (w WebhookNotifier) Notify(ctx context.Context, ec types.ErrorContext, result types.AnalysisResult) error {
	if w.URL == "" {
		return fmt.Errorf("webhook url is empty")
	}

	payload := map[string]interface{}{
		"context": ec,
		"result":  result,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
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
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
