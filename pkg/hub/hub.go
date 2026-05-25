// Package hub provides a client for the magego.hub.io API service.
// It manages snippet upload, download, search, versioning, and authentication.
package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/VDHewei/mage-makefile/pkg/config"
)

// HubManager handles the hub API client lifecycle including authentication.
type HubManager struct {
	client     *Client
	config     *config.Config
	tokenPath  string
	tokenMutex sync.RWMutex
}

// NewHubManager creates a new Hub manager with config loading.
func NewHubManager(cfg *config.Config) *HubManager {
	hm := &HubManager{
		config:   cfg,
		tokenPath: filepath.Join(os.Getenv("HOME"), ".mage_makefile", "hub_token"),
	}
	// Load existing token if available
	hm.loadToken()
	return hm
}

// getClient returns the hub API client, creating a new one if necessary.
func (hm *HubManager) getClient() *Client {
	hm.tokenMutex.RLock()
	defer hm.tokenMutex.RUnlock()

	if hm.client != nil {
		return hm.client
	}

	serverURL := hm.config.Hub.ServerURL
	if serverURL == "" {
		serverURL = "https://magego.hub.io"
	}

	hm.client = NewClient(serverURL, hm.config.Hub.Timeout, hm.config.Hub.MaxRetries)
	return hm.client
}

// SetClient sets a custom client (useful for testing).
func (hm *HubManager) SetClient(client *Client) {
	hm.tokenMutex.Lock()
	hm.client = client
	hm.tokenMutex.Unlock()
}

// loadToken loads the saved token from the token file.
func (hm *HubManager) loadToken() {
	hm.tokenMutex.RLock()
	defer hm.tokenMutex.RUnlock()

	data, err := os.ReadFile(hm.tokenPath)
	if err != nil {
		return
	}

	var token string
	if len(data) > 0 {
		token = string(data)
	}

	if token != "" {
		hm.client.SetToken(token)
		fmt.Printf("Using cached token from %s\n", hm.tokenPath)
	}
}

// saveToken saves the token to the token file.
func (hm *HubManager) saveToken(token string) {
	hm.tokenMutex.Lock()
	defer hm.tokenMutex.Unlock()

	dir := filepath.Dir(hm.tokenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Warning: could not create token dir: %v\n", err)
		return
	}

	if err := os.WriteFile(hm.tokenPath, []byte(token), 0600); err != nil {
		fmt.Printf("Warning: could not save token: %v\n", err)
	}
}

// Login authenticates with the hub server and stores the token.
func (hm *HubManager) Login(req LoginRequest) (*LoginResponse, error) {
	client := hm.getClient()
	resp, err := client.Login(req)
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.Token != "" {
		hm.saveToken(resp.Token)
		fmt.Printf("Logged in as %s (token expires: %v)\n", resp.Username, resp.Expires)
	}

	return resp, nil
}

// Logout clears the cached token.
func (hm *HubManager) Logout() {
	hm.tokenMutex.Lock()
	defer hm.tokenMutex.Unlock()

	hm.client.SetToken("")
	if err := os.Remove(hm.tokenPath); err != nil {
		fmt.Printf("Warning: could not remove token file: %v\n", err)
	}
	fmt.Println("Logged out and token cleared")
}

// IsAuthenticated checks if a valid token is cached.
func (hm *HubManager) IsAuthenticated() bool {
	hm.tokenMutex.RLock()
	defer hm.tokenMutex.RUnlock()
	return hm.client.Token() != ""
}

// GetClient returns the current client (for testing/debugging).
func (hm *HubManager) GetClient() *Client {
	return hm.getClient()
}

// URL returns the hub server URL.
func (hm *HubManager) URL() string {
	return hm.config.Hub.ServerURL
}
