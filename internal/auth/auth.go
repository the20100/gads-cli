package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	OAuthScope  = "https://www.googleapis.com/auth/adwords"
	RedirectURL = "http://localhost:8080"
)

// Credentials holds all authentication data for the Google Ads API.
type Credentials struct {
	ClientID          string    `json:"client_id"`
	ClientSecret      string    `json:"client_secret"`
	DeveloperToken    string    `json:"developer_token"`
	ManagerCustomerID string    `json:"manager_customer_id"`
	RefreshToken      string    `json:"refresh_token"`
	AccessToken       string    `json:"access_token"`
	TokenType         string    `json:"token_type"`
	TokenExpiry       time.Time `json:"token_expiry,omitempty"`
}

// GoogleCredentialsFile represents the JSON downloaded from Google Cloud Console.
type GoogleCredentialsFile struct {
	Installed *googleCredentialsEntry `json:"installed"`
	Web       *googleCredentialsEntry `json:"web"`
}

type googleCredentialsEntry struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func credentialsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gads", "credentials.json"), nil
}

// Load reads the credentials file. Returns empty Credentials (not error) if file doesn't exist.
func Load() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Credentials{}, nil
		}
		return nil, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// Save writes the credentials file with 0600 permissions.
func Save(creds *Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Clear removes the credentials file.
func Clear() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Path returns the credentials file path for display.
func Path() string {
	p, _ := credentialsPath()
	return p
}

// NewOAuthConfig creates an oauth2.Config for the Google Ads API.
func NewOAuthConfig(creds *Credentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{OAuthScope},
		RedirectURL:  RedirectURL,
	}
}

// ParseCredentialsFile parses a Google Cloud Console credentials JSON file.
func ParseCredentialsFile(path string) (clientID, clientSecret string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var f GoogleCredentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return "", "", err
	}
	entry := f.Installed
	if entry == nil {
		entry = f.Web
	}
	if entry == nil {
		return "", "", errors.New("credentials file must have 'installed' or 'web' key")
	}
	return entry.ClientID, entry.ClientSecret, nil
}
