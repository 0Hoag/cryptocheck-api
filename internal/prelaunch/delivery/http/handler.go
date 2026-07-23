package http

import (
	"net/http"
	"net/url"
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
)

const collection = "prelaunch_projects"

type handler struct{ db pkgMongo.Database }

func MapRoutes(r *gin.RouterGroup, db pkgMongo.Database, mw middleware.Middleware) {
	h := handler{db: db}
	r.GET("", mw.OptionalAuth(), h.list)
	r.GET("/:id", mw.OptionalAuth(), h.detail)
	auth := r.Group("")
	auth.Use(mw.Auth())
	auth.POST("", h.create)
	auth.PATCH("/:id", h.update)
	auth.DELETE("/:id", h.remove)
}

type projectRequest struct {
	Name         string     `json:"name"`
	Symbol       string     `json:"symbol"`
	WebsiteURL   string     `json:"website_url"`
	SocialURLs   []string   `json:"social_urls"`
	ClaimedChain string     `json:"claimed_chain"`
	LaunchAt     *time.Time `json:"launch_at"`
	Evidence     []string   `json:"evidence"`
}

func (r projectRequest) valid() bool {
	if len(strings.TrimSpace(r.Name)) < 2 || len(strings.TrimSpace(r.WebsiteURL)) == 0 {
		return false
	}
	u, err := url.ParseRequestURI(r.WebsiteURL)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != ""
}

func (h handler) create(c *gin.Context) {
	var req projectRequest
	if c.ShouldBindJSON(&req) != nil || !req.valid() {
		c.JSON(http.StatusBadRequest, response.Resp{ErrorCode: 400, Message: "name and valid website_url are required"})
		return
	}
	sc, ok := scope(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	owner, err := primitive.ObjectIDFromHex(sc.UserID)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	now := time.Now().UTC()
	p := models.PrelaunchProject{ID: h.db.NewObjectID(), OwnerID: owner, Name: strings.TrimSpace(req.Name), Symbol: strings.ToUpper(strings.TrimSpace(req.Symbol)), WebsiteURL: req.WebsiteURL, SocialURLs: req.SocialURLs, ClaimedChain: strings.TrimSpace(req.ClaimedChain), LaunchAt: req.LaunchAt, Evidence: req.Evidence, RiskFlags: riskFlags(req), CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection(collection).InsertOne(c.Request.Context(), p); err != nil {
		response.Error(c, err)
		return
	}
	p.IsOwner = true
	response.OK(c, p)
}

func (h handler) list(c *gin.Context) {
	cur, err := h.db.Collection(collection).Find(c.Request.Context(), bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var projects []models.PrelaunchProject
	if err := cur.All(c.Request.Context(), &projects); err != nil {
		response.Error(c, err)
		return
	}
	for index := range projects {
		projects[index] = withOwnership(c, projects[index])
	}
	response.OK(c, projects)
}

func (h handler) detail(c *gin.Context) {
	p, ok := h.load(c)
	if !ok {
		return
	}
	response.OK(c, withOwnership(c, p))
}

func (h handler) update(c *gin.Context) {
	p, ok := h.owned(c)
	if !ok {
		return
	}
	var req projectRequest
	if c.ShouldBindJSON(&req) != nil || !req.valid() {
		c.JSON(http.StatusBadRequest, response.Resp{ErrorCode: 400, Message: "name and valid website_url are required"})
		return
	}
	p.Name, p.Symbol, p.WebsiteURL, p.SocialURLs, p.ClaimedChain, p.LaunchAt, p.Evidence = strings.TrimSpace(req.Name), strings.ToUpper(strings.TrimSpace(req.Symbol)), req.WebsiteURL, req.SocialURLs, strings.TrimSpace(req.ClaimedChain), req.LaunchAt, req.Evidence
	p.RiskFlags, p.UpdatedAt = riskFlags(req), time.Now().UTC()
	_, err := h.db.Collection(collection).UpdateOne(c.Request.Context(), bson.M{"_id": p.ID}, bson.M{"$set": p})
	if err != nil {
		response.Error(c, err)
		return
	}
	p.IsOwner = true
	response.OK(c, p)
}

func (h handler) remove(c *gin.Context) {
	p, ok := h.owned(c)
	if !ok {
		return
	}
	now := time.Now().UTC()
	_, err := h.db.Collection(collection).UpdateOne(c.Request.Context(), bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h handler) load(c *gin.Context) (models.PrelaunchProject, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Resp{ErrorCode: 400, Message: "invalid project id"})
		return models.PrelaunchProject{}, false
	}
	var p models.PrelaunchProject
	if err := h.db.Collection(collection).FindOne(c.Request.Context(), bson.M{"_id": id, "deleted_at": bson.M{"$exists": false}}).Decode(&p); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "project not found"})
		return models.PrelaunchProject{}, false
	}
	return p, true
}
func (h handler) owned(c *gin.Context) (models.PrelaunchProject, bool) {
	p, ok := h.load(c)
	if !ok {
		return p, false
	}
	sc, auth := scope(c)
	if !auth {
		response.Unauthorized(c)
		return p, false
	}
	if p.OwnerID.Hex() != sc.UserID {
		response.Forbidden(c)
		return p, false
	}
	return p, true
}
func scope(c *gin.Context) (models.Scope, bool) {
	payload, ok := jwt.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return models.Scope{}, false
	}
	return jwt.NewScope(payload), true
}

func withOwnership(c *gin.Context, p models.PrelaunchProject) models.PrelaunchProject {
	sc, ok := scope(c)
	p.IsOwner = ok && p.OwnerID.Hex() == sc.UserID
	return p
}
func riskFlags(r projectRequest) []string {
	flags := []string{"No deployed contract: security score unavailable"}
	if len(r.SocialURLs) == 0 {
		flags = append(flags, "No social links supplied")
	}
	if r.LaunchAt == nil {
		flags = append(flags, "No launch date supplied")
	}
	if len(r.Evidence) == 0 {
		flags = append(flags, "No verification evidence supplied")
	}
	return flags
}
