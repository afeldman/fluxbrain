package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// GitHubNotifier opens issues for persistent failures.
type GitHubNotifier struct {
	Owner string
	Repo  string
	Token string
}

func (g GitHubNotifier) Channel() string { return "github" }

func (g GitHubNotifier) Notify(ctx context.Context, ec types.ErrorContext, result types.AnalysisResult) error {
	if g.Owner == "" || g.Repo == "" || g.Token == "" {
		return fmt.Errorf("github notifier is not fully configured")
	}

	title := fmt.Sprintf("Fluxbrain: %s/%s %s reconciliation failure", ec.Resource.Namespace, ec.Resource.Name, ec.Resource.Kind)
	body := fmt.Sprintf("Cluster: %s\nReason: %s\nSummary: %s\nRecommendations:\n- %s\nRetrySafe: %t\nRevision: %s",
		ec.Cluster, result.RootCause, result.Summary, join(result.Recommendations), result.RetrySafe, ec.Git.Revision)

	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", g.Owner, g.Repo)
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("github api returned %d", resp.StatusCode)
	}
	return nil
}
