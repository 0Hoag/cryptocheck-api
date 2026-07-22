package mongo

import (
	"context"
	"fmt"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	roleCollections = "social_roles"
)

func (repo impleRepository) getRoleCollection() mongo.Collection {
	collName := fmt.Sprintf("%s", roleCollections)
	return repo.db.Collection(collName)
}

func (repo impleRepository) DetailRole(ctx context.Context, id string) (models.Roles, error) {
	col := repo.getRoleCollection()

	filter, err := repo.buildDetailRoleQuery(ctx, id)
	if err != nil {
		repo.l.Errorf(ctx, "users.mongo.DetailRole.buildDetailRoleQuery: %v", err)
		return models.Roles{}, err
	}

	var role models.Roles
	if err := col.FindOne(ctx, filter).Decode(&role); err != nil {
		repo.l.Errorf(ctx, "users.mongo.DetailRole.FindOne: %v", err)
		return models.Roles{}, err
	}

	return role, nil
}

func (repo impleRepository) GetRoleByName(ctx context.Context, name models.Role) (models.Roles, error) {
	var role models.Roles
	if err := repo.getRoleCollection().FindOne(ctx, bson.M{"name": string(name)}).Decode(&role); err != nil {
		repo.l.Errorf(ctx, "users.mongo.GetRoleByName.FindOne: %v", err)
		return models.Roles{}, err
	}

	return role, nil
}
