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

var accountsVerbose bool

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessible customer accounts under the MCC",
	Long: `List all customer accounts accessible under the configured Manager Account (MCC).

Examples:
  gads-cli accounts list
  gads-cli accounts list --json
  gads-cli accounts list --verbose`,
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

		if accountsVerbose {
			fmt.Printf("Manager Account (MCC): %s\n", mccID)
			fmt.Printf("ListAccessibleCustomers returned %d resource(s):\n", len(from))
			for _, r := range from {
				fmt.Printf("  %s\n", r)
			}
			fmt.Println()
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
			if searchErr != nil {
				if accountsVerbose {
					fmt.Printf("[strategy 1] customer_client query failed: %v\n\n", searchErr)
				}
			} else {
				if accountsVerbose {
					fmt.Printf("[strategy 1] customer_client query returned %d rows\n", len(rows))
				}
				for i, raw := range rows {
					var row api.CustomerClientRow
					if parseErr := json.Unmarshal(raw, &row); parseErr != nil {
						if accountsVerbose {
							fmt.Printf("  [row %d] unmarshal failed: %v\n  raw: %s\n", i, parseErr, string(raw))
						}
						continue
					}
					if accountsVerbose {
						fmt.Printf("  [row %d] id=%s name=%q manager=%v level=%d\n",
							i, row.CustomerClient.ID, row.CustomerClient.DescriptiveName,
							row.CustomerClient.Manager, row.CustomerClient.Level)
					}
					if !row.CustomerClient.Manager {
						accounts = append(accounts, row.CustomerClient)
					}
				}
				if accountsVerbose {
					fmt.Println()
				}
			}
		}

		// Strategy 2: query each accessible customer individually (fallback).
		// For top-level accounts not under the configured MCC, retry using
		// the account's own ID as the login-customer-id.
		if len(accounts) == 0 && len(from) > 0 {
			if accountsVerbose {
				fmt.Println("[strategy 2] falling back to per-customer queries")
			}
			for _, resourceName := range from {
				custID := api.ResourceID(resourceName)
				if custID == mccID {
					if accountsVerbose {
						fmt.Printf("  skipping MCC %s\n", custID)
					}
					continue // skip the MCC itself
				}

				q := `SELECT customer.id, customer.descriptive_name, customer.currency_code, customer.time_zone, customer.manager, customer.test_account FROM customer`

				rows, qErr := apiClient.Search(custID, q)
				if qErr != nil {
					if accountsVerbose {
						fmt.Printf("  %s: default login failed (%v), retrying with self-login\n", custID, qErr)
					}
					// Retry using the account's own ID as login-customer-id
					rows, qErr = apiClient.WithLoginID(custID).Search(custID, q)
					if qErr != nil {
						if accountsVerbose {
							fmt.Printf("  %s: self-login also failed: %v\n", custID, qErr)
						}
						continue
					}
					if accountsVerbose {
						fmt.Printf("  %s: self-login succeeded, trying customer_client query\n", custID)
					}
					// This account is accessible — try customer_client to find sub-accounts
					ccQuery := `SELECT customer_client.id, customer_client.descriptive_name,
						customer_client.currency_code, customer_client.time_zone,
						customer_client.manager, customer_client.level, customer_client.hidden,
						customer_client.test_account
					FROM customer_client
					ORDER BY customer_client.id`
					ccRows, ccErr := apiClient.WithLoginID(custID).Search(custID, ccQuery)
					if ccErr == nil && len(ccRows) > 0 {
						if accountsVerbose {
							fmt.Printf("  %s: customer_client returned %d rows\n", custID, len(ccRows))
						}
						for i, raw := range ccRows {
							var row api.CustomerClientRow
							if parseErr := json.Unmarshal(raw, &row); parseErr != nil {
								if accountsVerbose {
									fmt.Printf("    [row %d] unmarshal failed: %v\n    raw: %s\n", i, parseErr, string(raw))
								}
								continue
							}
							if accountsVerbose {
								fmt.Printf("    [row %d] id=%s name=%q manager=%v level=%d\n",
									i, row.CustomerClient.ID, row.CustomerClient.DescriptiveName,
									row.CustomerClient.Manager, row.CustomerClient.Level)
							}
							if !row.CustomerClient.Manager {
								accounts = append(accounts, row.CustomerClient)
							}
						}
						continue
					}
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
	accountsListCmd.Flags().BoolVar(&accountsVerbose, "verbose", false, "Show diagnostic info (accessible customers, strategy errors)")
	accountsCmd.AddCommand(accountsListCmd)
	rootCmd.AddCommand(accountsCmd)
}
