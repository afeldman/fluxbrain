package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/afeldman/fluxbrain/pkg/types"
)

// Fingerprint generates a deterministic hash for an ErrorContext to enable deduplication.
// The fingerprint is based on resource identity, reason, and git revision.
func Fingerprint(ec types.ErrorContext) string {
	data, _ := json.Marshal(struct {
		Cluster   string `json:"cluster"`
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Reason    string `json:"reason"`
		GitRev    string `json:"gitRev"`
	}{
		Cluster:   ec.Cluster,
		Namespace: ec.Resource.Namespace,
		Kind:      string(ec.Resource.Kind),
		Name:      ec.Resource.Name,
		Reason:    ec.Reason,
		GitRev:    ec.Git.Revision,
	})

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:16])
}
