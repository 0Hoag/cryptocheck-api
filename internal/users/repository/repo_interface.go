package repository

import (
	"context"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
)

//go:generate mockery --name=Repository
type Repository interface {
	// User
	Create(ctx context.Context, opts CreateOptions) (models.User, error)
	Detail(ctx context.Context, sc models.Scope, id string) (models.User, error)
	GetOne(ctx context.Context, f Filter) (models.User, error)
	List(ctx context.Context, sc models.Scope, opts ListOptions) ([]models.User, error)
	Get(ctx context.Context, sc models.Scope, opts GetOptions) ([]models.User, paginator.Paginator, error)
	Update(ctx context.Context, sc models.Scope, opts UpdateOptions) error
	Delete(ctx context.Context, sc models.Scope, id string) error

	// Role
	DetailRole(ctx context.Context, id string) (models.Roles, error)
	GetRoleByName(ctx context.Context, name models.Role) (models.Roles, error)
}
