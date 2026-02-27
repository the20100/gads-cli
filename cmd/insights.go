package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/client"
	"github.com/the20100/gads-cli/internal/output"
)

var insightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Performance reporting via GAQL",
}

var (
	insightsAccount    string
	insightsCampaignID string
	insightsDays       int
	insightsStart      string
	insightsEnd        string
)

// buildDateRange returns a GAQL WHERE clause fragment for the date range.
func buildDateRange(days int, start, end string) string {
	if start != "" && end != "" {
		return fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", start, end)
	}
	if days <= 0 {
		days = 30
	}
	endDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	return fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", startDate, endDate)
}

// ---- insights campaigns ----

var insightsCampaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Campaign performance: impressions, clicks, cost, CTR, CPC, conversions, ROAS",
	Long: `Show campaign performance metrics for a given date range.

Examples:
  gads-cli insights campaigns --account=1234567890 --days=30
  gads-cli insights campaigns --account=1234567890 --start=2024-01-01 --end=2024-01-31
  gads-cli insights campaigns --account=1234567890 --days=7 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		cid := client.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsDays, insightsStart, insightsEnd)

		query := fmt.Sprintf(`SELECT campaign.id, campaign.name,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc, metrics.conversions, metrics.conversions_value
		FROM campaign
		WHERE %s
		  AND campaign.status != 'REMOVED'
		ORDER BY metrics.cost_micros DESC`, dateFilter)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var results []client.InsightsCampaignRow
		for _, raw := range rows {
			var row client.InsightsCampaignRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			results = append(results, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(results, output.IsPretty(cmd))
		}
		if len(results) == 0 {
			fmt.Println("No campaign data found for the specified period.")
			return nil
		}

		headers := []string{"ID", "NAME", "IMPRESSIONS", "CLICKS", "COST", "CTR", "CPC", "CONV", "ROAS"}
		tableRows := make([][]string, len(results))
		for i, r := range results {
			tableRows[i] = []string{
				r.Campaign.ID,
				output.Truncate(r.Campaign.Name, 30),
				client.FormatMetricInt(r.Metrics.Impressions),
				client.FormatMetricInt(r.Metrics.Clicks),
				client.MicrosToCurrency(r.Metrics.CostMicros),
				client.FormatCTR(r.Metrics.Ctr),
				client.MicrosToCurrency(r.Metrics.AverageCpc),
				fmt.Sprintf("%.1f", r.Metrics.Conversions),
				client.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- insights adgroups ----

var insightsAdGroupsCmd = &cobra.Command{
	Use:   "adgroups",
	Short: "Ad group performance metrics",
	Long: `Show ad group performance metrics for a given date range.

Examples:
  gads-cli insights adgroups --account=1234567890 --campaign=111222333 --days=30
  gads-cli insights adgroups --account=1234567890 --campaign=111222333 --start=2024-01-01 --end=2024-01-31`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if insightsCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := client.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsDays, insightsStart, insightsEnd)

		query := fmt.Sprintf(`SELECT campaign.id, ad_group.id, ad_group.name,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc, metrics.conversions, metrics.conversions_value
		FROM ad_group
		WHERE %s
		  AND campaign.id = '%s'
		  AND ad_group.status != 'REMOVED'
		ORDER BY metrics.cost_micros DESC`, dateFilter, insightsCampaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var results []client.InsightsAdGroupRow
		for _, raw := range rows {
			var row client.InsightsAdGroupRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			results = append(results, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(results, output.IsPretty(cmd))
		}
		if len(results) == 0 {
			fmt.Println("No ad group data found for the specified period.")
			return nil
		}

		headers := []string{"ID", "NAME", "IMPRESSIONS", "CLICKS", "COST", "CTR", "CPC", "CONV"}
		tableRows := make([][]string, len(results))
		for i, r := range results {
			tableRows[i] = []string{
				r.AdGroup.ID,
				output.Truncate(r.AdGroup.Name, 36),
				client.FormatMetricInt(r.Metrics.Impressions),
				client.FormatMetricInt(r.Metrics.Clicks),
				client.MicrosToCurrency(r.Metrics.CostMicros),
				client.FormatCTR(r.Metrics.Ctr),
				client.MicrosToCurrency(r.Metrics.AverageCpc),
				fmt.Sprintf("%.1f", r.Metrics.Conversions),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- insights keywords ----

var insightsKeywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Keyword performance metrics",
	Long: `Show keyword performance metrics for a given date range.

