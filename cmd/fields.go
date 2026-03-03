package cmd

import (
	"fmt"
	"strings"

	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/output"
)

// ---- Field ID constants ----
// Use these as comma-separated values in --fields flag.

const (
	// Campaign dimension
	FidCampaignID     = "campaign_id"
	FidCampaignName   = "campaign_name"
	FidCampaignStatus = "campaign_status"
	FidCampaignType   = "campaign_type"

	// Ad group dimension
	FidAdGroupID     = "adgroup_id"
	FidAdGroupName   = "adgroup_name"
	FidAdGroupStatus = "adgroup_status"

	// Keyword dimension
	FidKeywordText      = "keyword_text"
	FidKeywordMatchType = "keyword_match"
	FidKeywordStatus    = "keyword_status"
	FidQualityScore     = "quality_score"

	// Search term dimension
	FidSearchTerm       = "search_term"
	FidSearchTermStatus = "st_status"

	// Ad dimension
	FidAdID            = "ad_id"
	FidAdName          = "ad_name"
	FidAdStatus        = "ad_status"
	FidAdType          = "ad_type"
	FidFinalURL        = "final_url"
	FidFinalMobileURL  = "final_mobile_url"
	FidTrackingURL     = "tracking_url"
	FidFinalURLSuffix  = "url_suffix"
	FidDisplayURL      = "display_url"
	FidPath1           = "path1"
	FidPath2           = "path2"

	// RSA Headlines (responsive search ad)
	FidHeadline1    = "headline1"
	FidHeadline2    = "headline2"
	FidHeadline3    = "headline3"
	FidHeadline4    = "headline4"
	FidHeadline5    = "headline5"
	FidHeadline6    = "headline6"
	FidHeadline7    = "headline7"
	FidHeadline8    = "headline8"
	FidHeadline9    = "headline9"
	FidHeadline10   = "headline10"
	FidHeadline11   = "headline11"
	FidHeadline12   = "headline12"
	FidHeadline13   = "headline13"
	FidHeadline14   = "headline14"
	FidHeadline15   = "headline15"
	FidHeadline1Pos  = "headline1_pos"
	FidHeadline2Pos  = "headline2_pos"
	FidHeadline3Pos  = "headline3_pos"
	FidHeadline4Pos  = "headline4_pos"
	FidHeadline5Pos  = "headline5_pos"
	FidHeadline6Pos  = "headline6_pos"
	FidHeadline7Pos  = "headline7_pos"
	FidHeadline8Pos  = "headline8_pos"
	FidHeadline9Pos  = "headline9_pos"
	FidHeadline10Pos = "headline10_pos"
	FidHeadline11Pos = "headline11_pos"
	FidHeadline12Pos = "headline12_pos"
	FidHeadline13Pos = "headline13_pos"
	FidHeadline14Pos = "headline14_pos"
	FidHeadline15Pos = "headline15_pos"

	// RSA Descriptions
	FidDesc1    = "desc1"
	FidDesc2    = "desc2"
	FidDesc3    = "desc3"
	FidDesc4    = "desc4"
	FidDesc1Pos = "desc1_pos"
	FidDesc2Pos = "desc2_pos"
	FidDesc3Pos = "desc3_pos"
	FidDesc4Pos = "desc4_pos"

	// ETA (Expanded Text Ad, legacy)
	FidETAHeadline1  = "eta_headline1"
	FidETAHeadline2  = "eta_headline2"
	FidETAHeadline3  = "eta_headline3"
	FidETADesc1      = "eta_desc1"
	FidETADesc2      = "eta_desc2"

	// Metrics
	FidImpressions     = "impressions"
	FidClicks          = "clicks"
	FidCost            = "cost"
	FidCTR             = "ctr"
	FidCPC             = "cpc"
	FidConversions     = "conversions"
	FidConvValue       = "conv_value"
	FidROAS            = "roas"
	FidAbsTopImpPct    = "abs_top_imp_pct"
	FidTopImpPct       = "top_imp_pct"
	FidViewThroughConv = "view_through_conv"
	FidCostPerConv     = "cost_per_conv"
	FidConvRate        = "conv_rate"
	FidSearchImpShare  = "search_imp_share"
)

