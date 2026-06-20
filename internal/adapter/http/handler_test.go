package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateCustomerRejectsUnknownFields(t *testing.T) {
	server := echo.New()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/customers",
		strings.NewReader(`{"customerType":"INDIVIDUAL","unexpected":true}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	context := server.NewContext(request, response)
	handler := NewHandler(nil)

	err := handler.CreateCustomer(context)
	if err == nil {
		t.Fatal("expected unknown field error")
	}
	httpError, ok := err.(*echo.HTTPError)
	if !ok || httpError.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %#v", err)
	}
}
