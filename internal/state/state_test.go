package state

import (
	"testing"
	"time"

	"github.com/afeldman/fluxbrain/pkg/types"
)

func TestFingerprint(t *testing.T) {
	ec1 := types.ErrorContext{
		Cluster: "prod",
		Resource: types.ResourceRef{
			Kind:      types.FluxResourceKindKustomization,
			Name:      "app",
			Namespace: "default",
		},
		Reason: "ReconciliationFailed",
		Git: types.GitContext{
			Revision: "main/abc123",
		},
	}

	ec2 := types.ErrorContext{
		Cluster: "prod",
		Resource: types.ResourceRef{
			Kind:      types.FluxResourceKindKustomization,
			Name:      "app",
			Namespace: "default",
		},
		Reason: "ReconciliationFailed",
		Git: types.GitContext{
			Revision: "main/abc123",
		},
	}

	ec3 := types.ErrorContext{
		Cluster: "prod",
		Resource: types.ResourceRef{
			Kind:      types.FluxResourceKindKustomization,
			Name:      "app",
			Namespace: "default",
		},
		Reason: "ReconciliationFailed",
		Git: types.GitContext{
			Revision: "main/def456",
		},
	}

	fp1 := Fingerprint(ec1)
	fp2 := Fingerprint(ec2)
	fp3 := Fingerprint(ec3)

	if fp1 != fp2 {
		t.Errorf("identical contexts should produce same fingerprint: %s != %s", fp1, fp2)
	}

	if fp1 == fp3 {
		t.Errorf("different git revisions should produce different fingerprints: %s == %s", fp1, fp3)
	}
}

func TestMemoryStoreBackoff(t *testing.T) {
	store := NewMemoryStore(100*time.Millisecond, 1*time.Second)
	fp := "test-fingerprint"

	if store.InBackoff(fp) {
		t.Error("new fingerprint should not be in backoff")
	}

	store.RegisterFailure(fp)
	if !store.InBackoff(fp) {
		t.Error("fingerprint should be in backoff after first failure")
	}

	time.Sleep(150 * time.Millisecond)
	if store.InBackoff(fp) {
		t.Error("fingerprint should exit backoff after delay")
	}

	store.RegisterFailure(fp)
	store.RegisterFailure(fp)
	if !store.InBackoff(fp) {
		t.Error("fingerprint should be in backoff after multiple failures")
	}

	store.RegisterSuccess(fp)
	if store.InBackoff(fp) {
		t.Error("fingerprint should be cleared after success")
	}
}

func TestMemoryStoreMaxBackoff(t *testing.T) {
	store := NewMemoryStore(10*time.Millisecond, 50*time.Millisecond)
	fp := "test-max-backoff"

	for i := 0; i < 20; i++ {
		store.RegisterFailure(fp)
	}

	store.mu.RLock()
	e := store.data[fp]
	store.mu.RUnlock()

	backoff := e.NextTry.Sub(time.Now())
	if backoff > 100*time.Millisecond {
		t.Errorf("backoff exceeded max: %v", backoff)
	}
}
