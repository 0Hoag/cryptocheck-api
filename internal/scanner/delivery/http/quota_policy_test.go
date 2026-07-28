package http

import "testing"

func TestFreeDailyScanLimit(t *testing.T) {
	if freeDailyScanLimit != 2 {
		t.Fatalf("free daily scan limit = %d, want 2", freeDailyScanLimit)
	}
}

func TestPremiumQuotaIsUnlimited(t *testing.T) {
	quota := scanQuota{Plan: "premium", Unlimited: true}
	if !quota.Unlimited || quota.Limit != 0 || quota.Used != 0 {
		t.Fatalf("premium quota must be unlimited and not expose a numeric allowance: %+v", quota)
	}
}
