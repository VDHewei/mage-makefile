package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/auth/login":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(LoginResponse{
				Token:    "test-token",
				Username: "testuser",
				Expires:  time.Now().Add(time.Hour),
			})

		case "/api/v1/snippets":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(UploadResponse{
					Name:    "test-snippet",
					Version: "1.0.0",
					URL:     "https://hub.io/snippets/test-snippet",
				})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PaginatedList{
				Snippets: []Snippet{
					{Name: "snippet-1", Tags: []string{"build"}, Engine: "go"},
					{Name: "snippet-2", Tags: []string{"deploy"}, Engine: "js"},
				},
				Page:   1,
				PageSize: 10,
				Total:  2,
			})

		case "/api/v1/snippets/test-snippet":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Snippet{
				Name: "test-snippet", Code: "package main", Tags: []string{"test"},
			})

		case "/api/v1/snippets/search":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PaginatedList{
				Snippets: []Snippet{{Name: "search-result", Tags: []string{"build"}}},
				Page:     1, PageSize: 10, Total: 1,
			})

		case "/api/v1/snippets/test-snippet/versions":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]VersionInfo{
				{Version: "1.0.0", Snippet: Snippet{Name: "test-snippet"}},
				{Version: "0.9.0", Snippet: Snippet{Name: "test-snippet"}},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}
	}))
}

func TestLogin(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	resp, err := c.Login(LoginRequest{Username: "testuser", Password: "pass"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.Token != "test-token" {
		t.Errorf("unexpected token: %s", resp.Token)
	}
	if c.Token() != "test-token" {
		t.Errorf("client token not set: %s", c.Token())
	}
}

func TestPushRequiresAuth(t *testing.T) {
	c := NewClient("http://example.com", 5*time.Second, 0)
	_, err := c.Push(Snippet{Name: "test"})
	if err == nil {
		t.Error("expected error for unauthenticated push")
	}
}

func TestPush(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	c.Login(LoginRequest{Username: "test", Password: "pass"})
	resp, err := c.Push(Snippet{Name: "test-snippet", Code: "package main"})
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("unexpected version: %s", resp.Version)
	}
}

func TestPull(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	c.Login(LoginRequest{Username: "test", Password: "pass"})
	snippet, err := c.Pull("test-snippet", "latest")
	if err != nil {
		t.Fatalf("pull failed: %v", err)
	}
	if snippet.Name != "test-snippet" {
		t.Errorf("unexpected name: %s", snippet.Name)
	}
}

func TestSearch(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	result, err := c.Search(SearchRequest{Query: "build"})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 result, got %d", result.Total)
	}
}

func TestList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	c.Login(LoginRequest{Username: "test", Password: "pass"})
	result, err := c.List(1, 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 total, got %d", result.Total)
	}
}

func TestVersions(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, 5*time.Second, 2)
	c.Login(LoginRequest{Username: "test", Password: "pass"})
	versions, err := c.Versions("test-snippet")
	if err != nil {
		t.Fatalf("versions failed: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}
