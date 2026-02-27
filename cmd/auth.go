package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/auth"
	"golang.org/x/oauth2"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google Ads authentication",
}

// ---- auth login ----

var (
	authCredentialsFile string
	authDeveloperToken  string
	authManagerAccount  string
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google Ads via OAuth2",
	Long: `Start the OAuth2 login flow for Google Ads API access.

You need:
  1. A Google Cloud project with OAuth2 credentials (client_id + client_secret).
     Create one at https://console.cloud.google.com/apis/credentials
     Set redirect URI to: http://localhost:8080
  2. A Google Ads developer token from:
     https://ads.google.com/aw/apicenter
  3. Your Manager Account (MCC) customer ID.

Run with a credentials file:
  gads-cli auth login --credentials-file=~/Downloads/client_secret.json

Or provide values interactively when prompted.`,
	RunE: runAuthLogin,
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	// Load existing creds as baseline
	creds, err := auth.Load()
	if err != nil {
		creds = &auth.Credentials{}
	}

	// --- Collect client_id and client_secret ---
	if authCredentialsFile != "" {
		clientID, clientSecret, err := auth.ParseCredentialsFile(authCredentialsFile)
		if err != nil {
			return fmt.Errorf("reading credentials file: %w", err)
		}
		creds.ClientID = clientID
		creds.ClientSecret = clientSecret
		fmt.Printf("Loaded credentials from %s\n", authCredentialsFile)
	}

	// Read from stdin if not set
	if creds.ClientID == "" {
		creds.ClientID = promptRequired("Client ID: ")
	}
	if creds.ClientSecret == "" {
		creds.ClientSecret = promptRequired("Client Secret: ")
	}

	// --- Developer token ---
	if authDeveloperToken != "" {
		creds.DeveloperToken = authDeveloperToken
	} else if creds.DeveloperToken == "" {
		creds.DeveloperToken = promptRequired("Developer Token: ")
	}

	// --- Manager account (MCC) ---
	if authManagerAccount != "" {
		creds.ManagerCustomerID = authManagerAccount
	} else if creds.ManagerCustomerID == "" {
		creds.ManagerCustomerID = promptRequired("Manager Account (MCC) Customer ID: ")
	}

	// --- OAuth2 flow ---
	fmt.Println()
	fmt.Println("Starting OAuth2 authorization flow...")

	code, err := runOAuthFlow(creds)
	if err != nil {
		return err
	}

	// Exchange code for tokens
	oauthCfg := auth.NewOAuthConfig(creds)
	token, err := oauthCfg.Exchange(context.Background(), code, oauth2.AccessTypeOffline)
	if err != nil {
		return fmt.Errorf("exchanging auth code: %w", err)
	}

	creds.AccessToken = token.AccessToken
	creds.RefreshToken = token.RefreshToken
	creds.TokenType = token.TokenType
	creds.TokenExpiry = token.Expiry

	if err := auth.Save(creds); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	fmt.Printf("\nAuthentication successful!\n")
	fmt.Printf("Credentials saved to: %s\n", auth.Path())
	fmt.Printf("Manager account: %s\n", creds.ManagerCustomerID)
	return nil
}

func runOAuthFlow(creds *auth.Credentials) (string, error) {
	oauthCfg := auth.NewOAuthConfig(creds)

	// Start a local HTTP server before opening the browser
	mux := http.NewServeMux()
	codeCh := make(chan string, 1)

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return "", fmt.Errorf("failed to start local server on :8080 (is something else using it?): %w", err)
	}

	srv := &http.Server{Handler: mux}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprint(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
			codeCh <- code
		} else {
			errMsg := r.URL.Query().Get("error")
			fmt.Fprintf(w, "<html><body><h2>Authorization failed</h2><p>%s</p></body></html>", errMsg)
			codeCh <- ""
		}
	})

	go srv.Serve(ln) //nolint
	defer srv.Close()

	authURL := oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("\nOpening browser to authorize access...\n")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)
	openBrowser(authURL)

	fmt.Println("Waiting for authorization (5 minute timeout)...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	select {
	case code := <-codeCh:
		if code == "" {
			return "", fmt.Errorf("authorization denied or failed")
		}
		return code, nil
	case <-ctx.Done():
		return "", fmt.Errorf("authorization timed out after 5 minutes")
	}
}

// ---- auth token ----

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Show the current access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Load()
		if err != nil {
			return fmt.Errorf("loading credentials: %w", err)
		}
		if creds.RefreshToken == "" {
			return fmt.Errorf("not authenticated — run: gads-cli auth login")
		}
		fmt.Printf("Access Token:   %s\n", maskString(creds.AccessToken))
		fmt.Printf("Refresh Token:  %s\n", maskString(creds.RefreshToken))
		fmt.Printf("Token Type:     %s\n", creds.TokenType)
		if !creds.TokenExpiry.IsZero() {
			fmt.Printf("Token Expiry:   %s\n", creds.TokenExpiry.Format("2006-01-02 15:04:05 UTC"))
			if time.Now().After(creds.TokenExpiry) {
				fmt.Println("Status:         EXPIRED (will refresh on next use)")
			} else {
				fmt.Println("Status:         valid")
			}
		}
		return nil
	},
}

// ---- auth check ----

var authCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate the current credentials by making a test API call",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initAPIClient(); err != nil {
			return err
		}
		fmt.Println("Checking credentials...")
		accounts, err := apiClient.ListAccessibleCustomers()
		if err != nil {
			return fmt.Errorf("credentials check failed: %w", err)
		}
		fmt.Printf("Credentials valid. Found %d accessible account(s).\n", len(accounts))
		return nil
	},
}

// ---- auth status ----

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.Load()
		if err != nil {
			return fmt.Errorf("loading credentials: %w", err)
		}
		fmt.Printf("Config file: %s\n\n", auth.Path())
		if creds.RefreshToken == "" {
			fmt.Println("Status: not authenticated")
			fmt.Println("\nRun: gads-cli auth login")
			return nil
		}
		fmt.Printf("Status:           authenticated\n")
		fmt.Printf("Client ID:        %s\n", maskString(creds.ClientID))
		fmt.Printf("Developer Token:  %s\n", maskString(creds.DeveloperToken))
		fmt.Printf("Manager Account:  %s\n", creds.ManagerCustomerID)
		if !creds.TokenExpiry.IsZero() {
			fmt.Printf("Token Expiry:     %s\n", creds.TokenExpiry.Format("2006-01-02 15:04:05 UTC"))
		}
		return nil
	},
}

// ---- auth logout ----

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Clear(); err != nil {
			return fmt.Errorf("removing credentials: %w", err)
		}
		fmt.Println("Credentials removed.")
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringVar(&authCredentialsFile, "credentials-file", "", "Path to Google Cloud credentials JSON file")
	authLoginCmd.Flags().StringVar(&authDeveloperToken, "developer-token", "", "Google Ads developer token")
	authLoginCmd.Flags().StringVar(&authManagerAccount, "manager-account", "", "Manager Account (MCC) customer ID")

	authCmd.AddCommand(authLoginCmd, authTokenCmd, authCheckCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

// promptRequired reads a required value from stdin. Strips whitespace.
func promptRequired(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(msg)
		val, _ := reader.ReadString('\n')
		val = strings.TrimSpace(val)
		if val != "" {
			return val
		}
		fmt.Println("  (value required)")
	}
}
