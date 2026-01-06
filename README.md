# costdiff

Compare AWS costs between time periods from the command line.

```
AWS Cost Diff: Dec 2024 → Jan 2025

Total: $12,847.23 → $15,234.56 (+$2,387.33 / +18.6%)

Service                     Dec 2024    Jan 2025      Change
─────────────────────────────────────────────────────────────
EC2-Instances               $7,521.00   $9,363.00    +$1,842.00 (+24.5%)
RDS                         $2,010.00   $2,633.00      +$623.00 (+31.0%)
S3                          $1,300.00   $1,144.00      -$156.00 (-12.0%)
Lambda                        $975.00   $1,053.00       +$78.00 (+8.0%)
```

## Installation

### Homebrew (macOS/Linux)

```bash
brew install hserkanyilmaz/tap/costdiff
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/hserkanyilmaz/costdiff/releases).

```bash
# macOS (Apple Silicon)
curl -Lo costdiff https://github.com/hserkanyilmaz/costdiff/releases/latest/download/costdiff_Darwin_arm64
chmod +x costdiff
sudo mv costdiff /usr/local/bin/

# macOS (Intel)
curl -Lo costdiff https://github.com/hserkanyilmaz/costdiff/releases/latest/download/costdiff_Darwin_x86_64
chmod +x costdiff
sudo mv costdiff /usr/local/bin/

# Linux (x86_64)
curl -Lo costdiff https://github.com/hserkanyilmaz/costdiff/releases/latest/download/costdiff_Linux_x86_64
chmod +x costdiff
sudo mv costdiff /usr/local/bin/
```

### From Source

```bash
go install github.com/hserkanyilmaz/costdiff@latest
```

## Quick Start

```bash
# Compare last month vs current month
costdiff

# Compare specific months
costdiff --from 2024-10 --to 2024-12

# Show top cost drivers
costdiff top

# Drill down into a service's usage types
costdiff top --service "Amazon EC2" -g usage-type

# View daily trends
costdiff watch
```

## Commands

### `costdiff` (default)

Compare costs between two time periods.

```bash
costdiff                              # last month vs current
costdiff --from 2024-10 --to 2024-12  # specific months
costdiff -g region                    # group by region
costdiff -g account                   # group by linked account
costdiff -g tag --tag team            # group by tag
costdiff --threshold 100              # only show changes > $100
costdiff --min-cost 50                # only show items >= $50
costdiff -n 20                        # show top 20 items
costdiff -s cost                      # sort by current cost
costdiff -s diff-pct                  # sort by percentage change
costdiff -o json                      # output as JSON
costdiff -o csv                       # output as CSV
```

### `costdiff top`

Show current top cost drivers.

```bash
costdiff top                # top 10 services this month
costdiff top -n 20          # top 20
costdiff top -g region      # top by region
costdiff top --from 2024-10 # top costs for October 2024
```

### `costdiff watch`

Daily cost trend.

```bash
costdiff watch              # last 7 days
costdiff watch --days 30    # last 30 days
costdiff watch -o json      # output as JSON
```

### `costdiff version`

Print version information.

```bash
costdiff version
```

## Drill Down by Usage Type

See what's driving costs within a specific service by drilling down into usage types.

```bash
# Step 1: See top services
costdiff top

# Step 2: Drill into a specific service
costdiff top --service "Amazon Elastic Compute Cloud - Compute" -g usage-type
```

This shows granular cost breakdowns like:
- `BoxUsage:t3.medium` - EC2 instance hours
- `DataTransfer-Out-Bytes` - Data transfer costs
- `EBS:VolumeUsage.gp3` - Storage volumes
- `Requests-Tier1` - API request costs

### Examples

```bash
# EC2 usage breakdown
costdiff top --service "Amazon EC2" -g usage-type

# S3 usage breakdown
costdiff top --service "Amazon Simple Storage Service" -g usage-type

# Compare usage types between months
costdiff --service "Amazon RDS Service" -g usage-type

