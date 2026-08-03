package httpserver

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServerUsesBoundedConnectionTimeouts(t *testing.T) {
	engine := gin.New()
	server := (HTTPServer{gin: engine, port: 8080}).newHTTPServer()

	if server.Addr != ":8080" || server.Handler != engine {
		t.Fatalf("server = %#v, want configured address and gin handler", server)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second || server.WriteTimeout != 60*time.Second || server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected timeout configuration: header=%s read=%s write=%s idle=%s", server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout, server.IdleTimeout)
	}
}

func TestNormalizeServerRunErrorTreatsGracefulCloseAsSuccess(t *testing.T) {
	if err := normalizeServerRunError(http.ErrServerClosed); err != nil {
		t.Fatalf("normalizeServerRunError(ErrServerClosed) = %v, want nil", err)
	}
	boom := errors.New("listen failure")
	if err := normalizeServerRunError(boom); !errors.Is(err, boom) {
		t.Fatalf("normalizeServerRunError(%v) = %v, want original error", boom, err)
	}
}
