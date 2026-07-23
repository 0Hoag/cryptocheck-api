package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/follow"
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/pkg/paginator"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type followUsecaseStub struct {
	get func(input follow.GetInput) (follow.GetOutput, error)
}

func (s followUsecaseStub) Create(context.Context, models.Scope, follow.CreateInput) (models.Follow, error) {
	return models.Follow{}, nil
}
func (s followUsecaseStub) Detail(context.Context, models.Scope, string) (models.Follow, error) {
	return models.Follow{}, nil
}
func (s followUsecaseStub) List(context.Context, models.Scope, follow.ListInput) ([]models.Follow, error) {
	return nil, nil
}
func (s followUsecaseStub) Get(_ context.Context, _ models.Scope, input follow.GetInput) (follow.GetOutput, error) {
	return s.get(input)
}
func (s followUsecaseStub) Delete(context.Context, models.Scope, string) error { return nil }

func TestCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := "507f1f77bcf86cd799439011"
	calls := make([]follow.GetInput, 0, 2)
	h := New(nil, followUsecaseStub{get: func(input follow.GetInput) (follow.GetOutput, error) {
		calls = append(calls, input)
		if input.FolloweeID == userID {
			return follow.GetOutput{Paginator: paginator.Paginator{Total: 7}}, nil
		}
		return follow.GetOutput{Paginator: paginator.Paginator{Total: 3}}, nil
	}})
	router := gin.New()
	router.GET("/counts/:user_id", h.Counts)

	req := httptest.NewRequest(http.MethodGet, "/counts/"+userID, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var body struct {
		Data struct {
			Followers int64 `json:"followers"`
			Following int64 `json:"following"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.Equal(t, int64(7), body.Data.Followers)
	require.Equal(t, int64(3), body.Data.Following)
	require.Len(t, calls, 2)
	require.Equal(t, userID, calls[0].FolloweeID)
	require.Equal(t, userID, calls[1].AuthorID)
	require.Equal(t, int64(1), calls[0].PagQuery.Limit)
}

func TestCountsRejectsInvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(nil, followUsecaseStub{get: func(follow.GetInput) (follow.GetOutput, error) {
		t.Fatal("use case must not be called for invalid user ID")
		return follow.GetOutput{}, nil
	}})
	router := gin.New()
	router.GET("/counts/:user_id", h.Counts)

	req := httptest.NewRequest(http.MethodGet, "/counts/not-an-object-id", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
}
