package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/output"
)

var keywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Manage Google Ads keywords",
}

var (
	keywordAccount    string
	keywordCampaignID string
	keywordAdGroupID  string
	keywordText       string
	keywordMatchType  string
	keywordID         string // format: <adGroupId>~<criterionId>
)

// ---- keywords list ----

var keywordsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List keywords in a campaign",
	Long: `List keywords with match type, status, and quality score.

Examples:
  gads-cli keywords list --account=1234567890 --campaign=111222333
  gads-cli keywords list --account=1234567890 --campaign=111222333 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if keywordAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if keywordCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := api.CleanCustomerID(keywordAccount)

		query := fmt.Sprintf(`SELECT ad_group_criterion.criterion_id,
			ad_group_criterion.keyword.text, ad_group_criterion.keyword.match_type,
			ad_group_criterion.status, ad_group_criterion.negative,
			ad_group_criterion.quality_info.quality_score,
			ad_group_criterion.cpc_bid_micros,
			ad_group.id, ad_group.name, campaign.id
		FROM keyword_view
		WHERE ad_group_criterion.status != 'REMOVED'
		  AND campaign.id = '%s'
		ORDER BY ad_group_criterion.criterion_id`, keywordCampaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var keywords []api.KeywordRow
		for _, raw := range rows {
			var row api.KeywordRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			keywords = append(keywords, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(keywords, output.IsPretty(cmd))
		}
		if len(keywords) == 0 {
			fmt.Println("No keywords found.")
			return nil
		}

		headers := []string{"ID", "KEYWORD", "MATCH", "STATUS", "QS", "BID", "AD GROUP"}
		tableRows := make([][]string, len(keywords))
		for i, r := range keywords {
			qs := "-"
			if r.AdGroupCriterion.QualityInfo.QualityScore > 0 {
				qs = fmt.Sprintf("%d", r.AdGroupCriterion.QualityInfo.QualityScore)
			}
			negLabel := ""
			if r.AdGroupCriterion.Negative {
				negLabel = " [neg]"
			}
			tableRows[i] = []string{
				r.AdGroupCriterion.CriterionID,
				output.Truncate(r.AdGroupCriterion.Keyword.Text+negLabel, 40),
				r.AdGroupCriterion.Keyword.MatchType,
				r.AdGroupCriterion.Status,
				qs,
				api.MicrosToCurrency(r.AdGroupCriterion.CpcBidMicros),
				output.Truncate(r.AdGroup.Name, 24),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- keywords add ----

var keywordsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a keyword to an ad group",
	Long: `Add a new keyword to an ad group.

Examples:
  gads-cli keywords add --account=1234567890 --adgroup=444555666 --keyword="running shoes" --match-type=PHRASE
  gads-cli keywords add --account=1234567890 --adgroup=444555666 --keyword="buy sneakers" --match-type=EXACT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if keywordAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if keywordAdGroupID == "" {
			return fmt.Errorf("--adgroup is required")
		}
		if keywordText == "" {
			return fmt.Errorf("--keyword is required")
		}
		if keywordMatchType == "" {
			return fmt.Errorf("--match-type is required (BROAD, PHRASE, or EXACT)")
		}
		mt := strings.ToUpper(keywordMatchType)
		if mt != "BROAD" && mt != "PHRASE" && mt != "EXACT" {
			return fmt.Errorf("--match-type must be BROAD, PHRASE, or EXACT")
		}

		cid := api.CleanCustomerID(keywordAccount)
		adGroupResourceName := fmt.Sprintf("customers/%s/adGroups/%s", cid, keywordAdGroupID)

		ops := []map[string]any{
			{
				"create": map[string]any{
					"adGroup": adGroupResourceName,
					"status":  "ENABLED",
					"keyword": map[string]any{
						"text":      keywordText,
						"matchType": mt,
					},
				},
			},
		}
		resp, err := apiClient.MutateAdGroupCriteria(cid, ops)
		if err != nil {
			return err
		}
		if len(resp.Results) > 0 {
			fmt.Printf("Keyword added: \"%s\" [%s]\n", keywordText, mt)
			fmt.Printf("Resource: %s\n", resp.Results[0].ResourceName)
		}
		return nil
	},
}

// ---- keywords pause ----

var keywordsPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause a keyword",
	Long: `Pause a keyword. Provide the keyword ID in the format <adGroupId>~<criterionId>.

The keyword ID is shown in the 'ID' column of 'keywords list' — use the
compound key format: <adGroupId>~<criterionId>

Examples:
  gads-cli keywords pause --account=1234567890 --keyword=444555666~12345`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setKeywordStatus(keywordAccount, keywordID, "PAUSED")
	},
}

// ---- keywords remove ----

var keywordsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a keyword",
	Long: `Remove (soft-delete) a keyword. Provide the keyword ID as <adGroupId>~<criterionId>.

Examples:
  gads-cli keywords remove --account=1234567890 --keyword=444555666~12345`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if keywordAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if keywordID == "" {
			return fmt.Errorf("--keyword is required (format: <adGroupId>~<criterionId>)")
		}
		cid := api.CleanCustomerID(keywordAccount)
		resourceName := fmt.Sprintf("customers/%s/adGroupCriteria/%s", cid, keywordID)

		ops := []map[string]any{
			{"remove": resourceName},
		}
		if _, err := apiClient.MutateAdGroupCriteria(cid, ops); err != nil {
			return err
		}
		fmt.Printf("Keyword %s removed.\n", keywordID)
		return nil
	},
}

func setKeywordStatus(account, kwID, status string) error {
	if account == "" {
		return fmt.Errorf("--account is required")
	}
	if kwID == "" {
		return fmt.Errorf("--keyword is required (format: <adGroupId>~<criterionId>)")
	}
	cid := api.CleanCustomerID(account)
	resourceName := fmt.Sprintf("customers/%s/adGroupCriteria/%s", cid, kwID)

	ops := []map[string]any{
		{
			"updateMask": "status",
			"update": map[string]any{
				"resourceName": resourceName,
				"status":       status,
			},
		},
	}
	if _, err := apiClient.MutateAdGroupCriteria(cid, ops); err != nil {
		return err
	}
	fmt.Printf("Keyword %s status set to %s.\n", kwID, status)
	return nil
}

func init() {
	keywordsListCmd.Flags().StringVar(&keywordAccount, "account", "", "Customer account ID (required)")
	keywordsListCmd.Flags().StringVar(&keywordCampaignID, "campaign", "", "Campaign ID (required)")

	keywordsAddCmd.Flags().StringVar(&keywordAccount, "account", "", "Customer account ID (required)")
	keywordsAddCmd.Flags().StringVar(&keywordAdGroupID, "adgroup", "", "Ad group ID (required)")
	keywordsAddCmd.Flags().StringVar(&keywordText, "keyword", "", "Keyword text (required)")
	keywordsAddCmd.Flags().StringVar(&keywordMatchType, "match-type", "", "Match type: BROAD, PHRASE, or EXACT (required)")

	for _, c := range []*cobra.Command{keywordsPauseCmd, keywordsRemoveCmd} {
		c.Flags().StringVar(&keywordAccount, "account", "", "Customer account ID (required)")
		c.Flags().StringVar(&keywordID, "keyword", "", "Keyword ID in format <adGroupId>~<criterionId> (required)")
	}

	keywordsCmd.AddCommand(keywordsListCmd, keywordsAddCmd, keywordsPauseCmd, keywordsRemoveCmd)
	rootCmd.AddCommand(keywordsCmd)
}
