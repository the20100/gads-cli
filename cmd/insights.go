package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
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
	insightsPeriod     string
)

// parsePeriod converts a period shorthand to (start, end) date strings (YYYY-MM-DD).
//
// Supported formats:
//
//	last7d, last14d, last30d, last90d … any lastNd
//	7d, 30d … any Nd shorthand
//	lastNm  (months), Ny / lastNy (years)
//	lastWeek, currentWeek, lastMonth, currentMonth, lastYear, currentYear
//	today, yesterday
//	2024, 2025 … any 4-digit year
func parsePeriod(period string) (start, end string) {
	now := time.Now()
	today := now.Format("2006-01-02")
	p := strings.ToLower(strings.TrimSpace(period))

	switch p {
	case "today":
		return today, today
	case "yesterday":
		d := now.AddDate(0, 0, -1).Format("2006-01-02")
		return d, d
	case "lastweek":
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		lastSunday := now.AddDate(0, 0, -wd)
		lastMonday := lastSunday.AddDate(0, 0, -6)
		return lastMonday.Format("2006-01-02"), lastSunday.Format("2006-01-02")
	case "currentweek", "thisweek":
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		monday := now.AddDate(0, 0, -(wd - 1))
		return monday.Format("2006-01-02"), today
	case "lastmonth":
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastDay := first.AddDate(0, 0, -1)
		firstDay := time.Date(lastDay.Year(), lastDay.Month(), 1, 0, 0, 0, 0, time.UTC)
		return firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02")
	case "currentmonth", "thismonth":
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return first.Format("2006-01-02"), today
	case "lastyear":
		y := now.Year() - 1
		return fmt.Sprintf("%d-01-01", y), fmt.Sprintf("%d-12-31", y)
	case "currentyear", "thisyear":
		return fmt.Sprintf("%d-01-01", now.Year()), today
	case "1y", "last1y":
		return now.AddDate(-1, 0, 0).Format("2006-01-02"), today
	}

	// lastNd → last N days
	if strings.HasPrefix(p, "last") && strings.HasSuffix(p, "d") {
		if n, err := strconv.Atoi(p[4 : len(p)-1]); err == nil && n > 0 {
			return now.AddDate(0, 0, -n).Format("2006-01-02"), today
		}
	}
	// Nd shorthand
	if strings.HasSuffix(p, "d") {
		if n, err := strconv.Atoi(p[:len(p)-1]); err == nil && n > 0 {
			return now.AddDate(0, 0, -n).Format("2006-01-02"), today
		}
	}
	// lastNm / Nm → last N months
	if strings.HasSuffix(p, "m") {
		prefix := strings.TrimPrefix(p[:len(p)-1], "last")
		if n, err := strconv.Atoi(prefix); err == nil && n > 0 {
			return now.AddDate(0, -n, 0).Format("2006-01-02"), today
		}
	}
	// lastNy / Ny → last N years
	if strings.HasSuffix(p, "y") {
		prefix := strings.TrimPrefix(p[:len(p)-1], "last")
		if n, err := strconv.Atoi(prefix); err == nil && n > 0 {
			return now.AddDate(-n, 0, 0).Format("2006-01-02"), today
		}
	}
	// 4-digit year: 2024 → full year, current year → up to today
	if len(p) == 4 {
		if year, err := strconv.Atoi(p); err == nil && year >= 2000 && year <= now.Year()+1 {
			if year < now.Year() {
				return fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
			}
			return fmt.Sprintf("%d-01-01", year), today
		}
	}

	return "", ""
}

// buildDateRange returns a GAQL WHERE clause fragment for the date range.
// Priority: --period > --start/--end > --days (default 30).
func buildDateRange(period string, days int, start, end string) string {
	if period != "" {
		if s, e := parsePeriod(period); s != "" {
			return fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", s, e)
		}
	}
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
  gads-cli insights campaigns --account=1234567890 --period=last30d
  gads-cli insights campaigns --account=1234567890 --period=lastMonth
  gads-cli insights campaigns --account=1234567890 --period=2025
  gads-cli insights campaigns --account=1234567890 --start=2024-01-01 --end=2024-01-31
  gads-cli insights campaigns --account=1234567890 --days=7 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

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

		var results []api.InsightsCampaignRow
		for _, raw := range rows {
			var row api.InsightsCampaignRow
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
				api.FormatMetricInt(r.Metrics.Impressions),
				api.FormatMetricInt(r.Metrics.Clicks),
				api.MicrosToCurrency(r.Metrics.CostMicros),
				api.FormatCTR(r.Metrics.Ctr),
				api.MicrosToCurrency(r.Metrics.AverageCpc),
				fmt.Sprintf("%.1f", r.Metrics.Conversions),
				api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros),
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
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

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

		var results []api.InsightsAdGroupRow
		for _, raw := range rows {
			var row api.InsightsAdGroupRow
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
				api.FormatMetricInt(r.Metrics.Impressions),
				api.FormatMetricInt(r.Metrics.Clicks),
				api.MicrosToCurrency(r.Metrics.CostMicros),
				api.FormatCTR(r.Metrics.Ctr),
				api.MicrosToCurrency(r.Metrics.AverageCpc),
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
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

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

		var results []api.InsightsKeywordRow
		for _, raw := range rows {
			var row api.InsightsKeywordRow
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
				api.FormatMetricInt(r.Metrics.Impressions),
				api.FormatMetricInt(r.Metrics.Clicks),
				api.MicrosToCurrency(r.Metrics.CostMicros),
				api.FormatCTR(r.Metrics.Ctr),
				api.MicrosToCurrency(r.Metrics.AverageCpc),
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
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

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

		var results []api.SearchTermRow
		for _, raw := range rows {
			var row api.SearchTermRow
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
				api.FormatMetricInt(r.Metrics.Impressions),
				api.FormatMetricInt(r.Metrics.Clicks),
				api.MicrosToCurrency(r.Metrics.CostMicros),
				api.FormatCTR(r.Metrics.Ctr),
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
		c.Flags().StringVar(&insightsPeriod, "period", "", "Preset period: last7d, last30d, lastWeek, currentWeek, lastMonth, currentMonth, lastYear, currentYear, 2025, last3m, 1y …")
		c.Flags().IntVar(&insightsDays, "days", 30, "Number of days to look back (default 30, ignored when --period is set)")
		c.Flags().StringVar(&insightsStart, "start", "", "Start date YYYY-MM-DD (overrides --days, ignored when --period is set)")
		c.Flags().StringVar(&insightsEnd, "end", "", "End date YYYY-MM-DD (overrides --days, ignored when --period is set)")
	}
	for _, c := range []*cobra.Command{insightsAdGroupsCmd, insightsKeywordsCmd, insightsSearchTermsCmd} {
		c.Flags().StringVar(&insightsCampaignID, "campaign", "", "Campaign ID (required)")
	}

	insightsCmd.AddCommand(insightsCampaignsCmd, insightsAdGroupsCmd, insightsKeywordsCmd, insightsSearchTermsCmd)
	rootCmd.AddCommand(insightsCmd)
}
