package assistant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nostosx/agentx/model"
	"github.com/nostosx/servex-contrib/transport/httpx"
)

func TestRegister(t *testing.T) {
	service, err := New(model.Func(func(context.Context, *model.Request) (*model.Response, error) {
		t.Fatal("registration called the model")
		return nil, nil
	}), strings.Repeat("a", 32), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	server := httpx.NewServer()
	service.Register(t.Context(), server)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sessions/demo", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated protocol: %d", response.Code)
	}
}
