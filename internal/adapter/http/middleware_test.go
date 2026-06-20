package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saaof/order-platform/customer-service/internal/adapter/auth"
)

func TestRequirePermission(t *testing.T) {
	echoServer := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	context := echoServer.NewContext(request, response)
	context.Set(authUserContextKey, auth.User{Permissions: []string{"customers.read"}})

	handler := RequirePermission("customers.read")(func(echo.Context) error {
		return nil
	})
	if err := handler(context); err != nil {
		t.Fatalf("expected permission to pass: %v", err)
	}

	handler = RequirePermission("customers.write")(func(echo.Context) error {
		return nil
	})
	if err := handler(context); err == nil {
		t.Fatal("expected permission to be rejected")
	}
}
