package handler

import (
	"errors"
	"testing"

	"github.com/korotovsky/slack-mcp-server/pkg/provider"
	"go.uber.org/zap"
)

func TestUnitRequireCachesReadyWhileWarming(t *testing.T) {
	err := requireCachesReady(&provider.ApiProvider{}, zap.NewNop())
	if err == nil {
		t.Fatal("expected an error while the caches are not loaded")
	}
	if !errors.Is(err, provider.ErrUsersNotReady) {
		t.Fatalf("expected the wrapped users-not-ready error, got %v", err)
	}

	p := &provider.ApiProvider{}
	p.SkipCache()
	if err := requireCachesReady(p, zap.NewNop()); err != nil {
		t.Fatalf("expected nil once the caches are ready, got %v", err)
	}
}
