package provider

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestUnitBackgroundRefreshDeadline pins that a background users refresh whose
// fetch never returns on its own is cancelled at the deadline, releasing the
// refreshing flag so a later refresh can run.
func TestUnitBackgroundRefreshDeadline(t *testing.T) {
	var requests atomic.Int32
	stop := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		select { // hang until the client gives up or the test ends
		case <-r.Context().Done():
		case <-stop:
		}
	}))
	defer ts.Close()
	defer close(stop)

	client := slack.New("xoxc-test", slack.OptionAPIURL(ts.URL+"/"))
	ap := &ApiProvider{
		client:                   &fakeUsersAPI{c: client},
		logger:                   zap.NewNop(),
		usersCachePath:           t.TempDir() + "/users_cache.json",
		backgroundRefreshTimeout: 50 * time.Millisecond,
	}

	ap.spawnBackgroundUsersRefresh()
	require.True(t, ap.refreshingUsers.Load())

	require.Eventually(t, func() bool {
		return !ap.refreshingUsers.Load()
	}, 5*time.Second, 5*time.Millisecond, "refreshingUsers was never released; the refresh has no deadline")
	assert.GreaterOrEqual(t, requests.Load(), int32(1))

	// Recovery: a later spawn must be allowed once the hung one timed out.
	before := requests.Load()
	ap.spawnBackgroundUsersRefresh()
	require.Eventually(t, func() bool {
		return requests.Load() > before
	}, 5*time.Second, 5*time.Millisecond, "a later spawn was suppressed after the first timed out")
}

func TestUnitRefreshTimeoutDefault(t *testing.T) {
	assert.Equal(t, defaultBackgroundRefreshTimeout, (&ApiProvider{}).refreshTimeout())
	assert.Equal(t, time.Second, (&ApiProvider{backgroundRefreshTimeout: time.Second}).refreshTimeout())
}