// FieldGAQL maps field IDs to their Google Ads Query Language (GAQL) field names.
// Fields that are computed (roas) or multi-value (headlines) have a representative value.
var FieldGAQL = map[string]string{
	FidCampaignID:     "campaign.id",
	FidCampaignName:   "campaign.name",
	FidCampaignStatus: "campaign.status",
	FidCampaignType:   "campaign.advertising_channel_type",

	FidAdGroupID:     "ad_group.id",
	FidAdGroupName:   "ad_group.name",
	FidAdGroupStatus: "ad_group.status",

	FidKeywordText:      "ad_group_criterion.keyword.text",
	FidKeywordMatchType: "ad_group_criterion.keyword.match_type",
	FidKeywordStatus:    "ad_group_criterion.status",
	FidQualityScore:     "ad_group_criterion.quality_info.quality_score",

	FidSearchTerm:       "search_term_view.search_term",
	FidSearchTermStatus: "search_term_view.status",

	FidAdID:           "ad_group_ad.ad.id",
	FidAdName:         "ad_group_ad.ad.name",
	FidAdStatus:       "ad_group_ad.status",
	FidAdType:         "ad_group_ad.ad.type",
	FidFinalURL:       "ad_group_ad.ad.final_urls",
	FidFinalMobileURL: "ad_group_ad.ad.final_mobile_urls",
	FidTrackingURL:    "ad_group_ad.ad.tracking_url_template",
	FidFinalURLSuffix: "ad_group_ad.ad.final_url_suffix",
	FidDisplayURL:     "ad_group_ad.ad.display_url",
	FidPath1:          "ad_group_ad.ad.responsive_search_ad.path1",
	FidPath2:          "ad_group_ad.ad.responsive_search_ad.path2",

	// RSA headlines/descriptions all share the same GAQL parent field
	FidHeadline1:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline2:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline3:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline4:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline5:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline6:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline7:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline8:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline9:    "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline10:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline11:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline12:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline13:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline14:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline15:   "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline1Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline2Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline3Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline4Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline5Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline6Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline7Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline8Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline9Pos:  "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline10Pos: "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline11Pos: "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline12Pos: "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline13Pos: "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline14Pos: "ad_group_ad.ad.responsive_search_ad.headlines",
	FidHeadline15Pos: "ad_group_ad.ad.responsive_search_ad.headlines",

	FidDesc1:    "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc2:    "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc3:    "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc4:    "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc1Pos: "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc2Pos: "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc3Pos: "ad_group_ad.ad.responsive_search_ad.descriptions",
	FidDesc4Pos: "ad_group_ad.ad.responsive_search_ad.descriptions",

	FidETAHeadline1: "ad_group_ad.ad.expanded_text_ad.headline_part1",
	FidETAHeadline2: "ad_group_ad.ad.expanded_text_ad.headline_part2",
	FidETAHeadline3: "ad_group_ad.ad.expanded_text_ad.headline_part3",
	FidETADesc1:     "ad_group_ad.ad.expanded_text_ad.description",
	FidETADesc2:     "ad_group_ad.ad.expanded_text_ad.description2",

	FidImpressions:     "metrics.impressions",
	FidClicks:          "metrics.clicks",
	FidCost:            "metrics.cost_micros",
	FidCTR:             "metrics.ctr",
	FidCPC:             "metrics.average_cpc",
	FidConversions:     "metrics.conversions",
	FidConvValue:       "metrics.conversions_value",
	FidROAS:            "", // computed: conversions_value / cost
	FidAbsTopImpPct:    "metrics.absolute_top_impression_percentage",
	FidTopImpPct:       "metrics.top_impression_percentage",
	FidViewThroughConv: "metrics.view_through_conversions",
	FidCostPerConv:     "metrics.cost_per_conversion",
	FidConvRate:        "metrics.conversions_from_interactions_rate",
	FidSearchImpShare:  "metrics.search_impression_share",
}

// parseFieldList splits a comma-separated fields string into trimmed IDs.
func parseFieldList(fields string) []string {
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ===== CAMPAIGN COLUMNS =====

// CampaignCol is a column specification for campaign insights.
type CampaignCol struct {
	ID     string
	Label  string
	Format func(*api.InsightsCampaignRow) string
}

var campaignColDefs = []CampaignCol{
	{FidCampaignID, "CAMPAIGN ID", func(r *api.InsightsCampaignRow) string { return r.Campaign.ID }},
	{FidCampaignName, "CAMPAIGN", func(r *api.InsightsCampaignRow) string {
		return output.Truncate(r.Campaign.Name, 30)
	}},
	{FidCampaignStatus, "STATUS", func(r *api.InsightsCampaignRow) string {
		return strings.ToLower(r.Campaign.Status)
	}},
	{FidCampaignType, "TYPE", func(r *api.InsightsCampaignRow) string {
		return strings.ToLower(r.Campaign.AdvertisingChannelType)
	}},
	{FidImpressions, "IMPR", func(r *api.InsightsCampaignRow) string {
		return api.FormatMetricInt(r.Metrics.Impressions)
	}},
	{FidClicks, "CLICKS", func(r *api.InsightsCampaignRow) string {
		return api.FormatMetricInt(r.Metrics.Clicks)
	}},
	{FidCost, "COST", func(r *api.InsightsCampaignRow) string {
		return api.MicrosToCurrency(r.Metrics.CostMicros)
	}},
	{FidCTR, "CTR", func(r *api.InsightsCampaignRow) string {
		return api.FormatCTR(r.Metrics.Ctr)
	}},
	{FidCPC, "CPC", func(r *api.InsightsCampaignRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.AverageCpc)
	}},
	{FidConversions, "CONV", func(r *api.InsightsCampaignRow) string {
		return fmt.Sprintf("%.1f", r.Metrics.Conversions)
	}},
	{FidConvValue, "CONV VALUE", func(r *api.InsightsCampaignRow) string {
		return fmt.Sprintf("%.2f", r.Metrics.ConversionsValue)
	}},
	{FidROAS, "ROAS", func(r *api.InsightsCampaignRow) string {
		return api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros)
	}},
	{FidAbsTopImpPct, "ABS TOP%", func(r *api.InsightsCampaignRow) string {
		return api.FormatPct(r.Metrics.AbsoluteTopImpressionPercentage)
	}},
	{FidTopImpPct, "TOP%", func(r *api.InsightsCampaignRow) string {
		return api.FormatPct(r.Metrics.TopImpressionPercentage)
	}},
	{FidViewThroughConv, "VIEW CONV", func(r *api.InsightsCampaignRow) string {
		if r.Metrics.ViewThroughConversions == "" {
			return "0"
		}
		return r.Metrics.ViewThroughConversions
	}},
	{FidCostPerConv, "COST/CONV", func(r *api.InsightsCampaignRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.CostPerConversion)
	}},
	{FidConvRate, "CONV%", func(r *api.InsightsCampaignRow) string {
		return api.FormatPct(r.Metrics.ConversionsFromInteractionsRate)
	}},
	{FidSearchImpShare, "IMP SHARE", func(r *api.InsightsCampaignRow) string {
		return api.FormatPct(r.Metrics.SearchImpressionShare)
	}},
}

