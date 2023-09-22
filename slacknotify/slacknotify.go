package slacknotify

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"dsab.slacker/report"
)

func buildReportUrl(reportConfig report.ReportConfig, env report.ReportEnvironment) string {
	// TODO: Placeholder
	return fmt.Sprintf("%s/%s/%s", reportConfig.BaseUrl, reportConfig.ReportDate, env.Name)
}

func buildSummaryReportBlocks(reportConfig report.ReportConfig, reportJson report.ReportJson) []slack.Block {
	blocks := []slack.Block{
		slack.NewHeaderBlock(plaintext(":stethoscope: Bring-up Healthchecks")),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				markdown(":date: *Date:* 20/09/2023"),
				markdown(":rocket: *Jenkins job:* <https://google.com|1289>"),
			},
			nil,
		),
		slack.NewContextBlock("", markdown("Non-prod environments")),
		slack.NewDividerBlock(),
	}

	for _, env := range reportJson.Environments {
		blocks = append(blocks, buildEnvironmentSummarySection(reportConfig, env))
	}

	blocks = append(blocks,
		slack.NewActionBlock("",
			linkButton(":arrow_left: Yesterday's report", "https://google.com"),
			linkButton(":information_source: Learn more", "https://google.com"),
		),
	)

	return blocks
}

//-----------------------------------------------------------------------------------------

type SlackNotifier interface {
	SendSummaryReport(reportConfig report.ReportConfig, report report.ReportJson) (string, error)
	SendEnvironmentReport(respTimestamp string, env report.ReportEnvironment) error
}

// Interface assertions
var (
	_ SlackNotifier = (*slackNotifierConfig)(nil)
	_ SlackNotifier = (*debugSlackNotifier)(nil)
)

//-----------------------------------------------------------------------------------------

type debugSlackNotifier struct{}

func NewDebugNotifier() *debugSlackNotifier {
	return &debugSlackNotifier{}
}

func (c *debugSlackNotifier) SendSummaryReport(reportConfig report.ReportConfig, report report.ReportJson) (string, error) {
	blocks := buildSummaryReportBlocks(reportConfig, report)
	bytes, err := json.MarshalIndent(blocks, "", "  ")
	if err != nil {
		return "", err
	}
	log.Debug(string(bytes) + "\n")

	return "", nil
}

func (c *debugSlackNotifier) SendEnvironmentReport(respTimestamp string, env report.ReportEnvironment) error {
	attachments := buildEnvironmentReport(env)
	bytes, err := json.MarshalIndent(attachments, "", "  ")
	if err != nil {
		return err
	}
	log.Debug(string(bytes) + "\n")

	return nil
}

//-----------------------------------------------------------------------------------------

type slackNotifierConfig struct {
	channel  string
	client   *slack.Client
	username string
}

func (c *slackNotifierConfig) WithUsername(username string) *slackNotifierConfig {
	c.username = username
	return c
}

func NewNotifier(token string, channel string) *slackNotifierConfig {
	return &slackNotifierConfig{
		channel: channel,
		client:  slack.New(token),
	}
}

func (c *slackNotifierConfig) SendSummaryReport(reportConfig report.ReportConfig, report report.ReportJson) (string, error) {
	_, respTimestamp, err := c.client.PostMessageContext(
		context.TODO(),
		c.channel,
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionUsername(c.username),
		slack.MsgOptionBlocks(buildSummaryReportBlocks(reportConfig, report)...),
	)

	return respTimestamp, err
}

func (c *slackNotifierConfig) SendEnvironmentReport(respTimestamp string, env report.ReportEnvironment) error {
	_, _, err := c.client.PostMessageContext(
		context.TODO(),
		c.channel,
		slack.MsgOptionTS(respTimestamp),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionUsername(c.username),
		slack.MsgOptionAttachments(buildEnvironmentReport(env)...),
	)

	return err
}
