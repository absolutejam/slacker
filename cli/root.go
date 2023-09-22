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
	RootFlagChannel       = "channel"
	RootFlagToken         = "token"
	RootFlagReportDate    = "report-date"
	RootFlagReportBaseUrl = "report-base-url"
	RootFlagDryRun        = "dry-run"
	RootFlagVerbose       = "verbose"
)

func init() {
	viper.AutomaticEnv()

	RootCmd.Flags().String(RootFlagChannel, "", "Slack channel name to send to")
	RootCmd.MarkFlagRequired(RootFlagChannel)
	viper.BindPFlag(RootFlagChannel, RootCmd.Flags().Lookup(RootFlagChannel))

	RootCmd.Flags().String(RootFlagToken, "", "Slack API token to use")
	RootCmd.MarkFlagRequired(RootFlagToken)
	viper.BindPFlag(RootFlagToken, RootCmd.Flags().Lookup(RootFlagToken))

	RootCmd.Flags().String(RootFlagReportBaseUrl, "", "Base URL used to build links to reports")
	RootCmd.MarkFlagRequired(RootFlagReportBaseUrl)
	viper.BindPFlag(RootFlagReportBaseUrl, RootCmd.Flags().Lookup(RootFlagReportBaseUrl))

	RootCmd.Flags().String(RootFlagReportDate, time.Now().Format("02-01-2006"), "Report date in dd-mm-yyyy format")
	viper.BindPFlag(RootFlagReportDate, RootCmd.Flags().Lookup(RootFlagReportDate))

	RootCmd.Flags().Bool(RootFlagDryRun, false, "Use dry-run mode")
	viper.BindPFlag(RootFlagDryRun, RootCmd.Flags().Lookup(RootFlagDryRun))

	RootCmd.Flags().Bool(RootFlagVerbose, false, "Show Debug log output")
	viper.BindPFlag(RootFlagVerbose, RootCmd.Flags().Lookup(RootFlagVerbose))
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

var RootCmd = &cobra.Command{
	Use:   "slacker [FILE]",
	Short: "Parses a report JSON document and sends a report to Slack",

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			channel       = viper.GetString(RootFlagChannel)
			token         = viper.GetString(RootFlagToken)
			reportDate    = viper.GetString(RootFlagReportDate)
			reportBaseUrl = viper.GetString(RootFlagReportBaseUrl)
			dryRun        = viper.GetBool(RootFlagDryRun)
			verbose       = viper.GetBool(RootFlagVerbose)
		)

		reportJson, err := readJsonReportFromFileOrStdin(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read json report: %v", err)
		}

		if verbose {
			log.SetLevel(log.DebugLevel)
		}

		if dryRun {

			bytes, _ := json.MarshalIndent(reportJson, "", "  ")
			log.Debug(string(bytes))
		}

		reportConfig := report.ReportConfig{
			ReportDate: reportDate,
			BaseUrl:    reportBaseUrl,
		}

		var slacknotifier slacknotify.SlackNotifier
		if dryRun {
			slacknotifier = slacknotify.NewDebugNotifier()
		} else {
			slacknotifier = slacknotify.NewNotifier(token, channel)
		}

		log.Info("Building summary report")
		respTimestamp, err := slacknotifier.SendSummaryReport(reportConfig, *reportJson)
		if err != nil {
			return fmt.Errorf("failed to send summary report: %v", err)
		}

		log.Info("Building detailed environment reports")
		for _, env := range reportJson.Environments {
			// TODO: Refactor this so the check is elsewhere
			if env.Status == report.Completed {
				err := slacknotifier.SendEnvironmentReport(respTimestamp, env)
				if err != nil {
					return fmt.Errorf("failed to send environment report: %v", err)
				}
			} else {
				log.WithField("environment", env.Name).Warnf("Not sending environment report for %s as it has status %s", env.Name, env.Status)
			}
		}

		return nil
	},
}