var campaignColByID = func() map[string]CampaignCol {
	m := make(map[string]CampaignCol, len(campaignColDefs))
	for _, c := range campaignColDefs {
		m[c.ID] = c
	}
	return m
}()

var campaignPresets = map[string][]string{
	"default": {
		FidCampaignName, FidCampaignStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC, FidConversions, FidROAS,
	},
	"performance": {
		FidCampaignName, FidCampaignStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidROAS, FidAbsTopImpPct, FidTopImpPct, FidConvRate, FidCostPerConv,
	},
	"conversions": {
		FidCampaignName, FidCampaignStatus,
		FidConversions, FidConvValue, FidViewThroughConv, FidConvRate, FidCostPerConv, FidROAS,
	},
	"full": {
		FidCampaignID, FidCampaignName, FidCampaignStatus, FidCampaignType,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidConvValue, FidROAS,
		FidAbsTopImpPct, FidTopImpPct, FidViewThroughConv, FidConvRate, FidCostPerConv, FidSearchImpShare,
	},
}

// resolveCampaignCols returns ordered column specs based on preset and --fields override.
func resolveCampaignCols(preset, fields string) []CampaignCol {
	ids := campaignPresets["default"]
	if p, ok := campaignPresets[preset]; ok {
		ids = p
	}
	if fields != "" {
		ids = parseFieldList(fields)
	}
	var cols []CampaignCol
	for _, id := range ids {
		if c, ok := campaignColByID[id]; ok {
			cols = append(cols, c)
		}
	}
	return cols
}

// ===== AD GROUP COLUMNS =====

// AdGroupCol is a column specification for ad group insights.
type AdGroupCol struct {
	ID     string
	Label  string
	Format func(*api.InsightsAdGroupRow) string
}

var adGroupColDefs = []AdGroupCol{
	{FidCampaignName, "CAMPAIGN", func(r *api.InsightsAdGroupRow) string {
		return output.Truncate(r.Campaign.Name, 24)
	}},
	{FidAdGroupID, "ADGROUP ID", func(r *api.InsightsAdGroupRow) string { return r.AdGroup.ID }},
	{FidAdGroupName, "ADGROUP", func(r *api.InsightsAdGroupRow) string {
		return output.Truncate(r.AdGroup.Name, 30)
	}},
	{FidAdGroupStatus, "STATUS", func(r *api.InsightsAdGroupRow) string {
		return strings.ToLower(r.AdGroup.Status)
	}},
	{FidImpressions, "IMPR", func(r *api.InsightsAdGroupRow) string {
		return api.FormatMetricInt(r.Metrics.Impressions)
	}},
	{FidClicks, "CLICKS", func(r *api.InsightsAdGroupRow) string {
		return api.FormatMetricInt(r.Metrics.Clicks)
	}},
	{FidCost, "COST", func(r *api.InsightsAdGroupRow) string {
		return api.MicrosToCurrency(r.Metrics.CostMicros)
	}},
	{FidCTR, "CTR", func(r *api.InsightsAdGroupRow) string {
		return api.FormatCTR(r.Metrics.Ctr)
	}},
	{FidCPC, "CPC", func(r *api.InsightsAdGroupRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.AverageCpc)
	}},
	{FidConversions, "CONV", func(r *api.InsightsAdGroupRow) string {
		return fmt.Sprintf("%.1f", r.Metrics.Conversions)
	}},
	{FidConvValue, "CONV VALUE", func(r *api.InsightsAdGroupRow) string {
		return fmt.Sprintf("%.2f", r.Metrics.ConversionsValue)
	}},
	{FidROAS, "ROAS", func(r *api.InsightsAdGroupRow) string {
		return api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros)
	}},
	{FidAbsTopImpPct, "ABS TOP%", func(r *api.InsightsAdGroupRow) string {
		return api.FormatPct(r.Metrics.AbsoluteTopImpressionPercentage)
	}},
	{FidTopImpPct, "TOP%", func(r *api.InsightsAdGroupRow) string {
		return api.FormatPct(r.Metrics.TopImpressionPercentage)
	}},
	{FidViewThroughConv, "VIEW CONV", func(r *api.InsightsAdGroupRow) string {
		if r.Metrics.ViewThroughConversions == "" {
			return "0"
		}
		return r.Metrics.ViewThroughConversions
	}},
	{FidCostPerConv, "COST/CONV", func(r *api.InsightsAdGroupRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.CostPerConversion)
	}},
	{FidConvRate, "CONV%", func(r *api.InsightsAdGroupRow) string {
		return api.FormatPct(r.Metrics.ConversionsFromInteractionsRate)
	}},
	{FidSearchImpShare, "IMP SHARE", func(r *api.InsightsAdGroupRow) string {
		return api.FormatPct(r.Metrics.SearchImpressionShare)
	}},
}

