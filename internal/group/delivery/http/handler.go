package http

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/entitlement"
	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	appNotification "github.com/0Hoag/cryptocheck-api/internal/notification"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	groupsCollection      = "groups"
	membershipsCollection = "group_memberships"
)

type handler struct{ db pkgMongo.Database }

// EnsureIndexes protects the domain invariants even when requests arrive concurrently.
func EnsureIndexes(ctx context.Context, db pkgMongo.Database) error {
	activeGroup := bson.M{"deleted_at": bson.M{"$exists": false}}
	if _, err := db.Collection(groupsCollection).CreateIndex(ctx, bson.D{{Key: "slug", Value: 1}}, options.Index().SetName("unique_active_group_slug").SetUnique(true).SetPartialFilterExpression(activeGroup)); err != nil {
		return err
	}
	if _, err := db.Collection(membershipsCollection).CreateIndex(ctx, bson.D{{Key: "group_id", Value: 1}, {Key: "user_id", Value: 1}}, options.Index().SetName("unique_group_member").SetUnique(true)); err != nil {
		return err
	}
	_, err := db.Collection("posts").CreateIndex(ctx, bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}}, options.Index().SetName("group_post_feed"))
	return err
}

func MapRoutes(r *gin.RouterGroup, db pkgMongo.Database, mw middleware.Middleware) {
	h := handler{db: db}
	r.GET("", mw.OptionalAuth(), h.list)
	r.GET("/:id", mw.OptionalAuth(), h.detail)
	r.GET("/:id/posts", mw.OptionalAuth(), h.posts)
	auth := r.Group("")
	auth.Use(mw.Auth())
	auth.POST("", h.create)
	auth.PATCH("/:id", h.update)
	auth.DELETE("/:id", h.remove)
	auth.POST("/:id/join", h.join)
	auth.POST("/:id/posts", h.createPost)
	auth.DELETE("/:id/posts/:postID", h.removePost)
	auth.DELETE("/:id/members/me", h.leave)
	auth.GET("/:id/members", h.members)
	auth.PATCH("/:id/members/:userID", h.updateMember)
}

