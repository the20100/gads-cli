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
	insightsAll        bool
	insightsVerbose    bool
	insightsPreset     string
	insightsFields     string
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

Presets (--preset):
  default     Campaign name, status, impressions, clicks, cost, CTR, CPC, conversions, ROAS
  performance + absolute/top impression share, conversion rate, cost/conv
  conversions Focus on conversions, value, view-through, conv rate, cost/conv, ROAS
  full        All available fields

Field IDs for --fields (comma-separated):
  Dimensions: campaign_id, campaign_name, campaign_status, campaign_type
  Metrics:    impressions, clicks, cost, ctr, cpc, conversions, conv_value, roas,
              abs_top_imp_pct, top_imp_pct, view_through_conv, conv_rate, cost_per_conv,
              search_imp_share

Examples:
  gads-cli insights campaigns --account=1234567890 --period=last30d
  gads-cli insights campaigns --account=1234567890 --period=lastMonth --preset=performance
  gads-cli insights campaigns --account=1234567890 --period=2025 --preset=conversions
  gads-cli insights campaigns --account=1234567890 --start=2024-01-01 --end=2024-01-31
  gads-cli insights campaigns --account=1234567890 --days=7 --fields=campaign_name,impressions,clicks,cost,roas
  gads-cli insights campaigns --account=1234567890 --days=7 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

		impressionsFilter := ""
		if !insightsAll {
			impressionsFilter = "\n		  AND metrics.impressions > 0"
		}
		query := fmt.Sprintf(`SELECT
			campaign.id, campaign.name, campaign.status, campaign.advertising_channel_type,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc,
			metrics.conversions, metrics.conversions_value,
			metrics.absolute_top_impression_percentage, metrics.top_impression_percentage,
			metrics.view_through_conversions, metrics.cost_per_conversion,
			metrics.conversions_from_interactions_rate, metrics.search_impression_share
		FROM campaign
		WHERE %s
		  AND campaign.status != 'REMOVED'%s
		ORDER BY metrics.cost_micros DESC`, dateFilter, impressionsFilter)

		if insightsVerbose {
			fmt.Printf("[verbose] account: %s\n[verbose] query:\n%s\n\n", cid, query)
		}
		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if insightsVerbose {
			fmt.Printf("[verbose] API returned %d raw rows\n\n", len(rows))
		}

		var results []api.InsightsCampaignRow
		for _, raw := range rows {
			var row api.InsightsCampaignRow
			if err := json.Unmarshal(raw, &row); err != nil {
				if insightsVerbose {
					fmt.Printf("[verbose] unmarshal error: %v\nraw: %s\n", err, string(raw))
				}
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

		cols := resolveCampaignCols(insightsPreset, insightsFields)
		headers := campaignHeaders(cols)
		tableRows := make([][]string, len(results))
		for i, r := range results {
			r := r
			row := make([]string, len(cols))
			for j, col := range cols {
				row[j] = col.Format(&r)
			}
			tableRows[i] = row
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

Presets (--preset):
  default     Ad group name, status, impressions, clicks, cost, CTR, CPC, conversions, ROAS
  performance + campaign name, conv rate, cost/conv, impression share
  conversions Focus on conversions, value, view-through, conv rate, cost/conv, ROAS
  full        All available fields

Field IDs for --fields (comma-separated):
  Dimensions: campaign_name, adgroup_id, adgroup_name, adgroup_status
  Metrics:    impressions, clicks, cost, ctr, cpc, conversions, conv_value, roas,
              abs_top_imp_pct, top_imp_pct, view_through_conv, conv_rate, cost_per_conv,
              search_imp_share

Examples:
  gads-cli insights adgroups --account=1234567890 --campaign=111222333 --days=30
  gads-cli insights adgroups --account=1234567890 --campaign=111222333 --preset=performance
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

		impressionsFilter := ""
		if !insightsAll {
			impressionsFilter = "\n		  AND metrics.impressions > 0"
		}
		query := fmt.Sprintf(`SELECT
			campaign.id, campaign.name,
			ad_group.id, ad_group.name, ad_group.status,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc,
			metrics.conversions, metrics.conversions_value,
			metrics.absolute_top_impression_percentage, metrics.top_impression_percentage,
			metrics.view_through_conversions, metrics.cost_per_conversion,
			metrics.conversions_from_interactions_rate, metrics.search_impression_share
		FROM ad_group
		WHERE %s
		  AND campaign.id = '%s'
		  AND ad_group.status != 'REMOVED'%s
		ORDER BY metrics.cost_micros DESC`, dateFilter, insightsCampaignID, impressionsFilter)

		if insightsVerbose {
			fmt.Printf("[verbose] account: %s\n[verbose] query:\n%s\n\n", cid, query)
		}
		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if insightsVerbose {
			fmt.Printf("[verbose] API returned %d raw rows\n\n", len(rows))
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

		cols := resolveAdGroupCols(insightsPreset, insightsFields)
		headers := adGroupHeaders(cols)
		tableRows := make([][]string, len(results))
		for i, r := range results {
			r := r
			row := make([]string, len(cols))
			for j, col := range cols {
				row[j] = col.Format(&r)
			}
			tableRows[i] = row
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

Presets (--preset):
  default     Keyword text, match, status, impressions, clicks, cost, CTR, CPC, conversions, quality score
  performance + campaign name, ad group name, ROAS, conv rate, cost/conv
  conversions Focus on conversions, value, conv rate, cost/conv, ROAS
  full        All available fields

Field IDs for --fields (comma-separated):
  Dimensions: keyword_text, keyword_match, keyword_status, quality_score,
              campaign_name, adgroup_name
  Metrics:    impressions, clicks, cost, ctr, cpc, conversions, conv_value, roas,
              abs_top_imp_pct, top_imp_pct, view_through_conv, conv_rate, cost_per_conv,
              search_imp_share

Examples:
  gads-cli insights keywords --account=1234567890 --campaign=111222333 --days=30
  gads-cli insights keywords --account=1234567890 --campaign=111222333 --preset=performance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if insightsCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

		impressionsFilter := ""
		if !insightsAll {
			impressionsFilter = "\n		  AND metrics.impressions > 0"
		}
		query := fmt.Sprintf(`SELECT
			ad_group_criterion.keyword.text,
			ad_group_criterion.keyword.match_type,
			ad_group_criterion.status,
			ad_group_criterion.quality_info.quality_score,
			ad_group.id, ad_group.name, campaign.id, campaign.name,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc,
			metrics.conversions, metrics.conversions_value,
			metrics.absolute_top_impression_percentage, metrics.top_impression_percentage,
			metrics.view_through_conversions, metrics.cost_per_conversion,
			metrics.conversions_from_interactions_rate, metrics.search_impression_share
		FROM keyword_view
		WHERE %s
		  AND campaign.id = '%s'
		  AND ad_group_criterion.status != 'REMOVED'%s
		ORDER BY metrics.cost_micros DESC`, dateFilter, insightsCampaignID, impressionsFilter)

		if insightsVerbose {
			fmt.Printf("[verbose] account: %s\n[verbose] query:\n%s\n\n", cid, query)
		}
		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if insightsVerbose {
			fmt.Printf("[verbose] API returned %d raw rows\n\n", len(rows))
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

		cols := resolveKeywordCols(insightsPreset, insightsFields)
		headers := keywordHeaders(cols)
		tableRows := make([][]string, len(results))
		for i, r := range results {
			r := r
			row := make([]string, len(cols))
			for j, col := range cols {
				row[j] = col.Format(&r)
			}
			tableRows[i] = row
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

Presets (--preset):
  default     Search term, status, ad group, impressions, clicks, cost, CTR, conversions
  performance + campaign name, CPC, ROAS, conv rate, cost/conv
  conversions Focus on conversions, value, conv rate, cost/conv, ROAS
  full        All available fields

Field IDs for --fields (comma-separated):
  Dimensions: search_term, st_status, campaign_name, adgroup_name
  Metrics:    impressions, clicks, cost, ctr, cpc, conversions, conv_value, roas,
              view_through_conv, conv_rate, cost_per_conv

Examples:
  gads-cli insights search-terms --account=1234567890 --campaign=111222333 --days=30
  gads-cli insights search-terms --account=1234567890 --campaign=111222333 --preset=performance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if insightsCampaignID == "" {
			return fmt.Errorf("--campaign is required")
		}
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

		query := fmt.Sprintf(`SELECT
			search_term_view.search_term, search_term_view.status,
			campaign.id, campaign.name, ad_group.id, ad_group.name,
			metrics.impressions, metrics.clicks, metrics.cost_micros, metrics.ctr,
			metrics.average_cpc, metrics.conversions, metrics.conversions_value,
			metrics.view_through_conversions, metrics.cost_per_conversion,
			metrics.conversions_from_interactions_rate
		FROM search_term_view
		WHERE %s
		  AND campaign.id = '%s'
		ORDER BY metrics.impressions DESC`, dateFilter, insightsCampaignID)

		if insightsVerbose {
			fmt.Printf("[verbose] account: %s\n[verbose] query:\n%s\n\n", cid, query)
		}
		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if insightsVerbose {
			fmt.Printf("[verbose] API returned %d raw rows\n\n", len(rows))
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

		cols := resolveSearchTermCols(insightsPreset, insightsFields)
		headers := searchTermHeaders(cols)
		tableRows := make([][]string, len(results))
		for i, r := range results {
			r := r
			row := make([]string, len(cols))
			for j, col := range cols {
				row[j] = col.Format(&r)
			}
			tableRows[i] = row
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

// ---- insights ads ----

var insightsAdsCmd = &cobra.Command{
	Use:   "ads",
	Short: "Ad-level performance with creative details (RSA headlines, descriptions, URLs)",
	Long: `Show ad-level performance metrics and creative details for a given date range.

Presets (--preset):
  default     Campaign, ad group, ad name, status, type, final URL + core metrics
  performance Campaign, ad group, ad name, status + full metrics + impression share
  creatives   Campaign, ad group, ad name, status, type, URLs, path1/2 + all 15 headlines + 4 descriptions
  full        All dimensions, all RSA/ETA creative fields, all metrics

Field IDs for --fields (comma-separated):
  Dimensions: campaign_name, adgroup_name, ad_id, ad_name, ad_status, ad_type,
              final_url, final_mobile_url, tracking_url, url_suffix,
              display_url, path1, path2
  RSA:        headline1..headline15, headline1_pos..headline15_pos,
              desc1..desc4, desc1_pos..desc4_pos
  ETA legacy: eta_headline1, eta_headline2, eta_headline3, eta_desc1, eta_desc2
  Metrics:    impressions, clicks, cost, ctr, cpc, conversions, conv_value, roas,
              abs_top_imp_pct, top_imp_pct, view_through_conv, conv_rate, cost_per_conv,
              search_imp_share

Examples:
  gads-cli insights ads --account=1234567890 --days=30
  gads-cli insights ads --account=1234567890 --campaign=111222333 --preset=creatives
  gads-cli insights ads --account=1234567890 --campaign=111222333 --preset=performance --period=last30d
  gads-cli insights ads --account=1234567890 --days=7 --fields=campaign_name,ad_name,headline1,headline2,headline3,desc1,desc2
  gads-cli insights ads --account=1234567890 --days=30 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if insightsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		cid := api.CleanCustomerID(insightsAccount)
		dateFilter := buildDateRange(insightsPeriod, insightsDays, insightsStart, insightsEnd)

		impressionsFilter := ""
		if !insightsAll {
			impressionsFilter = "\n		  AND metrics.impressions > 0"
		}
		campaignFilter := ""
		if insightsCampaignID != "" {
			campaignFilter = fmt.Sprintf("\n		  AND campaign.id = '%s'", insightsCampaignID)
		}

		query := fmt.Sprintf(`SELECT
			ad_group_ad.ad.id, ad_group_ad.ad.name,
			ad_group_ad.status, ad_group_ad.ad.type,
			ad_group_ad.ad.final_urls, ad_group_ad.ad.final_mobile_urls,
			ad_group_ad.ad.tracking_url_template, ad_group_ad.ad.final_url_suffix,
			ad_group_ad.ad.display_url,
			ad_group_ad.ad.responsive_search_ad.headlines,
			ad_group_ad.ad.responsive_search_ad.descriptions,
			ad_group_ad.ad.responsive_search_ad.path1,
			ad_group_ad.ad.responsive_search_ad.path2,
			ad_group_ad.ad.expanded_text_ad.headline_part1,
			ad_group_ad.ad.expanded_text_ad.headline_part2,
			ad_group_ad.ad.expanded_text_ad.headline_part3,
			ad_group_ad.ad.expanded_text_ad.description,
			ad_group_ad.ad.expanded_text_ad.description2,
			ad_group.id, ad_group.name,
			campaign.id, campaign.name, campaign.status,
			metrics.impressions, metrics.clicks, metrics.cost_micros,
			metrics.ctr, metrics.average_cpc,
			metrics.conversions, metrics.conversions_value,
			metrics.absolute_top_impression_percentage, metrics.top_impression_percentage,
			metrics.view_through_conversions, metrics.cost_per_conversion,
			metrics.conversions_from_interactions_rate, metrics.search_impression_share
		FROM ad_group_ad
		WHERE %s
		  AND ad_group_ad.status != 'REMOVED'%s%s
		ORDER BY metrics.cost_micros DESC`, dateFilter, campaignFilter, impressionsFilter)

		if insightsVerbose {
			fmt.Printf("[verbose] account: %s\n[verbose] query:\n%s\n\n", cid, query)
		}
		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}
		if insightsVerbose {
			fmt.Printf("[verbose] API returned %d raw rows\n\n", len(rows))
		}

		var results []api.InsightsAdRow
		for _, raw := range rows {
			var row api.InsightsAdRow
			if err := json.Unmarshal(raw, &row); err != nil {
				if insightsVerbose {
					fmt.Printf("[verbose] unmarshal error: %v\nraw: %s\n", err, string(raw))
				}
				continue
			}
			results = append(results, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(results, output.IsPretty(cmd))
		}
		if len(results) == 0 {
			fmt.Println("No ad data found for the specified period.")
			return nil
		}

		cols := resolveAdCols(insightsPreset, insightsFields)
		headers := adHeaders(cols)
		tableRows := make([][]string, len(results))
		for i, r := range results {
			r := r
			row := make([]string, len(cols))
			for j, col := range cols {
				row[j] = col.Format(&r)
			}
			tableRows[i] = row
		}
		output.PrintTable(headers, tableRows)
		return nil
	},
}

func init() {
	allInsightsCmds := []*cobra.Command{
		insightsCampaignsCmd, insightsAdGroupsCmd,
		insightsKeywordsCmd, insightsSearchTermsCmd, insightsAdsCmd,
	}

	// Flags shared by all insights subcommands
	for _, c := range allInsightsCmds {
		c.Flags().StringVar(&insightsAccount, "account", "", "Customer account ID (required)")
		c.Flags().StringVar(&insightsPeriod, "period", "", "Preset period: last7d, last30d, lastWeek, currentWeek, lastMonth, currentMonth, lastYear, currentYear, 2025, last3m, 1y …")
		c.Flags().IntVar(&insightsDays, "days", 30, "Number of days to look back (default 30, ignored when --period is set)")
		c.Flags().StringVar(&insightsStart, "start", "", "Start date YYYY-MM-DD (overrides --days, ignored when --period is set)")
		c.Flags().StringVar(&insightsEnd, "end", "", "End date YYYY-MM-DD (overrides --days, ignored when --period is set)")
		c.Flags().BoolVar(&insightsAll, "all", false, "Include rows with 0 impressions (default: only show rows with activity)")
		c.Flags().BoolVar(&insightsVerbose, "verbose", false, "Print the GAQL query and raw row count for debugging")
		c.Flags().StringVar(&insightsPreset, "preset", "default", "Column preset: default, performance, conversions, full (ads also supports: creatives)")
		c.Flags().StringVar(&insightsFields, "fields", "", "Comma-separated field IDs to display, overrides --preset (e.g. campaign_name,impressions,clicks,cost,roas)")
	}

	// --campaign flag for subcommands that require a campaign filter
	for _, c := range []*cobra.Command{insightsAdGroupsCmd, insightsKeywordsCmd, insightsSearchTermsCmd} {
		c.Flags().StringVar(&insightsCampaignID, "campaign", "", "Campaign ID (required)")
	}
	// --campaign is optional for ads (filters to a specific campaign if provided)
	insightsAdsCmd.Flags().StringVar(&insightsCampaignID, "campaign", "", "Campaign ID (optional — filters to a single campaign)")

	insightsCmd.AddCommand(
		insightsCampaignsCmd, insightsAdGroupsCmd,
		insightsKeywordsCmd, insightsSearchTermsCmd, insightsAdsCmd,
	)
	rootCmd.AddCommand(insightsCmd)
}
