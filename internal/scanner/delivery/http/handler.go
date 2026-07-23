package http

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/scanner"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	"github.com/0Hoag/cryptocheck-api/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const scanHistoryCollection = "scanner_history"
const scanEngineVersion = "scanner-v1"

// @Summary Scanner token
// @Schemes
// @Description Scanner token
// @Param Access-Control-Allow-Origin header string false "Access-Control-Allow-Origin" default(*)
// @Param User-Agent header string false "User-Agent" default(Swagger-Codegen/1.0.0/go)
// @Param Authorization header string true "Bearer JWT token" default(Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vcC50YW5jYS52bi9hcGkvdjQvYXV0aC9zaWduaW4tdjIiLCJpYXQiOjE3MTY1ODUyNDAsIm5iZiI6MTcxNjU4NTI0MCwianRpIjoidFBJMldUa0JldThYdnJMZiIsInN1YiI6Ik5pdEpwZUp1dkF4M1pjYXdKIiwicHJ2IjoiMWM1NTIwZjcwYmFhNjU1ZGRjNTc0NmE2NzY0ZjM3MmExYjY1NWFhNiIsInNob3BfaWQiOiI1YzIwYTE5YzBiMDg4ODBmNTk0ZmM0NjgiLCJzaG9wX3VzZXJuYW1lIjoicmF2ZSIsInNob3BfcHJlZml4IjoidCIsInR5cGUiOiJhcGkifQ.DnxirM5IXQY3B9Vcc6Qco7c9f0ABGjoeLu_1LfHiRjE)"
// @Param lang header string false "Language" default(en)
// @Param id query string false "token"
// @Tags Scanner
// @Accept json
// @Produce json
// @Success 200 {object} detailResp
// @Failure 400 {object} response.Resp "Bad Request"
// @Router /news-feed/scanner [GET]
func (h handler) ScanToken(c *gin.Context) {
	ctx := c.Request.Context()

	req, err := h.processRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.ScanToken: %v", err)
		response.Error(c, err)
		return
	}

	token, err := h.uc.ScanToken(ctx, req.ToScanTokenInput())
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.ScanToken: %v", err)
		mapErr := h.mapError(err)
		response.Error(c, mapErr)
		return
	}
	h.recordHistory(ctx, req.Token, token)

	response.OK(c, h.ToScanTokenOutput(token))
}

// History returns only the current user's successful scan records.
func (h handler) History(c *gin.Context) {
	ctx := c.Request.Context()
	scope, ok := jwt.GetScopeFromContext(ctx)
	if !ok || scope.UserID == "" {
		response.Unauthorized(c)
		return
	}
	ownerID, err := primitive.ObjectIDFromHex(scope.UserID)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	limit := int64(20)
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed < 1 || parsed > 100 {
			c.JSON(400, gin.H{"error_code": 400, "message": "limit must be between 1 and 100"})
			return
		}
		limit = parsed
	}

	cursor, err := h.db.Collection(scanHistoryCollection).Find(ctx, bson.M{"owner_id": ownerID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.History: %v", err)
		response.Error(c, err)
		return
	}
	defer cursor.Close(ctx)

	items := make([]models.ScanHistory, 0)
	if err := cursor.All(ctx, &items); err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.History.decode: %v", err)
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h handler) recordHistory(ctx context.Context, input string, token scanner.ScanTokenOutput) {
	scope, ok := jwt.GetScopeFromContext(ctx)
	if !ok || scope.UserID == "" {
		return
	}
	ownerID, err := primitive.ObjectIDFromHex(scope.UserID)
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.recordHistory.invalidOwner: %v", err)
		return
	}

	history := models.ScanHistory{
		ID:             primitive.NewObjectID(),
		OwnerID:        ownerID,
		Input:          strings.TrimSpace(input),
		Network:        token.Network,
		AnalysisType:   token.AnalysisType,
		TrustScore:     token.TrustScore,
		ScoreAvailable: token.ScoreAvailable,
		EngineVersion:  scanEngineVersion,
		CreatedAt:      time.Now().UTC(),
	}
	if _, err := h.db.Collection(scanHistoryCollection).InsertOne(ctx, history); err != nil {
		// A completed scan remains useful even if the audit record cannot be stored.
		h.l.Errorf(ctx, "scanner.delivery.http.recordHistory: %v", err)
	}
}

// FindCandidates returns the strongest DexScreener matches before the client
// chooses a symbol that can exist on more than one chain.
func (h handler) FindCandidates(c *gin.Context) {
	ctx := c.Request.Context()
	req, err := h.processRequest(c)
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.FindCandidates: %v", err)
		response.Error(c, err)
		return
	}

	candidates, err := h.uc.FindCandidates(ctx, scanner.FindCandidatesInput{Query: req.Token})
	if err != nil {
		h.l.Errorf(ctx, "scanner.delivery.http.FindCandidates: %v", err)
		response.Error(c, h.mapError(err))
		return
	}
	response.OK(c, candidates)
}
