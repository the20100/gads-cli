package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/output"
)

var adgroupsCmd = &cobra.Command{
	Use:   "adgroups",
	Short: "Manage Google Ads ad groups",
}

var (
	adgroupAccount    string
	adgroupCampaignID string
	adgroupID         string
)

// ---- adgroups list ----

var adgroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ad groups in a campaign",
	Long: `List all ad groups in a campaign.

Examples:
  gads-cli adgroups list --account=1234567890 --campaign=111222333
  gads-cli adgroups list --account=1234567890 --campaign=111222333 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if adgroupAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if adgroupCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := api.CleanCustomerID(adgroupAccount)

		query := fmt.Sprintf(`SELECT ad_group.id, ad_group.name, ad_group.status, ad_group.type,
			ad_group.cpc_bid_micros, campaign.id, campaign.name
		FROM ad_group
		WHERE ad_group.status != 'REMOVED'
		  AND campaign.id = '%s'
		ORDER BY ad_group.id`, adgroupCampaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var adgroups []api.AdGroupRow
		for _, raw := range rows {
			var row api.AdGroupRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			adgroups = append(adgroups, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(adgroups, output.IsPretty(cmd))
		}
		if len(adgroups) == 0 {
			fmt.Println("No ad groups found.")
			return nil
		}

		headers := []string{"ID", "NAME", "STATUS", "TYPE", "DEFAULT BID"}
		tableRows := make([][]string, len(adgroups))
		for i, r := range adgroups {
			tableRows[i] = []string{
				r.AdGroup.ID,
				output.Truncate(r.AdGroup.Name, 40),
				r.AdGroup.Status,
				formatChannelType(r.AdGroup.Type),
				api.MicrosToCurrency(r.AdGroup.CpcBidMicros),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- adgroups pause ----

var adgroupsPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause an ad group",
	Long: `Set an ad group status to PAUSED.

Examples:
  gads-cli adgroups pause --account=1234567890 --adgroup=444555666`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setAdGroupStatus(adgroupAccount, adgroupID, "PAUSED")
	},
}

// ---- adgroups enable ----

var adgroupsEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable an ad group",
	Long: `Set an ad group status to ENABLED.

Examples:
  gads-cli adgroups enable --account=1234567890 --adgroup=444555666`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setAdGroupStatus(adgroupAccount, adgroupID, "ENABLED")
	},
}

func setAdGroupStatus(account, agID, status string) error {
	if account == "" {
		return fmt.Errorf("--account is required")
	}
	if agID == "" {
		return fmt.Errorf("--adgroup is required")
	}
	cid := api.CleanCustomerID(account)
	resourceName := fmt.Sprintf("customers/%s/adGroups/%s", cid, agID)

	ops := []map[string]any{
		{
			"updateMask": "status",
			"update": map[string]any{
				"resourceName": resourceName,
				"status":       status,
			},
		},
	}
	if _, err := apiClient.MutateAdGroups(cid, ops); err != nil {
		return err
	}
	fmt.Printf("Ad group %s status set to %s.\n", agID, status)
	return nil
}

func init() {
	adgroupsListCmd.Flags().StringVar(&adgroupAccount, "account", "", "Customer account ID (required)")
	adgroupsListCmd.Flags().StringVar(&adgroupCampaignID, "campaign", "", "Campaign ID (required)")

	for _, c := range []*cobra.Command{adgroupsPauseCmd, adgroupsEnableCmd} {
		c.Flags().StringVar(&adgroupAccount, "account", "", "Customer account ID (required)")
		c.Flags().StringVar(&adgroupID, "adgroup", "", "Ad group ID (required)")
	}

	adgroupsCmd.AddCommand(adgroupsListCmd, adgroupsPauseCmd, adgroupsEnableCmd)
	rootCmd.AddCommand(adgroupsCmd)
}
