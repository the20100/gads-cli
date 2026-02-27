package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/output"
)

var campaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Manage Google Ads campaigns",
}

var (
	campaignAccount  string
	campaignID       string
	campaignBudgetAm int64
)

// ---- campaigns list ----

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List campaigns in an account",
	Long: `List campaigns with status, budget, and type.

Examples:
  gads-cli campaigns list --account=1234567890
  gads-cli campaigns list --account=1234567890 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if campaignAccount == "" {
			return fmt.Errorf("--account is required")
		}
		cid := api.CleanCustomerID(campaignAccount)

		query := `SELECT campaign.id, campaign.name, campaign.status,
			campaign.advertising_channel_type, campaign.bidding_strategy_type,
			campaign.start_date, campaign.end_date,
			campaign_budget.id, campaign_budget.amount_micros
		FROM campaign
		WHERE campaign.status != 'REMOVED'
		ORDER BY campaign.id`

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var campaigns []api.CampaignRow
		for _, raw := range rows {
			var row api.CampaignRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			campaigns = append(campaigns, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(campaigns, output.IsPretty(cmd))
		}
		if len(campaigns) == 0 {
			fmt.Println("No campaigns found.")
			return nil
		}

		headers := []string{"ID", "NAME", "STATUS", "TYPE", "DAILY BUDGET", "START", "END"}
		tableRows := make([][]string, len(campaigns))
		for i, r := range campaigns {
			tableRows[i] = []string{
				r.Campaign.ID,
				output.Truncate(r.Campaign.Name, 36),
				r.Campaign.Status,
				formatChannelType(r.Campaign.AdvertisingChannelType),
				api.MicrosToCurrency(r.CampaignBudget.AmountMicros),
				r.Campaign.StartDate,
				emptyOrValue(r.Campaign.EndDate),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- campaigns get ----

var campaignsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get full details of a campaign",
	Long: `Get detailed information about a specific campaign.

Examples:
  gads-cli campaigns get --account=1234567890 --campaign=111222333`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if campaignAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if campaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := api.CleanCustomerID(campaignAccount)

		query := fmt.Sprintf(`SELECT campaign.id, campaign.name, campaign.status,
			campaign.advertising_channel_type, campaign.bidding_strategy_type,
			campaign.start_date, campaign.end_date,
			campaign_budget.id, campaign_budget.amount_micros
		FROM campaign
		WHERE campaign.id = '%s'`, campaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("campaign %s not found", campaignID)
		}

		var row api.CampaignRow
		if err := json.Unmarshal(rows[0], &row); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(row, output.IsPretty(cmd))
		}

		output.PrintKeyValue([][]string{
			{"ID", row.Campaign.ID},
			{"Name", row.Campaign.Name},
			{"Status", row.Campaign.Status},
			{"Type", formatChannelType(row.Campaign.AdvertisingChannelType)},
			{"Bidding", row.Campaign.BiddingStrategyType},
			{"Daily Budget", api.MicrosToCurrency(row.CampaignBudget.AmountMicros)},
			{"Budget ID", row.CampaignBudget.ID},
			{"Start Date", row.Campaign.StartDate},
			{"End Date", emptyOrValue(row.Campaign.EndDate)},
			{"Resource", row.Campaign.ResourceName},
		})
		return nil
	},
}

// ---- campaigns pause ----

var campaignsPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause a campaign",
	Long: `Set a campaign status to PAUSED.

Examples:
  gads-cli campaigns pause --account=1234567890 --campaign=111222333`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setCampaignStatus(campaignAccount, campaignID, "PAUSED")
	},
}

// ---- campaigns enable ----

var campaignsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a campaign",
	Long: `Set a campaign status to ENABLED.

Examples:
  gads-cli campaigns enable --account=1234567890 --campaign=111222333`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setCampaignStatus(campaignAccount, campaignID, "ENABLED")
	},
}

func setCampaignStatus(account, campID, status string) error {
	if account == "" {
		return fmt.Errorf("--account is required")
	}
	if campID == "" {
		return fmt.Errorf("--campaign is required")
	}
	cid := api.CleanCustomerID(account)
	resourceName := fmt.Sprintf("customers/%s/campaigns/%s", cid, campID)

	ops := []map[string]any{
		{
			"updateMask": "status",
			"update": map[string]any{
				"resourceName": resourceName,
				"status":       status,
			},
		},
	}
	if _, err := apiClient.MutateCampaigns(cid, ops); err != nil {
		return err
	}
	fmt.Printf("Campaign %s status set to %s.\n", campID, status)
	return nil
}

// ---- campaigns budget ----

var campaignsBudgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Update the daily budget of a campaign",
	Long: `Update the daily budget for a campaign. Amount is in micros (1 unit = 1,000,000 micros).

Examples:
  gads-cli campaigns budget --account=1234567890 --campaign=111222333 --amount=5000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if campaignAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if campaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		if campaignBudgetAm <= 0 {
			return fmt.Errorf("--amount is required and must be positive (in micros)")
		}
		cid := api.CleanCustomerID(campaignAccount)

		// First fetch the budget resource name from the campaign
		query := fmt.Sprintf(`SELECT campaign.id, campaign_budget.id
		FROM campaign
		WHERE campaign.id = '%s'`, campaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("campaign %s not found", campaignID)
		}
		var row api.CampaignRow
		if err := json.Unmarshal(rows[0], &row); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		if row.CampaignBudget.ID == "" {
			return fmt.Errorf("could not find budget for campaign %s", campaignID)
		}

		budgetResourceName := fmt.Sprintf("customers/%s/campaignBudgets/%s", cid, row.CampaignBudget.ID)
		ops := []map[string]any{
			{
				"updateMask": "amountMicros",
				"update": map[string]any{
					"resourceName": budgetResourceName,
					"amountMicros": strconv.FormatInt(campaignBudgetAm, 10),
				},
			},
		}
		if _, err := apiClient.MutateCampaignBudgets(cid, ops); err != nil {
			return err
		}
		fmt.Printf("Campaign %s budget updated to %s (budget ID: %s).\n",
			campaignID, api.MicrosToCurrency(strconv.FormatInt(campaignBudgetAm, 10)), row.CampaignBudget.ID)
		return nil
	},
}

func init() {
	// Shared flags
	for _, c := range []*cobra.Command{campaignsListCmd, campaignsGetCmd, campaignsPauseCmd, campaignsEnableCmd, campaignsBudgetCmd} {
		c.Flags().StringVar(&campaignAccount, "account", "", "Customer account ID (required)")
	}
	for _, c := range []*cobra.Command{campaignsGetCmd, campaignsPauseCmd, campaignsEnableCmd, campaignsBudgetCmd} {
		c.Flags().StringVar(&campaignID, "campaign", "", "Campaign ID (required)")
	}
	campaignsBudgetCmd.Flags().Int64Var(&campaignBudgetAm, "amount", 0, "New daily budget in micros (e.g. 5000000 = 5.00)")

	campaignsCmd.AddCommand(campaignsListCmd, campaignsGetCmd, campaignsPauseCmd, campaignsEnableCmd, campaignsBudgetCmd)
	rootCmd.AddCommand(campaignsCmd)
}

func formatChannelType(t string) string {
	return strings.ToLower(strings.ReplaceAll(t, "_", " "))
}

func emptyOrValue(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
