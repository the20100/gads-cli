package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/gads-cli/internal/api"
	"github.com/the20100/gads-cli/internal/output"
)

var adsCmd = &cobra.Command{
	Use:   "ads",
	Short: "View Google Ads responsive search ads",
}

var (
	adsAccount    string
	adsAdGroupID  string
)

// ---- ads list ----

var adsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List responsive search ads in an ad group",
	Long: `List responsive search ads (RSAs) with their headlines, descriptions, and status.

Examples:
  gads-cli ads list --account=1234567890 --adgroup=444555666
  gads-cli ads list --account=1234567890 --adgroup=444555666 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if adsAccount == "" {
			return fmt.Errorf("--account is required")
		}
		if adsAdGroupID == "" {
			return fmt.Errorf("--adgroup is required")
		}
		cid := api.CleanCustomerID(adsAccount)

		query := fmt.Sprintf(`SELECT ad_group_ad.ad.id, ad_group_ad.ad.type,
			ad_group_ad.ad.responsive_search_ad.headlines,
			ad_group_ad.ad.responsive_search_ad.descriptions,
			ad_group_ad.ad.final_urls, ad_group_ad.status,
			ad_group.id, campaign.id
		FROM ad_group_ad
		WHERE ad_group_ad.ad.type = 'RESPONSIVE_SEARCH_AD'
		  AND ad_group_ad.status != 'REMOVED'
		  AND ad_group.id = '%s'
		ORDER BY ad_group_ad.ad.id`, adsAdGroupID)

		rows, err := apiClient.Search(cid, query)
		if err != nil {
			return err
		}

		var ads []api.AdRow
		for _, raw := range rows {
			var row api.AdRow
			if err := json.Unmarshal(raw, &row); err != nil {
				continue
			}
			ads = append(ads, row)
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(ads, output.IsPretty(cmd))
		}
		if len(ads) == 0 {
			fmt.Println("No responsive search ads found.")
			return nil
		}

		for _, r := range ads {
			fmt.Printf("Ad ID: %s  Status: %s\n", r.AdGroupAd.Ad.ID, r.AdGroupAd.Status)
			// Show up to 3 headlines
			headlines := r.AdGroupAd.Ad.ResponsiveSearchAd.Headlines
			if len(headlines) > 0 {
				hl := make([]string, 0, 3)
				for j, h := range headlines {
					if j >= 3 {
						break
					}
					hl = append(hl, h.Text)
				}
				fmt.Printf("  Headlines:    %s\n", strings.Join(hl, " | "))
				if len(headlines) > 3 {
					fmt.Printf("                (+%d more)\n", len(headlines)-3)
				}
			}
			// Show up to 2 descriptions
			descs := r.AdGroupAd.Ad.ResponsiveSearchAd.Descriptions
			if len(descs) > 0 {
				dl := make([]string, 0, 2)
				for j, d := range descs {
					if j >= 2 {
						break
					}
					dl = append(dl, d.Text)
				}
				fmt.Printf("  Descriptions: %s\n", strings.Join(dl, " | "))
				if len(descs) > 2 {
					fmt.Printf("                (+%d more)\n", len(descs)-2)
				}
			}
			if len(r.AdGroupAd.Ad.FinalUrls) > 0 {
				fmt.Printf("  Final URL:    %s\n", r.AdGroupAd.Ad.FinalUrls[0])
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	adsListCmd.Flags().StringVar(&adsAccount, "account", "", "Customer account ID (required)")
	adsListCmd.Flags().StringVar(&adsAdGroupID, "adgroup", "", "Ad group ID (required)")

	adsCmd.AddCommand(adsListCmd)
	rootCmd.AddCommand(adsCmd)
}