var adGroupColByID = func() map[string]AdGroupCol {
	m := make(map[string]AdGroupCol, len(adGroupColDefs))
	for _, c := range adGroupColDefs {
		m[c.ID] = c
	}
	return m
}()

var adGroupPresets = map[string][]string{
	"default": {
		FidAdGroupName, FidAdGroupStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC, FidConversions, FidROAS,
	},
	"performance": {
		FidCampaignName, FidAdGroupName, FidAdGroupStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidROAS, FidConvRate, FidCostPerConv, FidAbsTopImpPct, FidTopImpPct,
	},
	"conversions": {
		FidCampaignName, FidAdGroupName, FidAdGroupStatus,
		FidConversions, FidConvValue, FidViewThroughConv, FidConvRate, FidCostPerConv, FidROAS,
	},
	"full": {
		FidCampaignName, FidAdGroupID, FidAdGroupName, FidAdGroupStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidConvValue, FidROAS,
		FidAbsTopImpPct, FidTopImpPct, FidViewThroughConv, FidConvRate, FidCostPerConv, FidSearchImpShare,
	},
}

func resolveAdGroupCols(preset, fields string) []AdGroupCol {
	ids := adGroupPresets["default"]
	if p, ok := adGroupPresets[preset]; ok {
		ids = p
	}
	if fields != "" {
		ids = parseFieldList(fields)
	}
	var cols []AdGroupCol
	for _, id := range ids {
		if c, ok := adGroupColByID[id]; ok {
			cols = append(cols, c)
		}
	}
	return cols
}

// ===== KEYWORD COLUMNS =====

// KeywordCol is a column specification for keyword insights.
type KeywordCol struct {
	ID     string
	Label  string
	Format func(*api.InsightsKeywordRow) string
}

