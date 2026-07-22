package repository

import (
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
)

// Post
type CreateOptions struct {
	PostID  string
	Content string
	Attach  []models.Attachment
}

type GetOneOptions struct {
	Filter
}

type Filter struct {
	ID     string
	IDs    []string
	PostID string
}

type ListOptions struct {
	Filter
}

type GetOptions struct {
	Filter
	PagQuery paginator.PaginatorQuery
}

type UpdateOptions struct {
	Comment models.Comment
	Content string
	Attach  []models.Attachment
}
