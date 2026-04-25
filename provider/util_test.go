package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoHTTPJSONDecodesIntoTarget(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token"; got != want {
			t.Fatalf("authorization = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("content-type = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "ok"})
	}))
	t.Cleanup(srv.Close)

	var result struct {
		Name string `json:"name"`
	}

	err := DoHTTPJSON(context.Background(), http.MethodPost, srv.URL, []byte(`{"x":1}`), "token", &result)
	if err != nil {
		t.Fatalf("DoHTTPJSON() error = %v", err)
	}
	if result.Name != "ok" {
		t.Fatalf("decoded result = %#v, want name ok", result)
	}
}

func TestDoHTTPJSONReturnsStatusError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	var result struct{}
	err := DoHTTPJSON(context.Background(), http.MethodGet, srv.URL, nil, "", &result)
	if err == nil {
		t.Fatal("DoHTTPJSON() error = nil, want non-nil")
	}
}
