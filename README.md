# gads-cli

A command-line tool for the [Google Ads API v23](https://developers.google.com/google-ads/api/rest/overview), built for the `the20100` ecosystem.

- **JSON output when piped**, human-readable tables in a terminal
- **OAuth2 with automatic token refresh** — credentials stored in `~/.config/gads/credentials.json`
- **Raw REST calls** — no client library dependency, just `net/http`
- Single static binary, zero runtime dependencies

---

## Install

```bash
git clone https://github.com/the20100/gads-cli
cd gads-cli
go build -o gads-cli .
mv gads-cli /usr/local/bin/
```

**Requirements:** Go 1.22+

---

## Authentication setup

Google Ads API requires three things:

1. **OAuth2 credentials** (client_id + client_secret) — from a Google Cloud project
2. **Developer token** — from your Google Ads Manager Account
3. **Manager Account (MCC) ID** — your top-level Google Ads account

### Step 1 — Create a Google Cloud project

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a project (or use an existing one)
3. Enable the **Google Ads API** under APIs & Services → Library
4. Create OAuth2 credentials: **OAuth client ID → Desktop app**
5. Add `http://localhost:8080` as an authorized redirect URI
6. Download the credentials JSON file

### Step 2 — Get a developer token

1. Sign in to [Google Ads](https://ads.google.com) with your Manager Account
2. Go to **Tools → API Center**
3. Apply for a developer token (basic access is sufficient for testing)

> **Important:** Developer token approval can take a few days for production access.
> For testing, a test account with a test developer token works immediately.

### Step 3 — Run the login flow

```bash
# Using the downloaded credentials file:
gads-cli auth login --credentials-file=~/Downloads/client_secret_xyz.json

# Or enter all values interactively:
gads-cli auth login
```

The CLI will open your browser for Google OAuth authorization.
After approving, credentials are saved to `~/.config/gads/credentials.json`.

### Verify setup

```bash
gads-cli auth status
gads-cli auth check      # makes a test API call
gads-cli accounts list   # list all accounts under your MCC
```

---

## Global flags

| Flag | Description |
|------|-------------|
| `--json` | Force JSON output (even in a terminal) |
| `--pretty` | Force pretty-printed JSON (implies `--json`) |

Output is **auto-detected**: JSON when stdout is piped, tables in a terminal.

---

## Commands

### `auth`

```bash
gads-cli auth login [--credentials-file=path] [--developer-token=TOKEN] [--manager-account=ID]
gads-cli auth status        # show saved credentials summary
gads-cli auth token         # show current access/refresh tokens
gads-cli auth check         # validate credentials with a live API call
gads-cli auth logout        # delete saved credentials
```

---

### `accounts`

```bash
# List all client accounts under the MCC
gads-cli accounts list
gads-cli accounts list --json
```

**Output columns:** ID, NAME, CURRENCY, TIMEZONE, MANAGER, TEST

---

### `campaigns`

```bash
# List campaigns
gads-cli campaigns list --account=1234567890
gads-cli campaigns list --account=1234567890 --json

# Get campaign details
gads-cli campaigns get --account=1234567890 --campaign=111222333

# Pause / enable
gads-cli campaigns pause  --account=1234567890 --campaign=111222333
gads-cli campaigns enable --account=1234567890 --campaign=111222333

# Update daily budget (amount in micros — 5000000 = 5.00 in account currency)
gads-cli campaigns budget --account=1234567890 --campaign=111222333 --amount=5000000
```

**Output columns (list):** ID, NAME, STATUS, TYPE, DAILY BUDGET, START, END

---

### `adgroups`

```bash
# List ad groups in a campaign
gads-cli adgroups list --account=1234567890 --campaign=111222333

# Pause / enable
gads-cli adgroups pause  --account=1234567890 --adgroup=444555666
gads-cli adgroups enable --account=1234567890 --adgroup=444555666
```

**Output columns (list):** ID, NAME, STATUS, TYPE, DEFAULT BID

---

### `keywords`

```bash
# List keywords in a campaign
gads-cli keywords list --account=1234567890 --campaign=111222333

# Add a keyword
gads-cli keywords add --account=1234567890 --adgroup=444555666 \
  --keyword="running shoes" --match-type=PHRASE

# Pause a keyword (ID format: <adGroupId>~<criterionId>)
gads-cli keywords pause  --account=1234567890 --keyword=444555666~12345

# Remove a keyword
gads-cli keywords remove --account=1234567890 --keyword=444555666~12345
```

The keyword ID uses the Google Ads composite key format `<adGroupId>~<criterionId>`,
shown in the `ID` column of `keywords list`.

**Match types:** `BROAD`, `PHRASE`, `EXACT`

**Output columns (list):** ID, KEYWORD, MATCH, STATUS, QS, BID, AD GROUP

---

### `ads`

```bash
# List responsive search ads in an ad group
gads-cli ads list --account=1234567890 --adgroup=444555666
gads-cli ads list --account=1234567890 --adgroup=444555666 --json
```

---

### `insights`

All insight commands accept the following flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--account` | — | Customer account ID *(required)* |
| `--period` | — | Period shorthand (see table below); highest priority |
| `--days N` | 30 | Look back N days (ignored when `--period` is set) |
| `--start YYYY-MM-DD` | — | Start date (overrides `--days`, ignored when `--period` is set) |
| `--end YYYY-MM-DD` | — | End date (overrides `--days`, ignored when `--period` is set) |
| `--all` | false | Include rows with 0 impressions |
| `--preset` | `default` | Column preset: `default`, `performance`, `conversions`, `full` (ads: also `creatives`) |
| `--fields` | — | Comma-separated field IDs, overrides `--preset` |

**`--period` values:**

| Value | Meaning |
|-------|---------|
| `today` | Today |
| `yesterday` | Yesterday |
| `last7d` / `7d` | Last 7 days |
| `last14d`, `last30d`, `last90d` … | Last N days |
| `currentWeek` / `thisWeek` | Monday → today |
| `lastWeek` | Previous Mon–Sun |
| `currentMonth` / `thisMonth` | 1st → today |
| `lastMonth` | Full previous month |
| `last3m`, `last6m`, `last12m` | Last N months |
| `currentYear` / `thisYear` | Jan 1 → today |
| `lastYear` | Full previous year |
| `1y` / `2y` … | Last N years |
| `2024`, `2025` … | Full calendar year |

---

#### `insights campaigns`

```bash
gads-cli insights campaigns --account=1234567890 --days=30
gads-cli insights campaigns --account=1234567890 --period=lastMonth --preset=performance
gads-cli insights campaigns --account=1234567890 --period=2025 --preset=conversions
gads-cli insights campaigns --account=1234567890 --days=7 --fields=campaign_name,impressions,clicks,cost,roas
gads-cli insights campaigns --account=1234567890 --start=2024-01-01 --end=2024-01-31
```

**Presets:**

| Preset | Fields |
|--------|--------|
| `default` | campaign_name, campaign_status, impressions, clicks, cost, ctr, cpc, conversions, roas |
| `performance` | + abs_top_imp_pct, top_imp_pct, conv_rate, cost_per_conv |
| `conversions` | campaign_name, campaign_status, conversions, conv_value, view_through_conv, conv_rate, cost_per_conv, roas |
| `full` | All available fields |

**Available field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `campaign_id` | `campaign.id` | Campaign ID |
| `campaign_name` | `campaign.name` | Campaign name |
| `campaign_status` | `campaign.status` | Campaign status |
| `campaign_type` | `campaign.advertising_channel_type` | Campaign type (SEARCH, DISPLAY, …) |
| `impressions` | `metrics.impressions` | Impressions |
| `clicks` | `metrics.clicks` | Clicks |
| `cost` | `metrics.cost_micros` | Cost (currency units) |
| `ctr` | `metrics.ctr` | Click-through rate |
| `cpc` | `metrics.average_cpc` | Avg. cost per click |
| `conversions` | `metrics.conversions` | Conversions |
| `conv_value` | `metrics.conversions_value` | Conversion value |
| `roas` | computed | ROAS (conv_value / cost) |
| `abs_top_imp_pct` | `metrics.absolute_top_impression_percentage` | % impr. at absolute top position |
| `top_imp_pct` | `metrics.top_impression_percentage` | % impr. at top of page |
| `view_through_conv` | `metrics.view_through_conversions` | View-through conversions |
| `cost_per_conv` | `metrics.cost_per_conversion` | Cost per conversion |
| `conv_rate` | `metrics.conversions_from_interactions_rate` | Conversion rate |
| `search_imp_share` | `metrics.search_impression_share` | Search impression share |

---

#### `insights adgroups`

```bash
gads-cli insights adgroups --account=1234567890 --campaign=111222333 --days=7
gads-cli insights adgroups --account=1234567890 --campaign=111222333 --preset=performance
```

Required: `--campaign`

**Presets:**

| Preset | Fields |
|--------|--------|
| `default` | adgroup_name, adgroup_status, impressions, clicks, cost, ctr, cpc, conversions, roas |
| `performance` | + campaign_name, conv_rate, cost_per_conv, abs_top_imp_pct, top_imp_pct |
| `conversions` | campaign_name, adgroup_name, adgroup_status, conversions, conv_value, view_through_conv, conv_rate, cost_per_conv, roas |
| `full` | All available fields |

**Additional field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `campaign_name` | `campaign.name` | Campaign name |
| `adgroup_id` | `ad_group.id` | Ad group ID |
| `adgroup_name` | `ad_group.name` | Ad group name |
| `adgroup_status` | `ad_group.status` | Ad group status |

*(All metrics from campaigns are also available)*

---

#### `insights keywords`

```bash
gads-cli insights keywords --account=1234567890 --campaign=111222333 --days=30
gads-cli insights keywords --account=1234567890 --campaign=111222333 --preset=performance
```

Required: `--campaign`

**Presets:**

| Preset | Fields |
|--------|--------|
| `default` | keyword_text, keyword_match, keyword_status, impressions, clicks, cost, ctr, cpc, conversions, quality_score |
| `performance` | + campaign_name, adgroup_name, roas, conv_rate, cost_per_conv |
| `conversions` | keyword_text, keyword_match, keyword_status, conversions, conv_value, conv_rate, cost_per_conv, roas |
| `full` | All available fields |

**Additional field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `keyword_text` | `ad_group_criterion.keyword.text` | Keyword text |
| `keyword_match` | `ad_group_criterion.keyword.match_type` | Match type |
| `keyword_status` | `ad_group_criterion.status` | Keyword status |
| `quality_score` | `ad_group_criterion.quality_info.quality_score` | Quality score (1–10) |
| `campaign_name` | `campaign.name` | Campaign name |
| `adgroup_name` | `ad_group.name` | Ad group name |

*(All metrics from campaigns are also available)*

---

#### `insights search-terms`

```bash
gads-cli insights search-terms --account=1234567890 --campaign=111222333 --days=14
gads-cli insights search-terms --account=1234567890 --campaign=111222333 --preset=performance
```

Required: `--campaign`

**Presets:**

| Preset | Fields |
|--------|--------|
| `default` | search_term, st_status, adgroup_name, impressions, clicks, cost, ctr, conversions |
| `performance` | + campaign_name, cpc, roas, conv_rate, cost_per_conv |
| `conversions` | search_term, st_status, campaign_name, adgroup_name, conversions, conv_value, conv_rate, cost_per_conv, roas |
| `full` | All available fields |

**Additional field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `search_term` | `search_term_view.search_term` | Search query |
| `st_status` | `search_term_view.status` | Search term status |
| `campaign_name` | `campaign.name` | Campaign name |
| `adgroup_name` | `ad_group.name` | Ad group name |

*(Most metrics from campaigns are also available)*

---

#### `insights ads`

Ad-level report with creative details (RSA headlines, descriptions, URLs). `--campaign` is optional.

```bash
gads-cli insights ads --account=1234567890 --days=30
gads-cli insights ads --account=1234567890 --campaign=111222333 --preset=creatives
gads-cli insights ads --account=1234567890 --days=7 --fields=campaign_name,ad_name,headline1,headline2,headline3,desc1,desc2
gads-cli insights ads --account=1234567890 --days=30 --json
```

**Presets:**

| Preset | Fields |
|--------|--------|
| `default` | campaign_name, adgroup_name, ad_name, ad_status, ad_type, final_url + core metrics |
| `performance` | campaign_name, adgroup_name, ad_name, ad_status + full metrics incl. impression share |
| `creatives` | campaign_name, adgroup_name, ad_name, ad_status, ad_type, final_url, display_url, path1, path2, headline1–15, desc1–4 |
| `full` | All dimensions + all RSA/ETA creative fields + all metrics |

**Ad dimension field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `campaign_name` | `campaign.name` | Campaign name |
| `adgroup_name` | `ad_group.name` | Ad group name |
| `ad_id` | `ad_group_ad.ad.id` | Ad ID |
| `ad_name` | `ad_group_ad.ad.name` | Ad name |
| `ad_status` | `ad_group_ad.status` | Ad status |
| `ad_type` | `ad_group_ad.ad.type` | Ad type (RESPONSIVE_SEARCH_AD, EXPANDED_TEXT_AD, …) |
| `final_url` | `ad_group_ad.ad.final_urls` | Final destination URL |
| `final_mobile_url` | `ad_group_ad.ad.final_mobile_urls` | Final mobile URL |
| `tracking_url` | `ad_group_ad.ad.tracking_url_template` | Tracking URL template |
| `url_suffix` | `ad_group_ad.ad.final_url_suffix` | Final URL suffix |
| `display_url` | `ad_group_ad.ad.display_url` | Display URL |
| `path1` | `ad_group_ad.ad.responsive_search_ad.path1` | URL display path 1 |
| `path2` | `ad_group_ad.ad.responsive_search_ad.path2` | URL display path 2 |

**RSA headline field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `headline1` … `headline15` | `ad_group_ad.ad.responsive_search_ad.headlines` | RSA headline text (by array index) |
| `headline1_pos` … `headline15_pos` | `ad_group_ad.ad.responsive_search_ad.headlines` | RSA headline pinned position |

**RSA description field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `desc1` … `desc4` | `ad_group_ad.ad.responsive_search_ad.descriptions` | RSA description text |
| `desc1_pos` … `desc4_pos` | `ad_group_ad.ad.responsive_search_ad.descriptions` | RSA description pinned position |

**ETA (legacy Expanded Text Ad) field IDs:**

| Field ID | GAQL field | Description |
|----------|------------|-------------|
| `eta_headline1` | `ad_group_ad.ad.expanded_text_ad.headline_part1` | ETA headline 1 |
| `eta_headline2` | `ad_group_ad.ad.expanded_text_ad.headline_part2` | ETA headline 2 |
| `eta_headline3` | `ad_group_ad.ad.expanded_text_ad.headline_part3` | ETA headline 3 |
| `eta_desc1` | `ad_group_ad.ad.expanded_text_ad.description` | ETA description 1 |
| `eta_desc2` | `ad_group_ad.ad.expanded_text_ad.description2` | ETA description 2 |

*(All metrics from campaigns are also available)*

---

### `info`

```bash
gads-cli info    # show binary path, OS, config file location, auth status
```

---

### `update`

Self-update by cloning the latest source from GitHub and rebuilding:

```bash
gads-cli update
```

Requires `git` and `go` to be installed.

---

## Scripting / agent use

When piped, all commands emit JSON automatically:

```bash
# List all campaign IDs
gads-cli campaigns list --account=1234567890 | jq '.[].campaign.id'

# Get cost for top campaigns last 7 days
gads-cli insights campaigns --account=1234567890 --days=7 \
  | jq '.[] | {name: .campaign.name, cost: .metrics.costMicros}'

# Check if a campaign is paused
gads-cli campaigns get --account=1234567890 --campaign=111222333 \
  | jq '.campaign.status'

# Export all RSA headline data as JSON
gads-cli insights ads --account=1234567890 --days=30 \
  | jq '.[] | {ad: .adGroupAd.ad.name, headlines: .adGroupAd.ad.responsiveSearchAd.headlines}'
```

---

## Credential file format

`~/.config/gads/credentials.json` (permissions: 0600):

```json
{
  "client_id": "your-client-id.apps.googleusercontent.com",
  "client_secret": "your-client-secret",
  "developer_token": "your-22-char-developer-token",
  "manager_customer_id": "1234567890",
  "refresh_token": "1//...",
  "access_token": "ya29...",
  "token_type": "Bearer",
  "token_expiry": "2024-01-15T10:30:00Z"
}
```

The access token is refreshed automatically when it expires. Only the refresh token
is permanent — it is obtained during `auth login` and persists across sessions.

---

## Notes

- **Budget amounts** are in micros: `1,000,000 micros = 1.00` in the account's currency.
  The `--amount` flag for `campaigns budget` takes micros directly.
- **Customer IDs** can be provided with or without hyphens (`123-456-7890` or `1234567890`).
- **API version:** Google Ads REST API v23 (`https://googleads.googleapis.com/v23/`)
- **Pagination:** handled automatically — all results are returned regardless of page size.
- **Insights presets:** use `--preset` for quick access to common column sets; `--fields` for fine-grained control.
- **RSA headlines** are returned as an array by the API and indexed 1–15 by position in the array.

---

## License

MIT
