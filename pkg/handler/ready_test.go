package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/korotovsky/slack-mcp-server/pkg/provider"
	"go.uber.org/zap"
)

func TestUnitRequireCachesReadyWhileWarming(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := requireCachesReady(ctx, &provider.ApiProvider{}, zap.NewNop())
	if err == nil {
		t.Fatal("expected an error while the caches are not loaded")
	}
	if !errors.Is(err, provider.ErrUsersNotReady) {
		t.Fatalf("expected the wrapped users-not-ready error, got %v", err)
	}

	p := &provider.ApiProvider{}
	p.SkipCache()
	if err := requireCachesReady(context.Background(), p, zap.NewNop()); err != nil {
		t.Fatalf("expected nil once the caches are ready, got %v", err)
	}
}

func TestUnitRequireCachesReadyWaitsForWarmup(t *testing.T) {
	p := &provider.ApiProvider{}
	go func() {
		time.Sleep(200 * time.Millisecond)
		p.SkipCache()
	}()
	start := time.Now()
	if err := requireCachesReady(context.Background(), p, zap.NewNop()); err != nil {
		t.Fatalf("expected the call to wait for warmup and succeed, got %v", err)
	}
	if waited := time.Since(start); waited > 5*time.Second {
		t.Fatalf("waited %v, expected to return shortly after the caches became ready", waited)
	}
}