var keywordColDefs = []KeywordCol{
	{FidCampaignName, "CAMPAIGN", func(r *api.InsightsKeywordRow) string {
		return output.Truncate(r.Campaign.Name, 20)
	}},
	{FidAdGroupName, "ADGROUP", func(r *api.InsightsKeywordRow) string {
		return output.Truncate(r.AdGroup.Name, 20)
	}},
	{FidKeywordText, "KEYWORD", func(r *api.InsightsKeywordRow) string {
		return output.Truncate(r.AdGroupCriterion.Keyword.Text, 30)
	}},
	{FidKeywordMatchType, "MATCH", func(r *api.InsightsKeywordRow) string {
		return strings.ToLower(r.AdGroupCriterion.Keyword.MatchType)
	}},
	{FidKeywordStatus, "KW STATUS", func(r *api.InsightsKeywordRow) string {
		return strings.ToLower(r.AdGroupCriterion.Status)
	}},
	{FidQualityScore, "QS", func(r *api.InsightsKeywordRow) string {
		return fmt.Sprintf("%d", r.AdGroupCriterion.QualityInfo.QualityScore)
	}},
	{FidImpressions, "IMPR", func(r *api.InsightsKeywordRow) string {
		return api.FormatMetricInt(r.Metrics.Impressions)
	}},
	{FidClicks, "CLICKS", func(r *api.InsightsKeywordRow) string {
		return api.FormatMetricInt(r.Metrics.Clicks)
	}},
	{FidCost, "COST", func(r *api.InsightsKeywordRow) string {
		return api.MicrosToCurrency(r.Metrics.CostMicros)
	}},
	{FidCTR, "CTR", func(r *api.InsightsKeywordRow) string {
		return api.FormatCTR(r.Metrics.Ctr)
	}},
	{FidCPC, "CPC", func(r *api.InsightsKeywordRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.AverageCpc)
	}},
	{FidConversions, "CONV", func(r *api.InsightsKeywordRow) string {
		return fmt.Sprintf("%.1f", r.Metrics.Conversions)
	}},
	{FidConvValue, "CONV VALUE", func(r *api.InsightsKeywordRow) string {
		return fmt.Sprintf("%.2f", r.Metrics.ConversionsValue)
	}},
	{FidROAS, "ROAS", func(r *api.InsightsKeywordRow) string {
		return api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros)
	}},
	{FidAbsTopImpPct, "ABS TOP%", func(r *api.InsightsKeywordRow) string {
		return api.FormatPct(r.Metrics.AbsoluteTopImpressionPercentage)
	}},
	{FidTopImpPct, "TOP%", func(r *api.InsightsKeywordRow) string {
		return api.FormatPct(r.Metrics.TopImpressionPercentage)
	}},
	{FidViewThroughConv, "VIEW CONV", func(r *api.InsightsKeywordRow) string {
		if r.Metrics.ViewThroughConversions == "" {
			return "0"
		}
		return r.Metrics.ViewThroughConversions
	}},
	{FidCostPerConv, "COST/CONV", func(r *api.InsightsKeywordRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.CostPerConversion)
	}},
	{FidConvRate, "CONV%", func(r *api.InsightsKeywordRow) string {
		return api.FormatPct(r.Metrics.ConversionsFromInteractionsRate)
	}},
	{FidSearchImpShare, "IMP SHARE", func(r *api.InsightsKeywordRow) string {
		return api.FormatPct(r.Metrics.SearchImpressionShare)
	}},
}

var keywordColByID = func() map[string]KeywordCol {
	m := make(map[string]KeywordCol, len(keywordColDefs))
	for _, c := range keywordColDefs {
		m[c.ID] = c
	}
	return m
}()

var keywordPresets = map[string][]string{
	"default": {
		FidKeywordText, FidKeywordMatchType, FidKeywordStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC, FidConversions, FidQualityScore,
	},
	"performance": {
		FidKeywordText, FidKeywordMatchType, FidKeywordStatus,
		FidCampaignName, FidAdGroupName,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidROAS, FidConvRate, FidCostPerConv, FidQualityScore,
	},
	"conversions": {
		FidKeywordText, FidKeywordMatchType, FidKeywordStatus,
		FidConversions, FidConvValue, FidConvRate, FidCostPerConv, FidROAS,
	},
	"full": {
		FidKeywordText, FidKeywordMatchType, FidKeywordStatus, FidQualityScore,
		FidCampaignName, FidAdGroupName,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidConvValue, FidROAS,
		FidAbsTopImpPct, FidTopImpPct, FidViewThroughConv, FidConvRate, FidCostPerConv, FidSearchImpShare,
	},
}

func resolveKeywordCols(preset, fields string) []KeywordCol {
	ids := keywordPresets["default"]
	if p, ok := keywordPresets[preset]; ok {
		ids = p
	}
	if fields != "" {
		ids = parseFieldList(fields)
	}
	var cols []KeywordCol
	for _, id := range ids {
		if c, ok := keywordColByID[id]; ok {
			cols = append(cols, c)
		}
	}
	return cols
}

// ===== SEARCH TERM COLUMNS =====

// SearchTermCol is a column specification for search term insights.
type SearchTermCol struct {
	ID     string
	Label  string
	Format func(*api.SearchTermRow) string
}

var searchTermColDefs = []SearchTermCol{
	{FidSearchTerm, "SEARCH TERM", func(r *api.SearchTermRow) string {
		return output.Truncate(r.SearchTermView.SearchTerm, 40)
	}},
	{FidSearchTermStatus, "STATUS", func(r *api.SearchTermRow) string {
		return strings.ToLower(r.SearchTermView.Status)
	}},
	{FidCampaignName, "CAMPAIGN", func(r *api.SearchTermRow) string {
		return output.Truncate(r.Campaign.Name, 20)
	}},
	{FidAdGroupName, "ADGROUP", func(r *api.SearchTermRow) string {
		return output.Truncate(r.AdGroup.Name, 24)
	}},
	{FidImpressions, "IMPR", func(r *api.SearchTermRow) string {
		return api.FormatMetricInt(r.Metrics.Impressions)
	}},
	{FidClicks, "CLICKS", func(r *api.SearchTermRow) string {
		return api.FormatMetricInt(r.Metrics.Clicks)
	}},
	{FidCost, "COST", func(r *api.SearchTermRow) string {
		return api.MicrosToCurrency(r.Metrics.CostMicros)
	}},
	{FidCTR, "CTR", func(r *api.SearchTermRow) string {
		return api.FormatCTR(r.Metrics.Ctr)
	}},
	{FidCPC, "CPC", func(r *api.SearchTermRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.AverageCpc)
	}},
	{FidConversions, "CONV", func(r *api.SearchTermRow) string {
		return fmt.Sprintf("%.1f", r.Metrics.Conversions)
	}},
	{FidConvValue, "CONV VALUE", func(r *api.SearchTermRow) string {
		return fmt.Sprintf("%.2f", r.Metrics.ConversionsValue)
	}},
	{FidROAS, "ROAS", func(r *api.SearchTermRow) string {
		return api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros)
	}},
	{FidViewThroughConv, "VIEW CONV", func(r *api.SearchTermRow) string {
		if r.Metrics.ViewThroughConversions == "" {
			return "0"
		}
		return r.Metrics.ViewThroughConversions
	}},
	{FidCostPerConv, "COST/CONV", func(r *api.SearchTermRow) string {
		return api.MicrosFloatToCurrency(r.Metrics.CostPerConversion)
	}},
	{FidConvRate, "CONV%", func(r *api.SearchTermRow) string {
		return api.FormatPct(r.Metrics.ConversionsFromInteractionsRate)
	}},
}

