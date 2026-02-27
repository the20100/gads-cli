package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const apiBase = "https://googleads.googleapis.com/v19"

// Client wraps an HTTP client with Google Ads API authentication headers.
type Client struct {
	http            *http.Client
	developerToken  string
	loginCustomerID string
}

// New creates a new Client. httpClient should already have OAuth2 transport.
func New(httpClient *http.Client, developerToken, loginCustomerID string) *Client {
	return &Client{
		http:            httpClient,
		developerToken:  developerToken,
		loginCustomerID: loginCustomerID,
	}
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("developer-token", c.developerToken)
	if c.loginCustomerID != "" {
		req.Header.Set("login-customer-id", c.loginCustomerID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to extract a human-readable error message from the API response.
		msg := extractErrorMessage(body)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, &GoogleAdsError{StatusCode: resp.StatusCode, Body: msg}
	}
	return body, nil
}

func extractErrorMessage(body []byte) string {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Details []struct {
				Errors []struct {
					Message string `json:"message"`
				} `json:"errors"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) != nil {
		return string(body)
	}
	// Prefer inner errors from GoogleAdsFailure
	for _, detail := range errResp.Error.Details {
		for _, e := range detail.Errors {
			if e.Message != "" {
				return e.Message
			}
		}
	}
	if errResp.Error.Message != "" {
		return errResp.Error.Message
	}
	return string(body)
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req)
}

func (c *Client) post(url string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req)
}

// ListAccessibleCustomers returns the resource names of all directly accessible customers.
func (c *Client) ListAccessibleCustomers() ([]string, error) {
	url := apiBase + "/customers:listAccessibleCustomers"
	body, err := c.get(url)
	if err != nil {
		return nil, err
	}
	var resp AccessibleCustomersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return resp.ResourceNames, nil
}

// Search executes a GAQL query and returns all result rows (handles pagination).
func (c *Client) Search(customerID, query string) ([]json.RawMessage, error) {
	url := fmt.Sprintf("%s/customers/%s/googleAds:search", apiBase, customerID)
	var allResults []json.RawMessage
	pageToken := ""

	for {
		payload := map[string]string{"query": query}
		if pageToken != "" {
			payload["pageToken"] = pageToken
		}
		body, err := c.post(url, payload)
		if err != nil {
			return nil, err
		}
		var resp SearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing search response: %w", err)
		}
		allResults = append(allResults, resp.Results...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return allResults, nil
}

// MutateCampaigns sends campaign mutation operations.
func (c *Client) MutateCampaigns(customerID string, operations []map[string]any) (*MutateResponse, error) {
	url := fmt.Sprintf("%s/customers/%s/campaigns:mutate", apiBase, customerID)
	return c.mutate(url, operations)
}

// MutateCampaignBudgets sends campaign budget mutation operations.
func (c *Client) MutateCampaignBudgets(customerID string, operations []map[string]any) (*MutateResponse, error) {
	url := fmt.Sprintf("%s/customers/%s/campaignBudgets:mutate", apiBase, customerID)
	return c.mutate(url, operations)
}

// MutateAdGroups sends ad group mutation operations.
func (c *Client) MutateAdGroups(customerID string, operations []map[string]any) (*MutateResponse, error) {
	url := fmt.Sprintf("%s/customers/%s/adGroups:mutate", apiBase, customerID)
	return c.mutate(url, operations)
}

// MutateAdGroupCriteria sends keyword (criterion) mutation operations.
func (c *Client) MutateAdGroupCriteria(customerID string, operations []map[string]any) (*MutateResponse, error) {
	url := fmt.Sprintf("%s/customers/%s/adGroupCriteria:mutate", apiBase, customerID)
	return c.mutate(url, operations)
}

func (c *Client) mutate(url string, operations []map[string]any) (*MutateResponse, error) {
	payload := map[string]any{"operations": operations}
	body, err := c.post(url, payload)
	if err != nil {
		return nil, err
	}
	var resp MutateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing mutate response: %w", err)
	}
	return &resp, nil
}

// ResourceID extracts the trailing numeric ID from a resource name.
// e.g. "customers/123/campaigns/456" → "456"
func ResourceID(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) == 0 {
		return resourceName
	}
	return parts[len(parts)-1]
}

// CleanCustomerID removes hyphens from a customer ID.
// e.g. "123-456-7890" → "1234567890"
func CleanCustomerID(id string) string {
	return strings.ReplaceAll(id, "-", "")
}

// MicrosToCurrency converts micros (int64-as-string) to a currency string.
// e.g. "5000000" → "5.00"
func MicrosToCurrency(micros string) string {
	if micros == "" {
		return "0.00"
	}
	n, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return micros
	}
	return fmt.Sprintf("%.2f", float64(n)/1_000_000)
}

// FormatMetricInt formats an int64-as-string metric for display.
func FormatMetricInt(s string) string {
	if s == "" {
		return "0"
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return s
	}
	return strconv.FormatInt(n, 10)
}

// FormatCTR formats a CTR float as a percentage string.
func FormatCTR(ctr float64) string {
	return fmt.Sprintf("%.2f%%", ctr*100)
}

// FormatROAS calculates and formats ROAS.
func FormatROAS(conversionsValue float64, costMicros string) string {
	n, err := strconv.ParseInt(costMicros, 10, 64)
	if err != nil || n == 0 {
		return "-"
	}
	cost := float64(n) / 1_000_000
	roas := conversionsValue / cost
	return fmt.Sprintf("%.2f", roas)
}
