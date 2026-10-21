package handler

import (
	"fmt"

	"github.com/korotovsky/slack-mcp-server/pkg/provider"
	"go.uber.org/zap"
)

// requireCachesReady returns a clear, retryable error while the users or
// channels cache is still loading. stdio serves requests while the caches warm
// in the background, so handlers that resolve names, types or handles from
// the caches must not run against an empty snapshot and return misleading
// "not found" results or bare IDs. The wrapped provider error stays matchable
// with errors.Is(err, provider.ErrUsersNotReady / ErrChannelsNotReady).
func requireCachesReady(p *provider.ApiProvider, logger *zap.Logger) error {
	ready, err := p.IsReady()
	if ready {
		return nil
	}
	logger.Warn("Slack caches are still warming up", zap.Error(err))
	return fmt.Errorf("%w (cache warming up after startup, retry shortly)", err)
}
