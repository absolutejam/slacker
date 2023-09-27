package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dsab.slacker/report"
	"dsab.slacker/slacknotify"
)

const (
	SlackFlagChannel            = "channel"
	SlackFlagToken              = "token"
	SlackFlagReportDate         = "report-date"
	SlackFlagReportBaseUrl      = "report-base-url"
	SlackFlagUpdateEnvironments = "update-environments"
	SlackFlagUpdateMessageTs    = "update-message-ts"
	SlackFlagDryRun             = "dry-run"
	SlackFlagVerbose            = "verbose"
	SlackFlagLookupLastReport   = "lookup-last-report"
)

func init() {
	viper.AutomaticEnv()

	SlackCmd.Flags().String(SlackFlagChannel, "", "[REQUIRED] Slack channel name to send to")
	viper.BindPFlag(SlackFlagChannel, SlackCmd.Flags().Lookup(SlackFlagChannel))

	SlackCmd.Flags().String(SlackFlagToken, "", "[REQUIRED] Slack API token to use")
	viper.BindPFlag(SlackFlagToken, SlackCmd.Flags().Lookup(SlackFlagToken))

	SlackCmd.Flags().String(SlackFlagReportBaseUrl, "", "[REQUIRED] Base URL used to build links to reports")
	viper.BindPFlag(SlackFlagReportBaseUrl, SlackCmd.Flags().Lookup(SlackFlagReportBaseUrl))

	SlackCmd.Flags().Bool(SlackFlagUpdateEnvironments, true, "Whether to update existing environment messages")
	viper.BindPFlag(SlackFlagUpdateEnvironments, SlackCmd.Flags().Lookup(SlackFlagUpdateEnvironments))

	SlackCmd.Flags().String(SlackFlagUpdateMessageTs, "", "The TS of a message to update & reply to")
	viper.BindPFlag(SlackFlagUpdateMessageTs, SlackCmd.Flags().Lookup(SlackFlagUpdateMessageTs))

	SlackCmd.Flags().String(SlackFlagReportDate, time.Now().Format("02-01-2006"), "Report date in dd-mm-yyyy format")
	viper.BindPFlag(SlackFlagReportDate, SlackCmd.Flags().Lookup(SlackFlagReportDate))

	SlackCmd.Flags().Bool(SlackFlagLookupLastReport, false, "Look up the last report automatically")
	viper.BindPFlag(SlackFlagLookupLastReport, SlackCmd.Flags().Lookup(SlackFlagLookupLastReport))

	SlackCmd.Flags().Bool(SlackFlagDryRun, false, "Use dry-run mode")
	viper.BindPFlag(SlackFlagDryRun, SlackCmd.Flags().Lookup(SlackFlagDryRun))
}

func readJsonReportFromFileOrStdin(cmd *cobra.Command, args []string) (*report.ReportJson, error) {
	var (
		reader io.Reader
		err    error
	)

	// This is validated by the command's `cobra.ExactArgs(0)`
	filename := args[0]
	if filename == "-" {
		reader = cmd.InOrStdin()
	} else {
		reader, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}

	bytes, err := io.ReadAll(reader)
	return report.FromJson(bytes)
}

type Output struct {
	ResponseTimestamp slacknotify.ResponseTimestamp
}

// determineUpdate checks to see if either `--update-message-ts` or `--lookup-last-report`
// have been set and will set the value of `update` either directly, or by looking up
// the existing report via. the provided `reportFinder`
func determineUpdate(updateMessageTs *slacknotify.ResponseTimestamp, reportFinder slacknotify.ReportFinder) error {
	var (
		updateMessageTsString = viper.GetString(SlackFlagUpdateMessageTs)
		lookupLastReport      = viper.GetBool(SlackFlagLookupLastReport)
		reportDate            = viper.GetString(SlackFlagReportDate)
	)

	if updateMessageTsString != "" && lookupLastReport {
		return fmt.Errorf("flags '--%s' and '--%s' are mutually exclusive", SlackFlagUpdateMessageTs, SlackFlagLookupLastReport)

		//
	} else if updateMessageTsString != "" {
		// Set the value of the provided `updateMessageTs` pointer to the the one from the flag
		log.Debugf("Using explicit update message timestamp: %v\n", updateMessageTsString)
		updateMessageTs.Ts = updateMessageTsString
		return nil

		//
	} else if lookupLastReport {
		// Set the value of the provided `updateMessageTs` pointer to the the one that was looked up
		r, err := reportFinder.FindReport(reportDate)
		if err != nil {
			return fmt.Errorf("failed to look up last report: %v\n", err)
		}
		if r != nil {
			log.Debugf("Found previous report: %v\n", r)
			updateMessageTs.Ts = r.Ts
			return nil
		}
		log.Warnf("could not find last report for %s - Falling back to sending a new message", reportDate)
	}

	return nil
}

