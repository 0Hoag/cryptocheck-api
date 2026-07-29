package entitlement

import (
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"testing"
	"time"
)

func TestActive(t *testing.T) {
	now := time.Date(2026, time.July, 29, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		subscription models.Subscription
		want         bool
	}{
		{
			name:         "active during entitlement period",
			subscription: models.Subscription{Status: "active", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour)},
			want:         true,
		},
		{
			name:         "cancelled subscription downgrades immediately",
			subscription: models.Subscription{Status: "cancelled", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour)},
			want:         false,
		},
		{
			name:         "expired subscription downgrades at end timestamp",
			subscription: models.Subscription{Status: "active", StartsAt: now.Add(-time.Hour), EndsAt: now},
			want:         false,
		},
		{
			name:         "future subscription is not active early",
			subscription: models.Subscription{Status: "active", StartsAt: now.Add(time.Nanosecond), EndsAt: now.Add(time.Hour)},
			want:         false,
		},
		{
			name:         "provider failure status downgrades",
			subscription: models.Subscription{Status: "past_due", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour)},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Active(tt.subscription, now); got != tt.want {
				t.Fatalf("Active() = %v, want %v", got, tt.want)
			}
		})
	}
}
