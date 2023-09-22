# slacker

A Go utility to send structured reports in a formatted output to Slack.

## Report format

See `examples/full.json` for an example.

## Slack format

This will send an initial message to the specified `--channel`, with a summary
of each environment.

Then, a reply will be generated for each environment with a `completed` status,
detailing each individual namespace and the failures within each section.

## Usage

```bash
./slacker my-report.json --channel="alerts" --report-base-url https://reports.com --token slack-api-token --verbose
# or via stdin
generate-report.sh | ./slacker --channel="alerts" --report-base-url https://reports.com --token slack-api-token --verbose
```

See `./slacker --help` for flags.

**NOTE:** Flags can be replaced with env vars, eg. `--report-base-url` can be provided as `REPORT_BASE_URL=...`

```
Parses a report JSON document and sends a report to Slack

Usage:
  slacker [FILE] [flags]

Flags:
      --channel string           Slack channel name to send to
      --dry-run                  Use dry-run mode
  -h, --help                     help for slacker
      --report-base-url string   Base URL used to build links to reports
      --report-date string       Report date in dd-mm-yyyy format (default "22-09-2023")
      --token string             Slack API token to use
      --verbose                  Show Debug log output
```