// sendNotifications sends notifications to the specific channels, optionally
// updating existing messages
func sendNotifications(slackNotifier slacknotify.SlackNotifier, reportFinder slacknotify.ReportFinder, reportJson report.ReportJson, updateMessageTs *slacknotify.ResponseTimestamp, updateEnvironmentMessages bool) (*slacknotify.ResponseTimestamp, error) {
	log.Info("Building summary report")

	parentMessageTs, err := slackNotifier.SendSummaryReport(reportJson, updateMessageTs)
	if err != nil {
		return nil, fmt.Errorf("failed to send summary report: %v", err)
	}

	log.Info("Building detailed environment reports")

	for _, env := range reportJson.Environments {
		if updateEnvironmentMessages {
			log.WithField("env", env.Name).Debug("Looking up existing environment")
			updateEnvironmentReportTs, err := reportFinder.FindEnvironmentReport(env.Name, parentMessageTs)
			if err != nil {
				return &parentMessageTs, err
			}

			err = slackNotifier.SendEnvironmentReport(parentMessageTs, env, updateEnvironmentReportTs)
			if err != nil {
				return &parentMessageTs, fmt.Errorf("failed to send environment report: %v", err)
			}
		} else {
			if env.Status == report.Completed {
				err := slackNotifier.SendEnvironmentReport(parentMessageTs, env, nil)
				if err != nil {
					return &parentMessageTs, fmt.Errorf("failed to send environment report: %v", err)
				}
			} else {
				log.WithField("environment", env.Name).Warnf("Not sending environment report for %s as it has status %s", env.Name, env.Status)
			}
		}
	}

	return nil, nil
}

var SlackCmd = &cobra.Command{
	Use:   "slack-report [FILE]",
	Short: "Parses a report JSON document and sends a report to Slack",

	Args: cobra.ExactArgs(1),
	Long: `Parses a report JSON, either from file or from stdin.

Example JSON report:
{
  "environments": [
    {
      "name": "abx-xyz-foo-2",
      "sections": [
        {
          "name": "Failed Deployments",
          "icon": ":package:",
          "failures": ["foo", "bar", "baz"]
        }
      ]
    }
  ]
}
`,
	Example: `# Automatically look up today's report and attempt to update it, or create a new message if it doesn't exist
slacker slack-report --channel alerts --token redacted --report-base-url https://my-reports --look-up-last-report report.json

# Send the report for a specific date, updating it if it already exists
slacker slack-report --channel alerts --token redacted --report-base-url https://my-reports --look-up-last-report --report-date "03-01-2023" report.json

# Send the report, specifying a specific message timestamp to overwrite
# (This will fail if the message doesn't exist)
slacker slack-report --channel alerts --token redacted --report-base-url https://my-reports --update-message-ts abcdef.abcdef report.json

# Collect the report JSON from stdin and create a brand new report
my-report.sh | slacker slack-report --token redacted --channel alerts --report-base-url https://my-reports -

# Using env vars for config instead of CLI flags
TOKEN=redacted CHANNEL=alerts REPORT_BASE_URL=https://my-reports slacker slack-report`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requireFlags(
			SlackFlagChannel,
			SlackFlagToken,
			SlackFlagReportBaseUrl,
		)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			channel            = viper.GetString(SlackFlagChannel)
			token              = viper.GetString(SlackFlagToken)
			reportDate         = viper.GetString(SlackFlagReportDate)
			reportBaseUrl      = viper.GetString(SlackFlagReportBaseUrl)
			updateEnvironments = viper.GetBool(SlackFlagUpdateEnvironments)
			dryRun             = viper.GetBool(SlackFlagDryRun)

			updateMessageTs slacknotify.ResponseTimestamp

			slackNotifier slacknotify.SlackNotifier
			reportFinder  slacknotify.ReportFinder
		)

		if dryRun {
			slackNotifier = slacknotify.NewDebugNotifier()
			reportFinder = slacknotify.NewNoOpReportFinder()
		} else {
			slackNotifier = slacknotify.NewNotifier(
				token,
				channel,
				report.ReportConfig{
					ReportDate: reportDate,
					BaseUrl:    reportBaseUrl,
				},
			)
			reportFinder = slacknotify.NewSlackReportFinder(token, channel)
		}

		reportJson, err := readJsonReportFromFileOrStdin(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read json report: %v", err)
		}

		if err := determineUpdate(&updateMessageTs, reportFinder); err != nil {
			return err
		}

		if dryRun {
			bytes, _ := json.MarshalIndent(reportJson, "", "  ")
			log.Debug(string(bytes))
		}

		summaryReportMessageTs, err := sendNotifications(slackNotifier, reportFinder, *reportJson, &updateMessageTs, updateEnvironments)
		if err != nil {
			return err
		}

		if summaryReportMessageTs != nil {
			output := Output{ResponseTimestamp: updateMessageTs}
			if json, err := json.MarshalIndent(output, "", "  "); err == nil {
				fmt.Println(string(json))
			}
		}

		return nil
	},
}
