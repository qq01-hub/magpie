package provider

import (
	"context"
	"time"
)

const checkinBudget = 10 * time.Minute

// withCheckinBudget gives one plugin account its own budget while retaining
// the caller's values, cancellation and any caller deadline
func withCheckinBudget(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, checkinBudget)
}
