package provider

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func emptyUsersSnapshot() *UsersCache {
	return &UsersCache{
		Users:    make(map[string]slack.User),
		UsersInv: make(map[string]string),
	}
}

func TestUnitRevalidateIfExpired(t *testing.T) {
	const ttl = time.Hour
	old := time.Now().Add(-2 * ttl).UnixNano()
	fresh := time.Now().Add(-ttl / 2).UnixNano()

	run := func(cacheTTL time.Duration, fetched int64, attempt int64) (calls int, attemptAfter int64) {
		ap := &ApiProvider{cacheTTL: cacheTTL, logger: zap.NewNop()}
		var f, a atomic.Int64
		f.Store(fetched)
		a.Store(attempt)
		ap.revalidateIfExpired(&f, &a, func() { calls++ })
		ap.revalidateIfExpired(&f, &a, func() { calls++ }) // immediate second read
		return calls, a.Load()
	}

	t.Run("expired snapshot triggers exactly one refresh across repeated reads", func(t *testing.T) {
		calls, attempt := run(ttl, old, 0)
		assert.Equal(t, 1, calls)
		assert.NotZero(t, attempt)
	})
	t.Run("fresh snapshot does nothing", func(t *testing.T) {
		calls, _ := run(ttl, fresh, 0)
		assert.Zero(t, calls)
	})
	t.Run("zero TTL disables the check", func(t *testing.T) {
		calls, _ := run(0, old, 0)
		assert.Zero(t, calls)
	})
	t.Run("unknown fetch time does nothing", func(t *testing.T) {
		calls, _ := run(ttl, 0, 0)
		assert.Zero(t, calls)
	})
	t.Run("a recent failed attempt is not retried until the retry interval passes", func(t *testing.T) {
		calls, _ := run(ttl, old, time.Now().Add(-time.Minute).UnixNano())
		assert.Zero(t, calls)
		calls, _ = run(ttl, old, time.Now().Add(-2*revalidateRetryInterval).UnixNano())
		assert.Equal(t, 1, calls)
	})
}

// A read that finds the users snapshot past its TTL serves the stale snapshot
// immediately and refreshes it in the background.
func TestUnitProvideUsersMapRevalidatesExpiredSnapshot(t *testing.T) {
	srv := &usersServer{totalPages: 1, usersPerPage: 3}
	ap, _, ts := newTestProvider(t, srv)
	defer ts.Close()

	ap.cacheTTL = time.Hour
	ap.usersSnapshot.Store(emptyUsersSnapshot())
	ap.usersReady.Store(true)
	stale := time.Now().Add(-2 * time.Hour).UnixNano()
	ap.usersFetchedAt.Store(stale)

	got := ap.ProvideUsersMap()
	assert.Empty(t, got.Users, "the read must serve the stale snapshot without waiting")

	require.Eventually(t, func() bool {
		return len(ap.ProvideUsersMap().Users) == 3 && !ap.refreshingUsers.Load()
	}, 5*time.Second, 10*time.Millisecond, "expired snapshot was never refreshed")
	assert.Greater(t, ap.usersFetchedAt.Load(), stale)
}
