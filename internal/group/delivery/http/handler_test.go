package http

import (
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
)

func TestGroupRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request groupRequest
		want    bool
	}{
		{"valid defaults", groupRequest{Name: "Bitcoin Vietnam", Slug: "bitcoin-vietnam"}.normalized(), true},
		{"reject malformed slug", groupRequest{Name: "Bitcoin Vietnam", Slug: "Bitcoin Vietnam"}.normalized(), false},
		{"reject invalid visibility", groupRequest{Name: "Bitcoin Vietnam", Slug: "bitcoin-vietnam", Visibility: "hidden"}.normalized(), false},
		{"allow private approval", groupRequest{Name: "Private Alpha", Slug: "private-alpha", Visibility: models.GroupVisibilityPrivate, JoinPolicy: models.GroupJoinApproval}.normalized(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.request.valid(); got != tt.want {
				t.Fatalf("valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemberRoleAndStatusContract(t *testing.T) {
	if models.GroupRoleOwner == models.GroupRoleMember || models.GroupRoleAdmin == models.GroupRoleModerator {
		t.Fatal("group roles must be distinct")
	}
	if models.GroupJoinOpen == models.GroupJoinApproval || models.GroupMembershipActive == models.GroupMembershipPending {
		t.Fatal("group states must be distinct")
	}
}
