package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/client"
	"github.com/the20100/gads-cli/internal/output"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage Google Ads accounts",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessible customer accounts under the MCC",
	Long: `List all customer accounts accessible under the configured Manager Account (MCC).

Examples:
  gads-cli accounts list
  gads-cli accounts list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		from, err := apiClient.ListAccessibleCustomers()
		if err != nil {
			return err
		}

		// For full details, query customer_client under the MCC.
		creds, _ := loadCredsForAccount()
		if creds == "" {
			// Fall back to just listing resource names
			if output.IsJSON(cmd) {
				return output.PrintJSON(from, output.IsPretty(cmd))
			}
			if len(from) == 0 {
				fmt.Println("No accessible accounts found.")
				return nil
			}
			fmt.Printf("Accessible accounts (%d):\n", len(from))
			for _, rn := range from {
				fmt.Printf("  %s\n", rn)
			}
			return nil
		}

		// Query customer_client for richer info
		query := `SELECT customer_client.id, customer_client.descriptive_name,
			customer_client.currency_code, customer_client.time_zone,
			customer_client.manager, customer_client.level, customer_client.hidden,
			customer_client.test_account
		FROM customer_client
		WHERE customer_client.level <= 1
		ORDER BY customer_client.id`

		rows, err := apiClient.Search(creds, query)
		if err != nil {
			// Fall back to listing resource names
			if output.IsJSON(cmd) {
				return output.PrintJSON(from, output.IsPretty(cmd))
			}
			fmt.Printf("Accessible accounts (%d):\n", len(from))
			for _, rn := range from {
				fmt.Printf("  %s  (run with MCC configured for full details)\n", rn)
			}
			return nil
		}

		var accounts []client.CustomerClient
		for _, raw := range rows {
			var row client.CustomerClientRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			accounts = append(accounts, row.CustomerClient)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(accounts, output.IsPretty(cmd))
		}
		if len(accounts) == 0 {
			fmt.Println("No client accounts found under MCC.")
			return nil
		}

		headers := []string{"ID", "NAME", "CURRENCY", "TIMEZONE", "MANAGER", "TEST"}
		rows2 := make([][]string, len(accounts))
		for i, a := range accounts {
			managerStr := ""
			if a.Manager {
				managerStr = "yes"
			}
			testStr := ""
			if a.TestAccount {
				testStr = "yes"
			}
			rows2[i] = []string{
				a.ID,
				output.Truncate(a.DescriptiveName, 40),
				a.CurrencyCode,
				output.Truncate(a.TimeZone, 30),
				managerStr,
				testStr,
			}
		}
		output.PrintTable(headers, rows2)
		return nil
	},
}

// loadCredsForAccount returns the MCC customer ID from loaded credentials.
func loadCredsForAccount() (string, error) {
	creds, err := loadCreds()
	if err != nil {
		return "", err
	}
	return creds.ManagerCustomerID, nil
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
	rootCmd.AddCommand(accountsCmd)
}
