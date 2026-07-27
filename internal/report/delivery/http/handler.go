package http

import (
	"net/http"
	"strings"
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

const reports = "content_reports"
const audits = "moderation_audits"

type handler struct{ db pkgMongo.Database }

func MapRoutes(r *gin.RouterGroup, db pkgMongo.Database, mw middleware.Middleware) {
	h := handler{db: db}
	a := r.Group("")
	a.Use(mw.Auth())
	a.POST("", h.create)
	a.GET("/mine", h.mine)
	a.GET("", h.list)
	a.PATCH("/:id", h.moderate)
}

type reportReq struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
	Details    string `json:"details"`
}

func (r reportReq) valid() bool {
	return (r.TargetType == "post" || r.TargetType == "comment") && primitive.IsValidObjectID(r.TargetID) && len(strings.TrimSpace(r.Reason)) >= 3 && len(r.Details) <= 1000
}

type moderateReq struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func (r moderateReq) valid() bool {
	return r.Status == "reviewed" || r.Status == "resolved" || r.Status == "rejected"
}
func (h handler) create(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	var r reportReq
	if c.ShouldBindJSON(&r) != nil || !r.valid() {
		bad(c, "valid target_type, target_id and reason are required")
		return
	}
	targetID, _ := primitive.ObjectIDFromHex(r.TargetID)
	col := "posts"
	if r.TargetType == "comment" {
		col = "comments"
	}
	if err := h.db.Collection(col).FindOne(c.Request.Context(), bson.M{"_id": targetID, "deleted_at": bson.M{"$exists": false}}).Decode(&bson.M{}); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "reported content not found"})
		return
	}
	now := time.Now().UTC()
	item := models.ContentReport{ID: h.db.NewObjectID(), ReporterID: uid, TargetType: r.TargetType, TargetID: targetID, Reason: strings.TrimSpace(r.Reason), Details: strings.TrimSpace(r.Details), Status: "open", CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection(reports).InsertOne(c.Request.Context(), item); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}
func (h handler) mine(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	h.query(c, bson.M{"reporter_id": uid})
}
func (h handler) list(c *gin.Context) {
	if !admin(c) {
		response.Forbidden(c)
		return
	}
	h.query(c, bson.M{})
}
func (h handler) query(c *gin.Context, filter bson.M) {
	cur, err := h.db.Collection(reports).Find(c.Request.Context(), filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(100))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var items []models.ContentReport
	if err := cur.All(c.Request.Context(), &items); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}
func (h handler) moderate(c *gin.Context) {
	if !admin(c) {
		response.Forbidden(c)
		return
	}
	actor, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		bad(c, "invalid report id")
		return
	}
	var req moderateReq
	if c.ShouldBindJSON(&req) != nil || !req.valid() {
		bad(c, "invalid moderation status")
		return
	}
	now := time.Now().UTC()
	handler := actor
	update := bson.M{"status": req.Status, "handled_by": handler, "updated_at": now}
	result, err := h.db.Collection(reports).UpdateOne(c.Request.Context(), bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		response.Error(c, err)
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "report not found"})
		return
	}
	audit := models.ModerationAudit{ID: h.db.NewObjectID(), ReportID: id, ActorID: actor, Action: req.Status, Note: strings.TrimSpace(req.Note), CreatedAt: now}
	if _, err := h.db.Collection(audits).InsertOne(c.Request.Context(), audit); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, audit)
}
func userID(c *gin.Context) (primitive.ObjectID, bool) {
	p, ok := jwt.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(p.UserID)
	return id, err == nil
}
func admin(c *gin.Context) bool {
	p, ok := jwt.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return false
	}
	for _, r := range p.Roles {
		if strings.EqualFold(r, "admin") {
			return true
		}
	}
	return false
}
func bad(c *gin.Context, m string) {
	c.JSON(http.StatusBadRequest, response.Resp{ErrorCode: 400, Message: m})
}
