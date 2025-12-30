package antigravity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"go-antigravity-api/pkg/models"
)

// AuthManager manages OAuth2 authentication
type AuthManager struct {
	config      *oauth2.Config
	token       *oauth2.Token
	credsPath   string
	ctx         context.Context
}

// NewAuthManager creates a new AuthManager
func NewAuthManager(credsPath string) *AuthManager {
	if credsPath == "" {
		homeDir, _ := os.UserHomeDir()
		credsPath = filepath.Join(homeDir, CredentialsDir, CredentialsFile)
	}

	return &AuthManager{
		config: &oauth2.Config{
			ClientID:     OAuthClientID,
			ClientSecret: OAuthClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost:8086",
			Scopes:       []string{"https://www.googleapis.com/auth/cloud-platform"},
		},
		credsPath: credsPath,
		ctx:       context.Background(),
	}
}

// Initialize loads or creates OAuth credentials
func (am *AuthManager) Initialize(forceRefresh bool) error {
	// Try to load existing credentials
	if err := am.loadCredentials(); err != nil {
		log.Printf("[Antigravity Auth] No existing credentials found: %v", err)
		return am.getNewToken()
	}

	// Check if token needs refresh
	if forceRefresh || am.isTokenExpiringSoon() {
		log.Println("[Antigravity Auth] Token expiring soon or force refresh requested. Refreshing token...")
		return am.refreshToken()
	}

	log.Println("[Antigravity Auth] Authentication configured successfully from file.")
	return nil
}

// loadCredentials loads credentials from file
func (am *AuthManager) loadCredentials() error {
	data, err := os.ReadFile(am.credsPath)
	if err != nil {
		return err
	}

	var creds models.OAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return err
	}

	am.token = &oauth2.Token{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		Expiry:       time.Unix(creds.ExpiryDate/1000, 0),
		TokenType:    creds.TokenType,
	}

	return nil
}

// saveCredentials saves credentials to file
func (am *AuthManager) saveCredentials() error {
	creds := models.OAuthCredentials{
		AccessToken:  am.token.AccessToken,
		RefreshToken: am.token.RefreshToken,
		ExpiryDate:   am.token.Expiry.Unix() * 1000,
		TokenType:    am.token.TokenType,
		Scope:        "https://www.googleapis.com/auth/cloud-platform",
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(am.credsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(am.credsPath, data, 0600)
}

// isTokenExpiringSoon checks if token is expiring within RefreshSkew seconds
func (am *AuthManager) isTokenExpiringSoon() bool {
	if am.token == nil || am.token.Expiry.IsZero() {
		return false
	}

	refreshTime := time.Now().Add(time.Duration(RefreshSkew) * time.Second)
	return am.token.Expiry.Before(refreshTime)
}

// refreshToken refreshes the access token
func (am *AuthManager) refreshToken() error {
	if am.token == nil || am.token.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	tokenSource := am.config.TokenSource(am.ctx, am.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	am.token = newToken
	if err := am.saveCredentials(); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	log.Printf("[Antigravity Auth] Token refreshed and saved to %s successfully.", am.credsPath)
	return nil
}

// getNewToken starts OAuth flow to get a new token
func (am *AuthManager) getNewToken() error {
	authURL := am.config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	
	fmt.Println("\n[Antigravity Auth] Please visit the following URL to authorize:")
	fmt.Println(authURL)
	fmt.Println("\n[Antigravity Auth] After authorization, paste the authorization code here:")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return fmt.Errorf("failed to read authorization code: %w", err)
	}

	token, err := am.config.Exchange(am.ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	am.token = token
	if err := am.saveCredentials(); err != nil {
		return fmt.Errorf("failed to save new token: %w", err)
	}

	log.Printf("[Antigravity Auth] New token obtained and saved to %s", am.credsPath)
	return nil
}

// GetAccessToken returns the current access token
func (am *AuthManager) GetAccessToken() (string, error) {
	if am.token == nil {
		return "", fmt.Errorf("no token available")
	}

	// Refresh if expiring soon
	if am.isTokenExpiringSoon() {
		if err := am.refreshToken(); err != nil {
			return "", err
		}
	}

	return am.token.AccessToken, nil
}

// GetTokenExpiry returns the token expiry time
func (am *AuthManager) GetTokenExpiry() time.Time {
	if am.token == nil {
		return time.Time{}
	}
	return am.token.Expiry
}
