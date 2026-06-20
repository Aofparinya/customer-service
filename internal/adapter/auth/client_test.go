package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientCachesValidTokens(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{
			"sub":"11111111-1111-1111-1111-111111111111",
			"email":"admin@example.com",
			"roles":["ADMIN"],
			"permissions":["customers.read"],
			"sessionId":"session",
			"type":"access",
			"exp":4102444800
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 30*time.Second)
	for range 2 {
		user, err := client.Validate(context.Background(), "access-token")
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if !user.HasPermission("customers.read") {
			t.Fatal("expected customers.read permission")
		}
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected one auth request, got %d", calls)
	}
}