var searchTermColByID = func() map[string]SearchTermCol {
	m := make(map[string]SearchTermCol, len(searchTermColDefs))
	for _, c := range searchTermColDefs {
		m[c.ID] = c
	}
	return m
}()

var searchTermPresets = map[string][]string{
	"default": {
		FidSearchTerm, FidSearchTermStatus, FidAdGroupName,
		FidImpressions, FidClicks, FidCost, FidCTR, FidConversions,
	},
	"performance": {
		FidSearchTerm, FidSearchTermStatus, FidCampaignName, FidAdGroupName,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidROAS, FidConvRate, FidCostPerConv,
	},
	"conversions": {
		FidSearchTerm, FidSearchTermStatus, FidCampaignName, FidAdGroupName,
		FidConversions, FidConvValue, FidConvRate, FidCostPerConv, FidROAS,
	},
	"full": {
		FidSearchTerm, FidSearchTermStatus, FidCampaignName, FidAdGroupName,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidConvValue, FidROAS, FidViewThroughConv, FidConvRate, FidCostPerConv,
	},
}

func resolveSearchTermCols(preset, fields string) []SearchTermCol {
	ids := searchTermPresets["default"]
	if p, ok := searchTermPresets[preset]; ok {
		ids = p
	}
	if fields != "" {
		ids = parseFieldList(fields)
	}
	var cols []SearchTermCol
	for _, id := range ids {
		if c, ok := searchTermColByID[id]; ok {
			cols = append(cols, c)
		}
	}
	return cols
}

// ===== AD COLUMNS =====

// AdCol is a column specification for ad insights.
type AdCol struct {
	ID     string
	Label  string
	Format func(*api.InsightsAdRow) string
}

// rsaHeadlineText returns the text of the nth RSA headline (n is 1-based).
func rsaHeadlineText(r *api.InsightsAdRow, n int) string {
	hl := r.AdGroupAd.Ad.ResponsiveSearchAd.Headlines
	if n-1 < len(hl) {
		return hl[n-1].Text
	}
	return ""
}

// rsaHeadlinePos returns the pinned position of the nth RSA headline (n is 1-based).
func rsaHeadlinePos(r *api.InsightsAdRow, n int) string {
	hl := r.AdGroupAd.Ad.ResponsiveSearchAd.Headlines
	if n-1 < len(hl) {
		return hl[n-1].PinnedField
	}
	return ""
}

// rsaDescText returns the text of the nth RSA description (n is 1-based).
func rsaDescText(r *api.InsightsAdRow, n int) string {
	ds := r.AdGroupAd.Ad.ResponsiveSearchAd.Descriptions
	if n-1 < len(ds) {
		return ds[n-1].Text
	}
	return ""
}

// rsaDescPos returns the pinned position of the nth RSA description (n is 1-based).
func rsaDescPos(r *api.InsightsAdRow, n int) string {
	ds := r.AdGroupAd.Ad.ResponsiveSearchAd.Descriptions
	if n-1 < len(ds) {
		return ds[n-1].PinnedField
	}
	return ""
}

// makeHeadlineCol creates an AdCol for RSA headline at position n (1-based).
func makeHeadlineCol(n int) AdCol {
	id := fmt.Sprintf("headline%d", n)
	label := fmt.Sprintf("HEADLINE%d", n)
	return AdCol{id, label, func(r *api.InsightsAdRow) string { return rsaHeadlineText(r, n) }}
}

// makeHeadlinePosCol creates an AdCol for RSA headline position at n (1-based).
func makeHeadlinePosCol(n int) AdCol {
	id := fmt.Sprintf("headline%d_pos", n)
	label := fmt.Sprintf("HL%d POS", n)
	return AdCol{id, label, func(r *api.InsightsAdRow) string { return rsaHeadlinePos(r, n) }}
}

