package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/config"
	"golang.org/x/oauth2"
)

var (
	jsonFlag   bool
	prettyFlag bool
	apiClient  *api.Client
)

var rootCmd = &cobra.Command{
	Use:   "gads-cli",
	Short: "Google Ads CLI — manage Google Ads via the REST API",
	Long: `gads-cli is a command-line tool for the Google Ads API v19.

It outputs JSON when piped (for agent/script use) and human-readable
tables when running in a terminal.

Authenticate first:
  gads-cli auth login

Then explore your accounts:
  gads-cli accounts list
  gads-cli campaigns list --account=<id>

Credential file: ~/.config/gads/credentials.json`,
	SilenceUsage: true,
}

// Execute is the entrypoint called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isSkipPreRunCommand(cmd) {
			return nil
		}
		return initAPIClient()
	}

	rootCmd.AddCommand(infoCmd)
}

// savingTokenSource wraps an oauth2.TokenSource and persists refreshed tokens to disk.
type savingTokenSource struct {
	source oauth2.TokenSource
	creds  *config.Credentials
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.source.Token()
	if err != nil {
		return nil, err
	}
	if token.AccessToken != s.creds.AccessToken {
		s.creds.AccessToken = token.AccessToken
		s.creds.TokenExpiry = token.Expiry
		if token.TokenType != "" {
			s.creds.TokenType = token.TokenType
		}
		_ = config.Save(s.creds)
	}
	return token, nil
}

func initAPIClient() error {
	creds, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}
	if creds.RefreshToken == "" {
		return fmt.Errorf("not authenticated — run: gads-cli auth login")
	}
	if creds.DeveloperToken == "" {
		return fmt.Errorf("developer token not set — run: gads-cli auth login")
	}

	oauthCfg := config.NewOAuthConfig(creds)
	token := &oauth2.Token{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		TokenType:    creds.TokenType,
		Expiry:       creds.TokenExpiry,
	}
	ts := oauthCfg.TokenSource(context.Background(), token)
	savingTS := &savingTokenSource{source: ts, creds: creds}
	httpClient := oauth2.NewClient(context.Background(), savingTS)

	apiClient = api.New(httpClient, creds.DeveloperToken, creds.ManagerCustomerID)
	return nil
}

// isSkipPreRunCommand returns true for commands that don't need API authentication.
func isSkipPreRunCommand(cmd *cobra.Command) bool {
	if isAuthCommand(cmd) {
		return true
	}
	name := cmd.Name()
	return name == "update" || name == "info" || name == "help"
}

// isAuthCommand returns true if cmd is in the auth subtree.
func isAuthCommand(cmd *cobra.Command) bool {
	for cmd != nil {
		if cmd.Name() == "auth" {
			return true
		}
		cmd = cmd.Parent()
	}
	return false
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show config path, auth status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	exe, _ := os.Executable()
	fmt.Printf("gads-cli — Google Ads CLI\n\n")
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Println("  config paths by OS:")
	fmt.Printf("    macOS:   ~/Library/Application Support/gads/credentials.json\n")
	fmt.Printf("    Linux:   ~/.config/gads/credentials.json\n")
	fmt.Printf("    Windows: %%AppData%%\\gads\\credentials.json\n")
	fmt.Printf("  config:  %s\n", config.Path())
	fmt.Println()

	creds, err := config.Load()
	if err != nil || creds.RefreshToken == "" {
		fmt.Println("  status:  not authenticated (run: gads-cli auth login)")
		return
	}
	fmt.Printf("  status:           authenticated\n")
	fmt.Printf("  manager account:  %s\n", creds.ManagerCustomerID)
	fmt.Printf("  developer token:  %s\n", maskOrEmpty(creds.DeveloperToken))
	if !creds.TokenExpiry.IsZero() {
		fmt.Printf("  token expiry:     %s\n", creds.TokenExpiry.Format("2006-01-02 15:04:05 UTC"))
	}
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}
