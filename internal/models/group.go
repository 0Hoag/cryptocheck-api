package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupVisibility string

const (
	GroupVisibilityPublic  GroupVisibility = "public"
	GroupVisibilityPrivate GroupVisibility = "private"
)

type GroupJoinPolicy string

const (
	GroupJoinOpen     GroupJoinPolicy = "open"
	GroupJoinApproval GroupJoinPolicy = "approval"
	GroupJoinInvite   GroupJoinPolicy = "invite"
)

type GroupRole string

const (
	GroupRoleOwner     GroupRole = "owner"
	GroupRoleAdmin     GroupRole = "admin"
	GroupRoleModerator GroupRole = "moderator"
	GroupRoleMember    GroupRole = "member"
)

type GroupMembershipStatus string

const (
	GroupMembershipActive  GroupMembershipStatus = "active"
	GroupMembershipPending GroupMembershipStatus = "pending"
)

// Group is a focused discussion space within the CryptoCheck community.
type Group struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID     primitive.ObjectID `bson:"owner_id" json:"-"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	AvatarURL   string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	Visibility  GroupVisibility    `bson:"visibility" json:"visibility"`
	JoinPolicy  GroupJoinPolicy    `bson:"join_policy" json:"join_policy"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"-"`
	Membership  *GroupMembership   `bson:"-" json:"membership,omitempty"`
}

type GroupMembership struct {
	ID        primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID    `bson:"group_id" json:"group_id"`
	UserID    primitive.ObjectID    `bson:"user_id" json:"user_id"`
	Role      GroupRole             `bson:"role" json:"role"`
	Status    GroupMembershipStatus `bson:"status" json:"status"`
	CreatedAt time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time             `bson:"updated_at" json:"updated_at"`
}