type groupRequest struct {
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	Visibility  models.GroupVisibility `json:"visibility"`
	JoinPolicy  models.GroupJoinPolicy `json:"join_policy"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (r groupRequest) normalized() groupRequest {
	r.Name, r.Slug = strings.TrimSpace(r.Name), strings.ToLower(strings.TrimSpace(r.Slug))
	r.Description, r.AvatarURL = strings.TrimSpace(r.Description), strings.TrimSpace(r.AvatarURL)
	if r.Visibility == "" {
		r.Visibility = models.GroupVisibilityPublic
	}
	if r.JoinPolicy == "" {
		r.JoinPolicy = models.GroupJoinOpen
	}
	return r
}

func (r groupRequest) valid() bool {
	if len(r.Name) < 2 || len(r.Name) > 80 || !slugPattern.MatchString(r.Slug) || len(r.Slug) > 80 || len(r.Description) > 1000 {
		return false
	}
	if r.Visibility != models.GroupVisibilityPublic && r.Visibility != models.GroupVisibilityPrivate {
		return false
	}
	if r.JoinPolicy != models.GroupJoinOpen && r.JoinPolicy != models.GroupJoinApproval && r.JoinPolicy != models.GroupJoinInvite {
		return false
	}
	if r.AvatarURL == "" {
		return true
	}
	u, err := url.ParseRequestURI(r.AvatarURL)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != ""
}

func (h handler) create(c *gin.Context) {
	var req groupRequest
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c, "name, slug, visibility and join_policy are invalid")
		return
	}
	req = req.normalized()
	if !req.valid() {
		badRequest(c, "name, slug, visibility and join_policy are invalid")
		return
	}
	owner, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	if req.Visibility == models.GroupVisibilityPrivate && !entitlement.Has(c.Request.Context(), h.db, owner, "premium", time.Now().UTC()) {
		c.JSON(http.StatusForbidden, response.Resp{ErrorCode: 403, Message: "an active premium subscription is required to create a private group"})
		return
	}
	if h.slugTaken(c, req.Slug, primitive.NilObjectID) {
		badRequest(c, "group slug is already in use")
		return
	}
	now := time.Now().UTC()
	g := models.Group{ID: h.db.NewObjectID(), OwnerID: owner, Name: req.Name, Slug: req.Slug, Description: req.Description, AvatarURL: req.AvatarURL, Visibility: req.Visibility, JoinPolicy: req.JoinPolicy, CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection(groupsCollection).InsertOne(c.Request.Context(), g); err != nil {
		response.Error(c, err)
		return
	}
	m := models.GroupMembership{ID: h.db.NewObjectID(), GroupID: g.ID, UserID: owner, Role: models.GroupRoleOwner, Status: models.GroupMembershipActive, CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection(membershipsCollection).InsertOne(c.Request.Context(), m); err != nil {
		response.Error(c, err)
		return
	}
	g.Membership = &m
	response.OK(c, g)
}

func (h handler) list(c *gin.Context) {
	filter := bson.M{"deleted_at": bson.M{"$exists": false}, "visibility": models.GroupVisibilityPublic}
	cur, err := h.db.Collection(groupsCollection).Find(c.Request.Context(), filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var groups []models.Group
	if err := cur.All(c.Request.Context(), &groups); err != nil {
		response.Error(c, err)
		return
	}
	for i := range groups {
		groups[i].Membership = h.membershipFor(c, groups[i].ID)
	}
	response.OK(c, groups)
}

func (h handler) detail(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	m := h.membershipFor(c, g.ID)
	if g.Visibility == models.GroupVisibilityPrivate && (m == nil || m.Status != models.GroupMembershipActive) {
		response.Forbidden(c)
		return
	}
	g.Membership = m
	response.OK(c, g)
}

func (h handler) update(c *gin.Context) {
	g, m, ok := h.manageable(c)
	if !ok {
		return
	}
	if m.Role != models.GroupRoleOwner {
		response.Forbidden(c)
		return
	}
	var req groupRequest
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c, "name, slug, visibility and join_policy are invalid")
		return
	}
	req = req.normalized()
	if !req.valid() {
		badRequest(c, "name, slug, visibility and join_policy are invalid")
		return
	}
	if h.slugTaken(c, req.Slug, g.ID) {
		badRequest(c, "group slug is already in use")
		return
	}
	if req.Visibility == models.GroupVisibilityPrivate && !entitlement.Has(c.Request.Context(), h.db, g.OwnerID, "premium", time.Now().UTC()) {
		c.JSON(http.StatusForbidden, response.Resp{ErrorCode: 403, Message: "an active premium subscription is required to make a group private"})
		return
	}
	g.Name, g.Slug, g.Description, g.AvatarURL, g.Visibility, g.JoinPolicy, g.UpdatedAt = req.Name, req.Slug, req.Description, req.AvatarURL, req.Visibility, req.JoinPolicy, time.Now().UTC()
	if _, err := h.db.Collection(groupsCollection).UpdateOne(c.Request.Context(), bson.M{"_id": g.ID}, bson.M{"$set": g}); err != nil {
		response.Error(c, err)
		return
	}
	g.Membership = m
	response.OK(c, g)
}

func (h handler) remove(c *gin.Context) {
	g, m, ok := h.manageable(c)
	if !ok {
		return
	}
	if m.Role != models.GroupRoleOwner {
		response.Forbidden(c)
		return
	}
	now := time.Now().UTC()
	if _, err := h.db.Collection(groupsCollection).UpdateOne(c.Request.Context(), bson.M{"_id": g.ID}, bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}}); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h handler) join(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	if g.JoinPolicy == models.GroupJoinInvite {
		response.Forbidden(c)
		return
	}
	var existing models.GroupMembership
	err := h.db.Collection(membershipsCollection).FindOne(c.Request.Context(), bson.M{"group_id": g.ID, "user_id": uid}).Decode(&existing)
	if err == nil {
		response.OK(c, existing)
		return
	}
	now, status := time.Now().UTC(), models.GroupMembershipActive
	if g.JoinPolicy == models.GroupJoinApproval {
		status = models.GroupMembershipPending
	}
	m := models.GroupMembership{ID: h.db.NewObjectID(), GroupID: g.ID, UserID: uid, Role: models.GroupRoleMember, Status: status, CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection(membershipsCollection).InsertOne(c.Request.Context(), m); err != nil {
		response.Error(c, err)
		return
	}
	eventType, message := "group.member_joined", "A member joined your group"
	if status == models.GroupMembershipPending {
		eventType, message = "group.join_requested", "A member requested to join your group"
	}
	_ = appNotification.Create(c.Request.Context(), h.db, g.OwnerID, uid, g.ID, eventType, message)
	response.OK(c, m)
}

func (h handler) leave(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	var m models.GroupMembership
	if err := h.db.Collection(membershipsCollection).FindOne(c.Request.Context(), bson.M{"group_id": g.ID, "user_id": uid}).Decode(&m); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "membership not found"})
		return
	}
	if m.Role == models.GroupRoleOwner {
		badRequest(c, "group owner must transfer ownership before leaving")
		return
	}
	if _, err := h.db.Collection(membershipsCollection).DeleteOne(c.Request.Context(), bson.M{"_id": m.ID}); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"left": true})
}

func (h handler) members(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	m := h.membershipFor(c, g.ID)
	if g.Visibility == models.GroupVisibilityPrivate && (m == nil || m.Status != models.GroupMembershipActive) {
		response.Forbidden(c)
		return
	}
	cur, err := h.db.Collection(membershipsCollection).Find(c.Request.Context(), bson.M{"group_id": g.ID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var members []models.GroupMembership
	if err := cur.All(c.Request.Context(), &members); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, members)
}

type memberRequest struct {
	Role   models.GroupRole             `json:"role"`
	Status models.GroupMembershipStatus `json:"status"`
}

type groupPostRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	SourceURL string `json:"source_url"`
}

func (r groupPostRequest) normalized() groupPostRequest {
	r.Title, r.Content, r.SourceURL = strings.TrimSpace(r.Title), strings.TrimSpace(r.Content), strings.TrimSpace(r.SourceURL)
	return r
}

func (r groupPostRequest) valid() bool {
	if len(r.Content) == 0 || len(r.Content) > 10000 || len(r.Title) > 250 {
		return false
	}
	if r.SourceURL == "" {
		return true
	}
	u, err := url.ParseRequestURI(r.SourceURL)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != ""
}

func (h handler) posts(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	m := h.membershipFor(c, g.ID)
	if g.Visibility == models.GroupVisibilityPrivate && (m == nil || m.Status != models.GroupMembershipActive) {
		response.Forbidden(c)
		return
	}
	cur, err := h.db.Collection("posts").Find(c.Request.Context(), bson.M{"group_id": g.ID, "deleted_at": bson.M{"$exists": false}}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var posts []models.Post
	if err := cur.All(c.Request.Context(), &posts); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, posts)
}

func (h handler) createPost(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	m := h.membershipFor(c, g.ID)
	if m == nil || m.Status != models.GroupMembershipActive {
		response.Forbidden(c)
		return
	}
	uid, ok := userID(c)
	if !ok {
		response.Unauthorized(c)
		return
	}
	var req groupPostRequest
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c, "content is required and source_url must be valid when supplied")
		return
	}
	req = req.normalized()
	if !req.valid() {
		badRequest(c, "content is required and source_url must be valid when supplied")
		return
	}
	now := time.Now().UTC()
	groupID := g.ID
	p := models.Post{ID: h.db.NewObjectID(), GroupID: &groupID, AuthorID: uid, Title: req.Title, Content: req.Content, SourceURL: req.SourceURL, Permission: models.PrivacyTypePublic, CreatedAt: now, UpdatedAt: now}
	if _, err := h.db.Collection("posts").InsertOne(c.Request.Context(), p); err != nil {
		response.Error(c, err)
		return
	}
	_ = appNotification.Create(c.Request.Context(), h.db, g.OwnerID, uid, g.ID, "group.post_created", "A new post was published in your group")
	response.OK(c, p)
}

func (h handler) removePost(c *gin.Context) {
	g, ok := h.group(c)
	if !ok {
		return
	}
	m := h.membershipFor(c, g.ID)
	if m == nil || m.Status != models.GroupMembershipActive {
		response.Forbidden(c)
		return
	}
	postID, err := primitive.ObjectIDFromHex(c.Param("postID"))
	if err != nil {
		badRequest(c, "invalid post id")
		return
	}
	var p models.Post
	if err := h.db.Collection("posts").FindOne(c.Request.Context(), bson.M{"_id": postID, "group_id": g.ID, "deleted_at": bson.M{"$exists": false}}).Decode(&p); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "group post not found"})
		return
	}
	uid, _ := userID(c)
	if p.AuthorID != uid && !canModerate(m.Role) {
		response.Forbidden(c)
		return
	}
	now := time.Now().UTC()
	if _, err := h.db.Collection("posts").UpdateOne(c.Request.Context(), bson.M{"_id": p.ID}, bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}}); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func canModerate(role models.GroupRole) bool {
	return role == models.GroupRoleOwner || role == models.GroupRoleAdmin || role == models.GroupRoleModerator
}

func (h handler) updateMember(c *gin.Context) {
	g, actor, ok := h.manageable(c)
	if !ok {
		return
	}
	if actor.Role != models.GroupRoleOwner && actor.Role != models.GroupRoleAdmin {
		response.Forbidden(c)
		return
	}
	targetID, err := primitive.ObjectIDFromHex(c.Param("userID"))
	if err != nil {
		badRequest(c, "invalid user id")
		return
	}
	var target models.GroupMembership
	if err := h.db.Collection(membershipsCollection).FindOne(c.Request.Context(), bson.M{"group_id": g.ID, "user_id": targetID}).Decode(&target); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "membership not found"})
		return
	}
	if target.Role == models.GroupRoleOwner {
		response.Forbidden(c)
		return
	}
	var req memberRequest
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c, "role or status is required")
		return
	}
	if req.Role != "" && req.Role != models.GroupRoleAdmin && req.Role != models.GroupRoleModerator && req.Role != models.GroupRoleMember {
		badRequest(c, "invalid member role")
		return
	}
	if req.Status != "" && req.Status != models.GroupMembershipActive && req.Status != models.GroupMembershipPending {
		badRequest(c, "invalid membership status")
		return
	}
	if actor.Role == models.GroupRoleAdmin && req.Role == models.GroupRoleAdmin {
		response.Forbidden(c)
		return
	}
	changes := bson.M{"updated_at": time.Now().UTC()}
	if req.Role != "" {
		changes["role"] = req.Role
		target.Role = req.Role
	}
	if req.Status != "" {
		changes["status"] = req.Status
		target.Status = req.Status
	}
	target.UpdatedAt = changes["updated_at"].(time.Time)
	if _, err := h.db.Collection(membershipsCollection).UpdateOne(c.Request.Context(), bson.M{"_id": target.ID}, bson.M{"$set": changes}); err != nil {
		response.Error(c, err)
		return
	}
	if req.Status == models.GroupMembershipActive {
		_ = appNotification.Create(c.Request.Context(), h.db, target.UserID, actor.UserID, g.ID, "group.membership_approved", "Your request to join a group was approved")
	}
	response.OK(c, target)
}

func (h handler) group(c *gin.Context) (models.Group, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid group id")
		return models.Group{}, false
	}
	var g models.Group
	if err := h.db.Collection(groupsCollection).FindOne(c.Request.Context(), bson.M{"_id": id, "deleted_at": bson.M{"$exists": false}}).Decode(&g); err != nil {
		c.JSON(http.StatusNotFound, response.Resp{ErrorCode: 404, Message: "group not found"})
		return models.Group{}, false
	}
	return g, true
}

func (h handler) manageable(c *gin.Context) (models.Group, *models.GroupMembership, bool) {
	g, ok := h.group(c)
	if !ok {
		return g, nil, false
	}
	m := h.membershipFor(c, g.ID)
	if m == nil || m.Status != models.GroupMembershipActive {
		response.Forbidden(c)
		return g, nil, false
	}
	return g, m, true
}

func (h handler) membershipFor(c *gin.Context, groupID primitive.ObjectID) *models.GroupMembership {
	uid, ok := userID(c)
	if !ok {
		return nil
	}
	var m models.GroupMembership
	if err := h.db.Collection(membershipsCollection).FindOne(c.Request.Context(), bson.M{"group_id": groupID, "user_id": uid}).Decode(&m); err != nil {
		return nil
	}
	return &m
}

func (h handler) slugTaken(c *gin.Context, slug string, excludedID primitive.ObjectID) bool {
	filter := bson.M{"slug": slug, "deleted_at": bson.M{"$exists": false}}
	if excludedID != primitive.NilObjectID {
		filter["_id"] = bson.M{"$ne": excludedID}
	}
	return h.db.Collection(groupsCollection).FindOne(c.Request.Context(), filter).Decode(&models.Group{}) == nil
}

func userID(c *gin.Context) (primitive.ObjectID, bool) {
	p, ok := jwt.GetPayloadFromContext(c.Request.Context())
	if !ok {
		return primitive.NilObjectID, false
	}
	id, err := primitive.ObjectIDFromHex(jwt.NewScope(p).UserID)
	return id, err == nil
}
func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, response.Resp{ErrorCode: 400, Message: message})
}