# Top 20 usage types, sorted by cost
costdiff top --service "AWS Lambda" -g usage-type -n 20 --sort cost
```

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--from` | `-f` | Start period (YYYY-MM or YYYY-MM-DD) | Last month |
| `--to` | `-t` | End period (YYYY-MM or YYYY-MM-DD) | Current month |
| `--group` | `-g` | Group by: service\|usage-type\|region\|account\|tag | service |
| `--service` | | Filter by AWS service name (for drill-down) | |
| `--tag` | | Tag key when grouping by tag | |
| `--metric` | `-m` | Cost metric (see below) | net-amortized |
| `--top` | `-n` | Number of results | 10 |
| `--format` | `-o` | Output: table\|json\|csv | table |
| `--sort` | `-s` | Sort by: diff\|diff-pct\|cost\|name | diff |
| `--profile` | `-p` | AWS profile | |
| `--region` | `-r` | AWS region | us-east-1 |
| `--threshold` | | Only show changes above $X | 0 |
| `--min-cost` | | Only show items where from or to cost >= $X | 0 |
| `--quiet` | `-q` | Suppress non-essential output | false |
| `--verbose` | `-v` | Debug output | false |

### Grouping Options

| Group | Description |
|-------|-------------|
| `service` | AWS service (default) |
| `usage-type` | Usage type within a service (use with `--service`) |
| `region` | AWS region |
| `account` | Linked AWS account |
| `tag` | Cost allocation tag (requires `--tag`) |

### Cost Metrics

| Metric | Description |
|--------|-------------|
| `net-amortized` | Net amortized cost (default) - includes discounts, RI/SP amortization, minus credits |
| `amortized` | Amortized cost - spreads upfront RI/SP payments across the term |
| `unblended` | Unblended cost - actual hourly rates |
| `blended` | Blended cost - average rate across organization |
| `net-unblended` | Net unblended cost - unblended minus credits |

```bash
# Use unblended costs instead of net amortized
costdiff -m unblended

# Compare amortized costs
costdiff -m amortized --from 2024-10 --to 2024-12
```

## AWS Configuration

costdiff uses the standard AWS credential chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (when running on EC2/ECS/Lambda)

### Using Named Profiles

```bash
# Use a specific profile
costdiff --profile production

# Or set the environment variable
export AWS_PROFILE=production
costdiff
```

## IAM Permissions

costdiff requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "ce:GetCostForecast"
      ],
      "Resource": "*"
    }
  ]
}
```

### Creating an IAM Policy

1. Go to AWS Console → IAM → Policies → Create Policy
2. Select JSON tab and paste the policy above
3. Name it `CostExplorerReadOnly`
4. Attach to your user or role

## Output Formats

### Table (default)

Human-readable table with colored output for increases (red) and decreases (green).

### JSON

```bash
costdiff -o json
```

```json
{
  "from_period": {
    "start": "2024-12-01",
    "end": "2025-01-01",
    "label": "Dec 2024"
  },
  "to_period": {
    "start": "2025-01-01",
    "end": "2025-02-01",
    "label": "Jan 2025"
  },
  "from_total": 12847.23,
  "to_total": 15234.56,
  "total_diff": 2387.33,
  "total_diff_percent": 18.58,
  "items": [...]
}
```

### CSV

```bash
costdiff -o csv > costs.csv
```

## Troubleshooting

### "AWS credentials not found"

Make sure you have configured AWS credentials. See [AWS Configuration](#aws-configuration).

### "Access denied"

Your IAM user/role needs Cost Explorer permissions. See [IAM Permissions](#iam-permissions).

### "Cost Explorer has not been enabled"

Cost Explorer must be enabled in your AWS account:

1. Go to AWS Console → Billing → Cost Explorer
2. Click "Enable Cost Explorer"
3. Wait up to 24 hours for data to be available

### "No cost data found"

- Cost Explorer data is typically available 24-48 hours after charges are incurred
- Make sure the date range includes dates with actual AWS usage
- Try a broader date range

## Development

```bash
# Clone the repository
git clone https://github.com/hserkanyilmaz/costdiff.git
cd costdiff

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint

# Build for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.
