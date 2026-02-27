package api

import "encoding/json"

// SearchResponse is the response from googleAds:search.
type SearchResponse struct {
	Results       []json.RawMessage `json:"results"`
	NextPageToken string            `json:"nextPageToken,omitempty"`
}

// GoogleAdsError represents an API error response.
type GoogleAdsError struct {
	StatusCode int
	Body       string
}

func (e *GoogleAdsError) Error() string {
	return e.Body
}

// AccessibleCustomersResponse is returned by customers:listAccessibleCustomers.
type AccessibleCustomersResponse struct {
	ResourceNames []string `json:"resourceNames"`
}

// CustomerClientRow is a GAQL result row for customer_client queries.
type CustomerClientRow struct {
	CustomerClient CustomerClient `json:"customerClient"`
}

// CustomerClient represents a client account under an MCC.
type CustomerClient struct {
	ID              string `json:"id"`
	DescriptiveName string `json:"descriptiveName"`
	CurrencyCode    string `json:"currencyCode"`
	TimeZone        string `json:"timeZone"`
	Manager         bool   `json:"manager"`
	Level           int32  `json:"level"`
	Hidden          bool   `json:"hidden"`
	TestAccount     bool   `json:"testAccount"`
}

// CampaignRow is a GAQL result row for campaign queries.
type CampaignRow struct {
	Campaign       Campaign       `json:"campaign"`
	CampaignBudget CampaignBudget `json:"campaignBudget"`
	Metrics        Metrics        `json:"metrics"`
}

// Campaign represents a Google Ads campaign.
type Campaign struct {
	ResourceName           string `json:"resourceName"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Status                 string `json:"status"`
	AdvertisingChannelType string `json:"advertisingChannelType"`
	BiddingStrategyType    string `json:"biddingStrategyType"`
	StartDate              string `json:"startDate"`
	EndDate                string `json:"endDate"`
	CampaignBudget         string `json:"campaignBudget"` // resource name string
}

// CampaignBudget represents a campaign budget.
type CampaignBudget struct {
	ResourceName string `json:"resourceName"`
	ID           string `json:"id"`
	AmountMicros string `json:"amountMicros"`
}

// AdGroupRow is a GAQL result row for ad_group queries.
type AdGroupRow struct {
	AdGroup  AdGroup  `json:"adGroup"`
	Campaign Campaign `json:"campaign"`
	Metrics  Metrics  `json:"metrics"`
}

// AdGroup represents a Google Ads ad group.
type AdGroup struct {
	ResourceName string `json:"resourceName"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	CpcBidMicros string `json:"cpcBidMicros"`
	Campaign     string `json:"campaign"` // resource name string
}

// KeywordRow is a GAQL result row for keyword queries.
type KeywordRow struct {
	AdGroupCriterion AdGroupCriterion `json:"adGroupCriterion"`
	AdGroup          AdGroup          `json:"adGroup"`
	Campaign         Campaign         `json:"campaign"`
	Metrics          Metrics          `json:"metrics"`
}

// AdGroupCriterion represents a keyword criterion.
type AdGroupCriterion struct {
	ResourceName string `json:"resourceName"`
	CriterionID  string `json:"criterionId"`
	Status       string `json:"status"`
	Negative     bool   `json:"negative"`
	Keyword      struct {
		Text      string `json:"text"`
		MatchType string `json:"matchType"`
	} `json:"keyword"`
	QualityInfo struct {
		QualityScore int `json:"qualityScore"`
	} `json:"qualityInfo"`
	CpcBidMicros string `json:"cpcBidMicros"`
}

// AdRow is a GAQL result row for ad_group_ad queries.
type AdRow struct {
	AdGroupAd AdGroupAd `json:"adGroupAd"`
	AdGroup   AdGroup   `json:"adGroup"`
	Campaign  Campaign  `json:"campaign"`
	Metrics   Metrics   `json:"metrics"`
}

// AdGroupAd represents an ad within an ad group.
type AdGroupAd struct {
	ResourceName string `json:"resourceName"`
	Status       string `json:"status"`
	Ad           Ad     `json:"ad"`
}

// Ad represents the ad itself.
type Ad struct {
	ID                 string   `json:"id"`
	Type               string   `json:"type"`
	FinalUrls          []string `json:"finalUrls"`
	ResponsiveSearchAd struct {
		Headlines    []AdTextAsset `json:"headlines"`
		Descriptions []AdTextAsset `json:"descriptions"`
	} `json:"responsiveSearchAd"`
}

// AdTextAsset is a headline or description in a responsive search ad.
type AdTextAsset struct {
	Text             string `json:"text"`
	PinnedField      string `json:"pinnedField,omitempty"`
	AssetPerformance string `json:"assetPerformanceLabel,omitempty"`
}

// InsightsCampaignRow is a GAQL result row for campaign insights.
type InsightsCampaignRow struct {
	Campaign Campaign `json:"campaign"`
	Metrics  Metrics  `json:"metrics"`
}

// InsightsAdGroupRow is a GAQL result row for ad group insights.
type InsightsAdGroupRow struct {
	AdGroup  AdGroup  `json:"adGroup"`
	Campaign Campaign `json:"campaign"`
	Metrics  Metrics  `json:"metrics"`
}

// InsightsKeywordRow is a GAQL result row for keyword insights.
type InsightsKeywordRow struct {
	AdGroupCriterion AdGroupCriterion `json:"adGroupCriterion"`
	AdGroup          AdGroup          `json:"adGroup"`
	Campaign         Campaign         `json:"campaign"`
	Metrics          Metrics          `json:"metrics"`
}

// SearchTermRow is a GAQL result row for search term reports.
type SearchTermRow struct {
	SearchTermView SearchTermView `json:"searchTermView"`
	AdGroup        AdGroup        `json:"adGroup"`
	Campaign       Campaign       `json:"campaign"`
	Metrics        Metrics        `json:"metrics"`
}

// SearchTermView represents a search term view entry.
type SearchTermView struct {
	ResourceName string `json:"resourceName"`
	SearchTerm   string `json:"searchTerm"`
	Status       string `json:"status"`
}

// Metrics holds performance metrics returned by GAQL.
// Integer fields are returned as strings by the Google Ads API.
type Metrics struct {
	Impressions      string  `json:"impressions"`
	Clicks           string  `json:"clicks"`
	CostMicros       string  `json:"costMicros"`
	Ctr              float64 `json:"ctr"`
	AverageCpc       string  `json:"averageCpc"`
	Conversions      float64 `json:"conversions"`
	ConversionsValue float64 `json:"conversionsValue"`
}

// MutateResponse is the response from mutate endpoints.
type MutateResponse struct {
	Results []struct {
		ResourceName string `json:"resourceName"`
	} `json:"results"`
}
