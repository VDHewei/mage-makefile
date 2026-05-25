// Package hub provides a client for the magego.hub.io API service.
package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the magego.hub.io API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	timeout    time.Duration
	maxRetries int
}

// NewClient creates a new hub API client.
func NewClient(serverURL string, timeout time.Duration, maxRetries int) *Client {
	return &Client{
		baseURL:  serverURL,
		httpClient: &http.Client{Timeout: timeout},
		timeout:    timeout,
		maxRetries: maxRetries,
	}
}

// SetToken sets the authentication token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Token returns the current token.
func (c *Client) Token() string {
	return c.token
}

// Login authenticates and returns a token.
func (c *Client) Login(req LoginRequest) (*LoginResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal login request: %w", err)
	}

	resp, err := c.do("POST", "/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		var loginErr LoginError
		if err := json.NewDecoder(resp.Body).Decode(&loginErr); err != nil {
			return nil, fmt.Errorf("login failed (401)")
		}
		return nil, fmt.Errorf("login failed: %s", loginErr.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: status %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("decode login response: %w", err)
	}
	c.token = loginResp.Token
	return &loginResp, nil
}

// Push uploads a snippet and returns the upload result.
func (c *Client) Push(snippet Snippet) (*UploadResponse, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	body, err := json.Marshal(snippet)
	if err != nil {
		return nil, fmt.Errorf("marshal snippet: %w", err)
	}

	resp, err := c.doWithAuth("POST", "/api/v1/snippets", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("push failed: status %d", resp.StatusCode)
	}

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("decode upload response: %w", err)
	}
	return &uploadResp, nil
}

// Pull downloads a snippet by name and optional version.
func (c *Client) Pull(name, version string) (*Snippet, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	params := url.Values{}
	params.Set("version", version)
	path := fmt.Sprintf("/api/v1/snippets/%s?%s", url.PathEscape(name), params.Encode())

	resp, err := c.doWithAuth("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("snippet %q not found", name)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pull failed: status %d", resp.StatusCode)
	}

	var snippet Snippet
	if err := json.NewDecoder(resp.Body).Decode(&snippet); err != nil {
		return nil, fmt.Errorf("decode snippet: %w", err)
	}
	snippet.Name = name
	return &snippet, nil
}

// Search searches for snippets by query and tags.
func (c *Client) Search(req SearchRequest) (*PaginatedList, error) {
	params := url.Values{}
	if req.Query != "" {
		params.Set("q", req.Query)
	}
	for _, tag := range req.Tags {
		params.Add("tags", tag)
	}
	if req.Author != "" {
		params.Set("author", req.Author)
	}
	if req.Engine != "" {
		params.Set("engine", req.Engine)
	}
	if req.Platform != "" {
		params.Set("platform", req.Platform)
	}

	path := fmt.Sprintf("/api/v1/snippets/search?%s", params.Encode())
	resp, err := c.doWithAuth("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("search failed: status %d", resp.StatusCode)
	}

	var result PaginatedList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search result: %w", err)
	}
	return &result, nil
}

// List lists all snippets with pagination.
func (c *Client) List(page, pageSize int) (*PaginatedList, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := fmt.Sprintf("/api/v1/snippets?%s", params.Encode())
	resp, err := c.doWithAuth("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("list failed: status %d", resp.StatusCode)
	}

	var result PaginatedList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list result: %w", err)
	}
	return &result, nil
}

// Versions gets the version history for a snippet.
func (c *Client) Versions(name string) ([]VersionInfo, error) {
	if c.token == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	path := fmt.Sprintf("/api/v1/snippets/%s/versions", url.PathEscape(name))
	resp, err := c.doWithAuth("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("versions failed: status %d", resp.StatusCode)
	}

	var versions []VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decode versions: %w", err)
	}
	return versions, nil
}

// do performs an unauthenticated HTTP request with retries.
func (c *Client) do(method, path string, body io.Reader) (*http.Response, error) {
	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("request failed: %w", err)
		}
		if resp.StatusCode < 500 {
			return resp, nil
		}
		resp.Body.Close()
		if attempt < c.maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	return nil, fmt.Errorf("request failed after %d retries", c.maxRetries+1)
}

// doWithAuth performs an authenticated request.
func (c *Client) doWithAuth(method, path string, body io.Reader) (*http.Response, error) {
	resp, err := c.do(method, path, body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Token may have expired
		resp.Body.Close()
		return nil, fmt.Errorf("unauthorized: token expired or invalid")
	}

	return resp, nil
}
