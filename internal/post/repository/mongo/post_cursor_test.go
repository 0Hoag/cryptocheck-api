package mongo

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/post/repository"
	storage "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	driverMongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type postListDatabase struct{ collection storage.Collection }

func (db postListDatabase) Collection(string) storage.Collection { return db.collection }
func (postListDatabase) Client() storage.Client                  { return nil }
func (postListDatabase) NewObjectID() primitive.ObjectID         { return primitive.NilObjectID }

type postListCollection struct{ cursor storage.Cursor }

func (c postListCollection) Find(context.Context, interface{}, ...*options.FindOptions) (storage.Cursor, error) {
	return c.cursor, nil
}
func (postListCollection) FindOne(context.Context, interface{}) storage.SingleResult { return nil }
func (postListCollection) InsertOne(context.Context, interface{}) (interface{}, error) {
	return nil, nil
}
func (postListCollection) InsertMany(context.Context, []interface{}) ([]interface{}, error) {
	return nil, nil
}
func (postListCollection) DeleteOne(context.Context, interface{}) (int64, error)      { return 0, nil }
func (postListCollection) DeleteMany(context.Context, interface{}) (int64, error)     { return 0, nil }
func (postListCollection) DeleteSoftOne(context.Context, interface{}) (int64, error)  { return 0, nil }
func (postListCollection) DeleteSoftMany(context.Context, interface{}) (int64, error) { return 0, nil }
func (postListCollection) CountDocuments(context.Context, interface{}, ...*options.CountOptions) (int64, error) {
	return 0, nil
}
func (postListCollection) Aggregate(context.Context, interface{}) (storage.Cursor, error) {
	return nil, nil
}
func (postListCollection) UpdateOne(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*driverMongo.UpdateResult, error) {
	return nil, nil
}
func (postListCollection) UpdateMany(context.Context, interface{}, interface{}, ...*options.UpdateOptions) (*driverMongo.UpdateResult, error) {
	return nil, nil
}
func (postListCollection) CreateIndex(context.Context, interface{}, *options.IndexOptions) (string, error) {
	return "", nil
}

type postListCursor struct{ closed bool }

func (c *postListCursor) All(_ context.Context, result interface{}) error {
	posts, ok := result.(*[]models.Post)
	if !ok {
		return &invalidListDestinationError{}
	}
	*posts = []models.Post{{Title: "Decoded post"}}
	return nil
}
func (c *postListCursor) Close(context.Context) error          { c.closed = true; return nil }
func (*postListCursor) Next(context.Context) bool              { return false }
func (*postListCursor) Decode(interface{}) error               { return nil }
func (*postListCursor) One(context.Context, interface{}) error { return nil }

type invalidListDestinationError struct{}

func (*invalidListDestinationError) Error() string { return "post list destination must be a pointer" }

func TestListDecodesPostsIntoPointerAndClosesCursor(t *testing.T) {
	ctx := context.Background()
	cursor := &postListCursor{}
	repo := impleRepository{db: postListDatabase{collection: postListCollection{cursor: cursor}}}

	posts, err := repo.List(ctx, models.Scope{}, repository.ListOptions{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(posts) != 1 || posts[0].Title != "Decoded post" {
		t.Fatalf("List() posts = %#v, want decoded post", posts)
	}
	if !cursor.closed {
		t.Fatal("List() did not close the cursor")
	}
}