// makeDescCol creates an AdCol for RSA description at position n (1-based).
func makeDescCol(n int) AdCol {
	id := fmt.Sprintf("desc%d", n)
	label := fmt.Sprintf("DESC%d", n)
	return AdCol{id, label, func(r *api.InsightsAdRow) string { return rsaDescText(r, n) }}
}

// makeDescPosCol creates an AdCol for RSA description position at n (1-based).
func makeDescPosCol(n int) AdCol {
	id := fmt.Sprintf("desc%d_pos", n)
	label := fmt.Sprintf("DESC%d POS", n)
	return AdCol{id, label, func(r *api.InsightsAdRow) string { return rsaDescPos(r, n) }}
}

var adColDefs func() []AdCol = func() []AdCol {
	cols := []AdCol{
		{FidCampaignName, "CAMPAIGN", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.Campaign.Name, 20)
		}},
		{FidAdGroupName, "ADGROUP", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroup.Name, 20)
		}},
		{FidAdID, "AD ID", func(r *api.InsightsAdRow) string { return r.AdGroupAd.Ad.ID }},
		{FidAdName, "AD NAME", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.Name, 30)
		}},
		{FidAdStatus, "STATUS", func(r *api.InsightsAdRow) string {
			return strings.ToLower(r.AdGroupAd.Status)
		}},
		{FidAdType, "AD TYPE", func(r *api.InsightsAdRow) string {
			return strings.ToLower(r.AdGroupAd.Ad.Type)
		}},
		{FidFinalURL, "FINAL URL", func(r *api.InsightsAdRow) string {
			if len(r.AdGroupAd.Ad.FinalUrls) > 0 {
				return output.Truncate(r.AdGroupAd.Ad.FinalUrls[0], 50)
			}
			return ""
		}},
		{FidFinalMobileURL, "MOBILE URL", func(r *api.InsightsAdRow) string {
			if len(r.AdGroupAd.Ad.FinalMobileUrls) > 0 {
				return output.Truncate(r.AdGroupAd.Ad.FinalMobileUrls[0], 40)
			}
			return ""
		}},
		{FidTrackingURL, "TRACKING URL", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.TrackingUrlTemplate, 40)
		}},
		{FidFinalURLSuffix, "URL SUFFIX", func(r *api.InsightsAdRow) string {
			return r.AdGroupAd.Ad.FinalUrlSuffix
		}},
		{FidDisplayURL, "DISPLAY URL", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.DisplayUrl, 30)
		}},
		{FidPath1, "PATH1", func(r *api.InsightsAdRow) string {
			return r.AdGroupAd.Ad.ResponsiveSearchAd.Path1
		}},
		{FidPath2, "PATH2", func(r *api.InsightsAdRow) string {
			return r.AdGroupAd.Ad.ResponsiveSearchAd.Path2
		}},
		// ETA legacy fields
		{FidETAHeadline1, "ETA HL1", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.ExpandedTextAd.HeadlinePart1, 30)
		}},
		{FidETAHeadline2, "ETA HL2", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.ExpandedTextAd.HeadlinePart2, 30)
		}},
		{FidETAHeadline3, "ETA HL3", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.ExpandedTextAd.HeadlinePart3, 30)
		}},
		{FidETADesc1, "ETA DESC1", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.ExpandedTextAd.Description, 40)
		}},
		{FidETADesc2, "ETA DESC2", func(r *api.InsightsAdRow) string {
			return output.Truncate(r.AdGroupAd.Ad.ExpandedTextAd.Description2, 40)
		}},
		// Metrics
		{FidImpressions, "IMPR", func(r *api.InsightsAdRow) string {
			return api.FormatMetricInt(r.Metrics.Impressions)
		}},
		{FidClicks, "CLICKS", func(r *api.InsightsAdRow) string {
			return api.FormatMetricInt(r.Metrics.Clicks)
		}},
		{FidCost, "COST", func(r *api.InsightsAdRow) string {
			return api.MicrosToCurrency(r.Metrics.CostMicros)
		}},
		{FidCTR, "CTR", func(r *api.InsightsAdRow) string {
			return api.FormatCTR(r.Metrics.Ctr)
		}},
		{FidCPC, "CPC", func(r *api.InsightsAdRow) string {
			return api.MicrosFloatToCurrency(r.Metrics.AverageCpc)
		}},
		{FidConversions, "CONV", func(r *api.InsightsAdRow) string {
			return fmt.Sprintf("%.1f", r.Metrics.Conversions)
		}},
		{FidConvValue, "CONV VALUE", func(r *api.InsightsAdRow) string {
			return fmt.Sprintf("%.2f", r.Metrics.ConversionsValue)
		}},
		{FidROAS, "ROAS", func(r *api.InsightsAdRow) string {
			return api.FormatROAS(r.Metrics.ConversionsValue, r.Metrics.CostMicros)
		}},
		{FidAbsTopImpPct, "ABS TOP%", func(r *api.InsightsAdRow) string {
			return api.FormatPct(r.Metrics.AbsoluteTopImpressionPercentage)
		}},
		{FidTopImpPct, "TOP%", func(r *api.InsightsAdRow) string {
			return api.FormatPct(r.Metrics.TopImpressionPercentage)
		}},
		{FidViewThroughConv, "VIEW CONV", func(r *api.InsightsAdRow) string {
			if r.Metrics.ViewThroughConversions == "" {
				return "0"
			}
			return r.Metrics.ViewThroughConversions
		}},
		{FidCostPerConv, "COST/CONV", func(r *api.InsightsAdRow) string {
			return api.MicrosFloatToCurrency(r.Metrics.CostPerConversion)
		}},
		{FidConvRate, "CONV%", func(r *api.InsightsAdRow) string {
			return api.FormatPct(r.Metrics.ConversionsFromInteractionsRate)
		}},
		{FidSearchImpShare, "IMP SHARE", func(r *api.InsightsAdRow) string {
			return api.FormatPct(r.Metrics.SearchImpressionShare)
		}},
	}
	// RSA headline columns 1-15
	for n := 1; n <= 15; n++ {
		cols = append(cols, makeHeadlineCol(n))
		cols = append(cols, makeHeadlinePosCol(n))
	}
	// RSA description columns 1-4
	for n := 1; n <= 4; n++ {
		cols = append(cols, makeDescCol(n))
		cols = append(cols, makeDescPosCol(n))
	}
	return cols
}

