package httpserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0Hoag/cryptocheck-api/pkg/mongo/mocks"
	"github.com/stretchr/testify/mock"
)

func TestHealthzIsAvailableWithoutDependencies(t *testing.T) {
	srv := New(nil, Config{})
	recorder := httptest.NewRecorder()
	srv.gin.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestReadyzReportsMongoAvailability(t *testing.T) {
	db := mocks.NewDatabase(t)
	client := mocks.NewClient(t)
	db.On("Client").Return(client)
	client.On("Ping", mock.Anything).Return(nil)

	srv := New(nil, Config{DB: db})
	recorder := httptest.NewRecorder()
	srv.gin.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestReadyzDoesNotReportReadyWhenMongoFails(t *testing.T) {
	db := mocks.NewDatabase(t)
	client := mocks.NewClient(t)
	db.On("Client").Return(client)
	client.On("Ping", mock.Anything).Return(errors.New("unavailable"))

	srv := New(nil, Config{DB: db})
	recorder := httptest.NewRecorder()
	srv.gin.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
