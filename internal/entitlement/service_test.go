package entitlement

import (
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"testing"
	"time"
)

func TestActive(t *testing.T) {
	now := time.Now()
	if !Active(models.Subscription{Status: "active", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour)}, now) {
		t.Fatal("active subscription rejected")
	}
	if Active(models.Subscription{Status: "cancelled", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour)}, now) {
		t.Fatal("inactive subscription accepted")
	}
}