var adColByID = func() map[string]AdCol {
	m := make(map[string]AdCol)
	for _, c := range adColDefs() {
		m[c.ID] = c
	}
	return m
}()

// allHeadlineIDs returns field IDs headline1 through headline15.
func allHeadlineIDs() []string {
	ids := make([]string, 15)
	for i := range ids {
		ids[i] = fmt.Sprintf("headline%d", i+1)
	}
	return ids
}

// allDescIDs returns field IDs desc1 through desc4.
func allDescIDs() []string {
	return []string{FidDesc1, FidDesc2, FidDesc3, FidDesc4}
}

var adPresets = map[string][]string{
	"default": {
		FidCampaignName, FidAdGroupName, FidAdName, FidAdStatus, FidAdType, FidFinalURL,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC, FidConversions, FidROAS,
	},
	"performance": {
		FidCampaignName, FidAdGroupName, FidAdName, FidAdStatus,
		FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
		FidConversions, FidROAS, FidConvRate, FidCostPerConv, FidAbsTopImpPct, FidTopImpPct,
	},
	"creatives": func() []string {
		ids := []string{
			FidCampaignName, FidAdGroupName, FidAdName, FidAdStatus, FidAdType,
			FidFinalURL, FidDisplayURL, FidPath1, FidPath2,
		}
		ids = append(ids, allHeadlineIDs()...)
		ids = append(ids, allDescIDs()...)
		return ids
	}(),
	"full": func() []string {
		ids := []string{
			FidCampaignName, FidAdGroupName,
			FidAdID, FidAdName, FidAdStatus, FidAdType,
			FidFinalURL, FidFinalMobileURL, FidTrackingURL, FidFinalURLSuffix,
			FidDisplayURL, FidPath1, FidPath2,
			FidETAHeadline1, FidETAHeadline2, FidETAHeadline3, FidETADesc1, FidETADesc2,
		}
		ids = append(ids, allHeadlineIDs()...)
		ids = append(ids, allDescIDs()...)
		ids = append(ids,
			FidImpressions, FidClicks, FidCost, FidCTR, FidCPC,
			FidConversions, FidConvValue, FidROAS,
			FidAbsTopImpPct, FidTopImpPct, FidViewThroughConv, FidConvRate, FidCostPerConv, FidSearchImpShare,
		)
		return ids
	}(),
}

func resolveAdCols(preset, fields string) []AdCol {
	ids := adPresets["default"]
	if p, ok := adPresets[preset]; ok {
		ids = p
	}
	if fields != "" {
		ids = parseFieldList(fields)
	}
	var cols []AdCol
	for _, id := range ids {
		if c, ok := adColByID[id]; ok {
			cols = append(cols, c)
		}
	}
	return cols
}

// colHeaders extracts the label from a slice of any ColSpec-like struct.
// Since Go doesn't allow generic methods on non-generic types without type params,
// we provide helpers per type.
func campaignHeaders(cols []CampaignCol) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.Label
	}
	return h
}

func adGroupHeaders(cols []AdGroupCol) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.Label
	}
	return h
}

func keywordHeaders(cols []KeywordCol) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.Label
	}
	return h
}

func searchTermHeaders(cols []SearchTermCol) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.Label
	}
	return h
}

func adHeaders(cols []AdCol) []string {
	h := make([]string, len(cols))
	for i, c := range cols {
		h[i] = c.Label
	}
	return h
}
