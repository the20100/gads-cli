package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/config"
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

		creds, _ := config.Load()
		mccID := ""
		if creds != nil {
			mccID = api.CleanCustomerID(creds.ManagerCustomerID)
		}

		var accounts []api.CustomerClient

		// Strategy 1: customer_client GAQL query (efficient, single call)
		if mccID != "" {
			query := `SELECT customer_client.id, customer_client.descriptive_name,
				customer_client.currency_code, customer_client.time_zone,
				customer_client.manager, customer_client.level, customer_client.hidden,
				customer_client.test_account
			FROM customer_client
			ORDER BY customer_client.id`

			rows, searchErr := apiClient.Search(mccID, query)
			if searchErr == nil {
				for _, raw := range rows {
					var row api.CustomerClientRow
					if json.Unmarshal(raw, &row) == nil && !row.CustomerClient.Manager {
						accounts = append(accounts, row.CustomerClient)
					}
				}
			}
		}

		// Strategy 2: query each accessible customer individually (fallback)
		if len(accounts) == 0 && len(from) > 0 {
			for _, resourceName := range from {
				custID := api.ResourceID(resourceName)
				if custID == mccID {
					continue // skip the MCC itself
				}
				rows, qErr := apiClient.Search(custID, `SELECT customer.id, customer.descriptive_name, customer.currency_code, customer.time_zone, customer.manager, customer.test_account FROM customer`)
				if qErr != nil {
					continue
				}
				for _, raw := range rows {
					var row struct {
						Customer api.CustomerClient `json:"customer"`
					}
					if json.Unmarshal(raw, &row) == nil && !row.Customer.Manager {
						accounts = append(accounts, row.Customer)
					}
				}
			}
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(accounts, output.IsPretty(cmd))
		}
		if len(accounts) == 0 {
			fmt.Println("No client accounts found.")
			return nil
		}

		headers := []string{"ID", "NAME", "CURRENCY", "TIMEZONE", "TEST"}
		rows2 := make([][]string, len(accounts))
		for i, a := range accounts {
			testStr := ""
			if a.TestAccount {
				testStr = "yes"
			}
			rows2[i] = []string{
				a.ID,
				output.Truncate(a.DescriptiveName, 40),
				a.CurrencyCode,
				output.Truncate(a.TimeZone, 30),
				testStr,
			}
		}
		output.PrintTable(headers, rows2)
		return nil
	},
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
	rootCmd.AddCommand(accountsCmd)
}
