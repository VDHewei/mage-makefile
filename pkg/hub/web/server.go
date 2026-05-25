// Package web provides the Fiber-based HTTP server for the Hub.
package web

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/VDHewei/mage-makefile/pkg/config"
	"github.com/VDHewei/mage-makefile/pkg/hub"
)

// Server wraps the Fiber app and config.
type Server struct {
	app *fiber.App
	cfg *config.Config
}

// New creates a new Hub web server.
func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		BodyLimit:            10 * 1024 * 1024, // 10MB max upload
		DisableStartupMessage: true,
	})

	s := &Server{app: app, cfg: cfg}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// API routes under /api/v1/
	api := s.app.Group("/api/v1")

	api.Post("/auth/login", s.handleLogin)
	api.Get("/snippets/search", s.handleSearch)
	api.Get("/snippets", s.handleList)
	api.Post("/snippets", s.handlePush)
	api.Get("/snippets/:name", s.handleGetSnippet)
	api.Get("/snippets/:name/versions", s.handleVersions)
}

// GetApp returns the Fiber app for serving.
func (s *Server) App() *fiber.App {
	return s.app
}

// handleIndex serves the home page HTML.
func (s *Server) handleIndex(c *fiber.Ctx) error {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load index.html"})
	}
	return c.Send(data)
}

// handleSnippets serves the snippets browse page.
func (s *Server) handleSnippets(c *fiber.Ctx) error {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load index.html"})
	}
	html := string(data)
	html = s.replacePageContent(html, "/snippets", "Browse Snippets")
	return c.SendString(html)
}

// handleNewSnippet serves the upload form.
func (s *Server) handleNewSnippet(c *fiber.Ctx) error {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load index.html"})
	}
	html := string(data)
	html = s.replacePageContent(html, "/snippets/new", "Upload Snippet")
	return c.SendString(html)
}

// handleSnippet serves a single snippet detail page.
func (s *Server) handleSnippet(c *fiber.Ctx) error {
	name := c.Params("name")
	client := s.getHubClient()
	_, err := client.Pull(name, "latest")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load index.html"})
	}
	html := string(data)
	html = s.replacePageContent(html, "/snippets/"+name, fmt.Sprintf("%s Detail", name))
	return c.SendString(html)
}

// handleAbout serves the about page.
func (s *Server) handleAbout(c *fiber.Ctx) error {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load index.html"})
	}
	html := string(data)
	html = s.replacePageContent(html, "/about", "About")
	return c.SendString(html)
}

// replacePageContent modifies the HTML to set the correct page.
func (s *Server) replacePageContent(html, url, title string) string {
	// This is a simplified approach - just return the html as-is
	// In production, you would use templates
	return html
}

// handleLogin handles authentication.
func (s *Server) handleLogin(c *fiber.Ctx) error {
	username := c.FormValue("username")
	apiKey := c.FormValue("api_key")

	if apiKey != "" {
		token := fmt.Sprintf("api_key:%s", apiKey)
		c.Locals("hub_authenticated", true)
		c.Locals("hub_token", token)
		return c.Status(302).Redirect("/snippets")
	}

	if username != "" {
		token := fmt.Sprintf("user:%s", username)
		c.Locals("hub_authenticated", true)
		c.Locals("hub_token", token)
		return c.Status(302).Redirect("/snippets")
	}

	return c.JSON(fiber.Map{
		"error": "Invalid credentials or username not provided",
	})
}

// handleSearch performs a search.
func (s *Server) handleSearch(c *fiber.Ctx) error {
	query := c.Query("q", "")
	client := s.getHubClient()
	result, err := client.Search(hub.SearchRequest{Query: query})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Search failed"})
	}
	return c.JSON(result)
}

// handleList lists all snippets.
func (s *Server) handleList(c *fiber.Ctx) error {
	page := 1
	pageSize := 10
	if p := c.QueryInt("page", 1); p > 0 {
		page = p
	}
	if ps := c.QueryInt("page_size", 10); ps > 0 {
		pageSize = ps
	}
	client := s.getHubClient()
	result, err := client.List(page, pageSize)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "List failed"})
	}
	return c.JSON(result)
}

// handlePush handles snippet upload.
func (s *Server) handlePush(c *fiber.Ctx) error {
	code := string(c.Body())

	// Parse metadata from code
	meta := parseMagefileMetadata(code)

	snippet := hub.Snippet{
		Name:        fmt.Sprintf("%s-latest", meta.Name),
		Code:        code,
		Description: meta.Description,
		Tags:        meta.Tags,
		Author:      meta.Author,
		Platform:    meta.Platform,
		Engine:      meta.Engine,
		Metadata:    meta.Metadata,
		CreatedAt:   meta.CreatedAt,
	}

	// Upload
	client := s.getHubClient()
	resp, err := client.Push(snippet)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"name":    resp.Name,
		"version": resp.Version,
		"url":     resp.URL,
	})
}

// handleGetSnippet gets a single snippet.
func (s *Server) handleGetSnippet(c *fiber.Ctx) error {
	name := c.Params("name")
	version := c.Query("version", "latest")
	client := s.getHubClient()
	snippet, err := client.Pull(name, version)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(snippet)
}

// handleVersions gets version history.
func (s *Server) handleVersions(c *fiber.Ctx) error {
	name := c.Params("name")
	client := s.getHubClient()
	versions, err := client.Versions(name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(versions)
}

// getHubClient returns the hub API client.
func (s *Server) getHubClient() *hub.Client {
	hubMgr := hub.NewHubManager(s.cfg)
	return hubMgr.GetClient()
}

// parseMagefileMetadataFromCode parses magefile metadata from code.
func parseMagefileMetadataFromCode(code string) hub.Metadata {
	lines := splitLines(code)
	meta := hub.Metadata{
		Name:        "unknown",
		Description: "magefile snippet",
		Tags:        []string{},
		Author:      "unknown",
		Platform:    "",
		Engine:      "",
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "// @hub") {
			parts := strings.Split(line[7:], " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if strings.HasPrefix(part, "name=") {
					meta.Name = strings.Trim(strings.TrimPrefix(part, "name="), "'\"")
				} else if strings.HasPrefix(part, "description=") {
					meta.Description = strings.Trim(strings.TrimPrefix(part, "description="), "'\"")
				} else if strings.HasPrefix(part, "author=") {
					meta.Author = strings.Trim(strings.TrimPrefix(part, "author="), "'\"")
				} else if strings.HasPrefix(part, "platform=") {
					meta.Platform = strings.Trim(strings.TrimPrefix(part, "platform="), "'\"")
				} else if strings.HasPrefix(part, "engine=") {
					meta.Engine = strings.Trim(strings.TrimPrefix(part, "engine="), "'\"")
				} else if strings.HasPrefix(part, "tags=") {
					tagsStr := strings.Trim(strings.TrimPrefix(part, "tags="), "'\"")
					meta.Tags = strings.Split(tagsStr, ",")
				}
			}
		}
	}

	return meta
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// Metadata represents magefile metadata.
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

// Map represents a navigation link.
type Map struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// nav returns navigation links.
func (s *Server) nav() []Map {
	return []Map{
		{Name: "Home", URL: "/"},
		{Name: "Snippets", URL: "/snippets"},
		{Name: "Upload", URL: "/snippets/new"},
		{Name: "About", URL: "/about"},
	}
}
