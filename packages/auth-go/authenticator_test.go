package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticate_StaticAPIKey(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	id, err := authn.Authenticate(context.Background(), "Bearer secret-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.AuthMethod != "api_key" || id.Subject != "bootstrap-api-key" {
		t.Fatalf("unexpected identity: %+v", id)
	}
}

func TestAuthenticate_MissingBearer(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	_, err := authn.Authenticate(context.Background(), "")
	if err != ErrMissingBearer {
		t.Fatalf("expected ErrMissingBearer, got %v", err)
	}
}

func TestAuthenticate_InvalidKey(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	_, err := authn.Authenticate(context.Background(), "Bearer wrong")
	if err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestMiddleware_SkipHealthz(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	called := false
	handler := Middleware(authn, MiddlewareOptions{Skip: SkipHealthz})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !called {
		t.Fatal("expected healthz handler to run without auth")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_RequiresAuth(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	handler := Middleware(authn, MiddlewareOptions{Skip: SkipHealthz})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMiddleware_SetsIdentityContext(t *testing.T) {
	authn := NewAuthenticator(Config{StaticAPIKeys: []string{"secret-key"}})
	var got Identity
	handler := Middleware(authn, MiddlewareOptions{Skip: SkipHealthz})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got.AuthMethod != "api_key" {
		t.Fatalf("expected api_key auth, got %+v", got)
	}
}

func TestConfig_Enabled(t *testing.T) {
	if (Config{}).Enabled() {
		t.Fatal("empty config should not enable auth")
	}
	if !(Config{IssuerURL: "http://issuer"}).Enabled() {
		t.Fatal("issuer should enable auth")
	}
	if !(Config{StaticAPIKeys: []string{"k"}}).Enabled() {
		t.Fatal("static key should enable auth")
	}
}

func TestMiddleware_DisabledPassthrough(t *testing.T) {
	authn := NewAuthenticator(Config{})
	called := false
	handler := Middleware(authn, MiddlewareOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("expected passthrough, called=%v code=%d", called, rec.Code)
	}
}
