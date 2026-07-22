package follow

import (
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
)

// Follow
type CreateInput struct {
	FolloweeID string
}

type Filter struct {
	ID         string
	IDs        []string
	AuthorID   string
	FolloweeID string
}

type ListInput struct {
	Filter
}

type GetInput struct {
	Filter
	PagQuery paginator.PaginatorQuery
}

type GetOutput struct {
	Follows   []models.Follow
	Paginator paginator.Paginator
}
