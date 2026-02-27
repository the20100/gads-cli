# gads-cli

A command-line tool for the [Google Ads API v19](https://developers.google.com/google-ads/api/rest/overview), built for the `the20100` ecosystem.

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

All insight commands accept `--days` (default: 30) **or** `--start`/`--end` for a specific range.

```bash
# Campaign performance
gads-cli insights campaigns --account=1234567890 --days=30
gads-cli insights campaigns --account=1234567890 --start=2024-01-01 --end=2024-01-31

# Ad group performance
gads-cli insights adgroups --account=1234567890 --campaign=111222333 --days=7

# Keyword performance
gads-cli insights keywords --account=1234567890 --campaign=111222333 --days=30

# Search terms report
gads-cli insights search-terms --account=1234567890 --campaign=111222333 --days=14
```

**Campaigns output columns:** ID, NAME, IMPRESSIONS, CLICKS, COST, CTR, CPC, CONV, ROAS

Cost is displayed in currency units (micros ÷ 1,000,000).
ROAS = conversion value ÷ cost.

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
- **API version:** Google Ads REST API v19 (`https://googleads.googleapis.com/v19/`)
- **Pagination:** handled automatically — all results are returned regardless of page size.

---

## License

MIT