Examples:
  gads-cli insights keywords --account=1234567890 --campaign=111222333 --days=30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if insightsCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := client.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsDays, insightsStart, insightsEnd)

		query := fmt.Sprintf(`SELECT ad_group_criterion.keyword.text,
			ad_group_criterion.keyword.match_type,
			ad_group.id, ad_group.name, campaign.id,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc, metrics.conversions, metrics.conversions_value
		FROM keyword_view
		WHERE %s
		  AND campaign.id = '%s'
		  AND ad_group_criterion.status != 'REMOVED'
		ORDER BY metrics.cost_micros DESC`, dateFilter, insightsCampaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var results []client.InsightsKeywordRow
		for _, raw := range rows {
			var row client.InsightsKeywordRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			results = append(results, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(results, output.IsPretty(cmd))
		}
		if len(results) == 0 {
			fmt.Println("No keyword data found for the specified period.")
			return nil
		}

		headers := []string{"KEYWORD", "MATCH", "IMPRESSIONS", "CLICKS", "COST", "CTR", "CPC", "CONV"}
		tableRows := make([][]string, len(results))
		for i, r := range results {
			tableRows[i] = []string{
				output.Truncate(r.AdGroupCriterion.Keyword.Text, 30),
				r.AdGroupCriterion.Keyword.MatchType,
				client.FormatMetricInt(r.Metrics.Impressions),
				client.FormatMetricInt(r.Metrics.Clicks),
				client.MicrosToCurrency(r.Metrics.CostMicros),
				client.FormatCTR(r.Metrics.Ctr),
				client.MicrosToCurrency(r.Metrics.AverageCpc),
				fmt.Sprintf("%.1f", r.Metrics.Conversions),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- insights search-terms ----

var insightsSearchTermsCmd = &cobra.Command{
	Use:   "search-terms",
	Short: "Search terms report",
	Long: `Show the search terms that triggered your ads.

Examples:
  gads-cli insights search-terms --account=1234567890 --campaign=111222333 --days=30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if insightsCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := client.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsDays, insightsStart, insightsEnd)

		query := fmt.Sprintf(`SELECT search_term_view.search_term, search_term_view.status,
			campaign.id, campaign.name, ad_group.id, ad_group.name,
			metrics.impressions, metrics.clicks, metrics.cost_micros, metrics.ctr
		FROM search_term_view
		WHERE %s
		  AND campaign.id = '%s'
		ORDER BY metrics.impressions DESC`, dateFilter, insightsCampaignID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var results []client.SearchTermRow
		for _, raw := range rows {
			var row client.SearchTermRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			results = append(results, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(results, output.IsPretty(cmd))
		}
		if len(results) == 0 {
			fmt.Println("No search term data found for the specified period.")
			return nil
		}

		headers := []string{"SEARCH TERM", "STATUS", "IMPRESSIONS", "CLICKS", "COST", "CTR", "AD GROUP"}
		tableRows := make([][]string, len(results))
		for i, r := range results {
			tableRows[i] = []string{
				output.Truncate(r.SearchTermView.SearchTerm, 40),
				strings.ToLower(r.SearchTermView.Status),
				client.FormatMetricInt(r.Metrics.Impressions),
				client.FormatMetricInt(r.Metrics.Clicks),
				client.MicrosToCurrency(r.Metrics.CostMicros),
				client.FormatCTR(r.Metrics.Ctr),
				output.Truncate(r.AdGroup.Name, 24),
			}
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

func init() {
	// All insights subcommands share these flags
	for _, c := range []*cobra.Command{
		insightsCampaignsCmd, insightsAdGroupsCmd,
		insightsKeywordsCmd, insightsSearchTermsCmd,
	} {
		c.Flags().StringVar(&insightsAccount, "account", "", "Customer account ID (required)")
		c.Flags().IntVar(&insightsDays, "days", 30, "Number of days to look back (default 30)")
		c.Flags().StringVar(&insightsStart, "start", "", "Start date YYYY-MM-DD (overrides --days)")
		c.Flags().StringVar(&insightsEnd, "end", "", "End date YYYY-MM-DD (overrides --days)")
	}
	for _, c := range []*cobra.Command{insightsAdGroupsCmd, insightsKeywordsCmd, insightsSearchTermsCmd} {
		c.Flags().StringVar(&insightsCampaignID, "campaign", "", "Campaign ID (required)")
	}

	insightsCmd.AddCommand(insightsCampaignsCmd, insightsAdGroupsCmd, insightsKeywordsCmd, insightsSearchTermsCmd)
	rootCmd.AddCommand(insightsCmd)
}
