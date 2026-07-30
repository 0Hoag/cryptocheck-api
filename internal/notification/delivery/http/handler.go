package http

import (
	"net/http"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type handler struct{ db pkgMongo.Database }

func MapRoutes(r *gin.RouterGroup, db pkgMongo.Database, mw middleware.Middleware) {
	h := handler{db: db}
	auth := r.Group("")
	auth.Use(mw.Auth())
	auth.GET("", h.list)
	auth.POST("/:id/read", h.markRead)
	auth.POST("/read-all", h.markAllRead)
}

func (h handler) list(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	cur, err := h.db.Collection("user_notifications").Find(c.Request.Context(), bson.M{"recipient_id": uid}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	// A nil slice becomes JSON null. Clients consume this endpoint as a list,
	// so preserve the list contract even before the account has notifications.
	items := make([]models.UserNotification, 0)
	if err := cur.All(c.Request.Context(), &items); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h handler) markRead(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.StatusError(c, http.StatusBadRequest, 400, "invalid notification id")
		return
	}
	now := time.Now().UTC()
	result, err := h.db.Collection("user_notifications").UpdateOne(c.Request.Context(), bson.M{"_id": id, "recipient_id": uid, "read_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"read_at": now}})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"updated": result.ModifiedCount > 0})
}

func (h handler) markAllRead(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	now := time.Now().UTC()
	result, err := h.db.Collection("user_notifications").UpdateMany(c.Request.Context(), bson.M{"recipient_id": uid, "read_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"read_at": now}})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"updated": result.ModifiedCount})
}

func userID(c *gin.Context) (primitive.ObjectID, bool) {
	p, ok := jwt.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(jwt.NewScope(p).UserID)
	return id, err == nil
}
