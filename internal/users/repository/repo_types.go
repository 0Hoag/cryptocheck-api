package repository

import (
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
)

// Post
type CreateOptions struct {
	UserName    string
	AvatarURL   string
	Phone       string
	Password    string
	Birthday    time.Time
	Roles       []string
	Permissions []string
}

type Filter struct {
	ID       string
	IDs      []string
	UserName string
	Phone    string
}

type GetOneOptions struct {
	Filter
}

type ListOptions struct {
	Filter
}

type GetOptions struct {
	Filter
	PagQuery paginator.PaginatorQuery
}

type UpdateOptions struct {
	User      models.User
	UserName  string
	AvatarURL string
}
