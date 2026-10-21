package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// Tiers are process-wide: every Limiter() call on a tier must return the same
// limiter so concurrent callers share one budget. Reads only Burst and Limit:
// tests must not spend the process-wide tokens.
func TestUnitTiersAreSharedSingletons(t *testing.T) {
	for name, tier := range map[string]Tier{
		"Tier2": Tier2, "Tier2boost": Tier2boost, "Tier3": Tier3, "SearchMessages": SearchMessages,
	} {
		assert.Same(t, tier.Limiter(), tier.Limiter(), name)
	}
}

func TestUnitTierRates(t *testing.T) {
	assert.Equal(t, rate.Every(3*time.Second), Tier2.Limiter().Limit())
	assert.Equal(t, 3, Tier2.Limiter().Burst())
	assert.Equal(t, rate.Every(300*time.Millisecond), Tier2boost.Limiter().Limit())
	assert.Equal(t, 5, Tier2boost.Limiter().Burst())
	assert.Equal(t, rate.Every(1200*time.Millisecond), Tier3.Limiter().Limit())
	assert.Equal(t, 4, Tier3.Limiter().Burst())
}

// Search shared Tier2 (burst 3) with the conversations.list pager, so the
// fourth back-to-back search waited ~2.5s and a channel refresh could starve
// it. It keeps Slack's 20/min sustained rate but gets its own bucket and a
// burst that covers an agent's usual run of queries.
func TestUnitSearchMessagesHasItsOwnBurstyBucket(t *testing.T) {
	search := SearchMessages.Limiter()
	assert.NotSame(t, Tier2.Limiter(), search, "search must not share the conversations.list bucket")
	assert.GreaterOrEqual(t, search.Burst(), 10)
	assert.Equal(t, rate.Every(3*time.Second), search.Limit(), "sustained rate stays at Slack's Tier 2 (20/min)")
}
