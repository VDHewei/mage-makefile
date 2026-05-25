// Package hub provides a client for the magego.hub.io API service.
// It handles snippet push/pull/search/list/version operations
// with authentication and pagination support.
package hub

import "time"

// Snippet represents a magefile snippet stored on the hub.
type Snippet struct {
	Name        string            `json:"name"`
	Code        string            `json:"code"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Platform    string            `json:"platform"`
	Engine      string            `json:"engine"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// VersionInfo holds version information for a snippet.
type VersionInfo struct {
	Version   string    `json:"version"`
	Snippet   Snippet   `json:"snippet"`
	CreatedAt time.Time `json:"created_at"`
}

// PaginatedList holds a paginated list of snippets.
type PaginatedList struct {
	Snippets []Snippet `json:"snippets"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int       `json:"total"`
}

// SearchRequest holds parameters for a snippet search.
type SearchRequest struct {
	Query  string   `json:"query"`
	Tags   []string `json:"tags"`
	Author string   `json:"author"`
	Engine string   `json:"engine"`
	// Platform filter: "", "linux", "darwin", "windows"
	Platform string `json:"platform"`
}

// LoginRequest holds login credentials.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"api_key,omitempty"`
}

// LoginResponse holds the authentication result.
type LoginResponse struct {
	Token    string    `json:"token"`
	Username string    `json:"username"`
	Expires  time.Time `json:"expires"`
}

// LoginError is returned on failed login.
type LoginError struct {
	Error string `json:"error"`
}

// UploadResponse holds the result of a push/upload.
type UploadResponse struct {
	Name    string    `json:"name"`
	Version string    `json:"version"`
	URL     string    `json:"url"`
}

// Metadata represents magefile metadata for snippets.
type Metadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Platform    string            `json:"platform"`
	Engine      string            `json:"engine"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// APIError is a generic API error response.
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
