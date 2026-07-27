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

func TestModeratorPermissions(t *testing.T) {
	if !canModerate(models.GroupRoleOwner) || !canModerate(models.GroupRoleAdmin) || !canModerate(models.GroupRoleModerator) {
		t.Fatal("owner, admin and moderator must be able to moderate group posts")
	}
	if canModerate(models.GroupRoleMember) {
		t.Fatal("regular member must not be able to moderate group posts")
	}
}

func TestGroupPostRequestValidation(t *testing.T) {
	if !((groupPostRequest{Content: "Market update", SourceURL: "https://example.com"}).normalized().valid()) {
		t.Fatal("a valid group post was rejected")
	}
	if (groupPostRequest{Content: "", SourceURL: "https://example.com"}).normalized().valid() {
		t.Fatal("empty group post must be rejected")
	}
}
