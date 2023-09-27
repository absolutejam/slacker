package slacknotify

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"dsab.slacker/report"
)

const (
	BRING_UP_HEALTHCHECK             = "bring_up_healthcheck"
	BRING_UP_HEALTHCHECK_ENVIRONMENT = "bring_up_healthcheck_environment"
)

//-----------------------------------------------------------------------------------------

type ResponseTimestamp struct {
	Ts string
}

func NewResponseTimestamp(ts string) ResponseTimestamp {
	return ResponseTimestamp{
		Ts: ts,
	}
}

func (rt *ResponseTimestamp) IsEmpty() bool {
	return rt.Ts == ""
}

//-----------------------------------------------------------------------------------------

type SlackNotifier interface {
	SendSummaryReport(report report.ReportJson, updateMessageTs *ResponseTimestamp) (summaryReportTs ResponseTimestamp, err error)
	SendEnvironmentReport(parentMessageTs ResponseTimestamp, env report.ReportEnvironment, updateMessageTs *ResponseTimestamp) error
}

// Interface assertions
var (
	_ SlackNotifier = (*slackNotifierConfig)(nil)
	_ SlackNotifier = (*debugNotifier)(nil)
)

//-----------------------------------------------------------------------------------------
// Live

type slackNotifierConfig struct {
	channel      string
	reportConfig report.ReportConfig
	client       *slack.Client
	username     string
}

func (c *slackNotifierConfig) WithUsername(username string) *slackNotifierConfig {
	c.username = username
	return c
}

func NewNotifier(token string, channel string, reportConfig report.ReportConfig) *slackNotifierConfig {
	return &slackNotifierConfig{
		channel:      channel,
		client:       slack.New(token),
		reportConfig: reportConfig,
	}
}

func (c *slackNotifierConfig) SendSummaryReport(report report.ReportJson, updateMessageTs *ResponseTimestamp) (summaryReportTs ResponseTimestamp, err error) {
	var respTimestamp string

	opts := []slack.MsgOption{
		slack.MsgOptionMetadata(slack.SlackMetadata{
			EventType: BRING_UP_HEALTHCHECK,
			EventPayload: map[string]interface{}{
				"date": c.reportConfig.ReportDate,
			},
		}),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionUsername(c.username),
		slack.MsgOptionBlocks(buildSummaryReportBlocks(c.reportConfig, report)...),
	}

	if updateMessageTs != nil {
		log.Debug("Updating existing summary report")
		_, respTimestamp, _, err = c.client.UpdateMessageContext(
			context.TODO(),
			c.channel,
			updateMessageTs.Ts,
			opts...,
		)
	} else {
		log.Debug("Creating new summary report")
		_, respTimestamp, err = c.client.PostMessageContext(
			context.TODO(),
			c.channel,
			opts...,
		)
	}
	return NewResponseTimestamp(respTimestamp), err
}

func (c *slackNotifierConfig) SendEnvironmentReport(parentMessageTs ResponseTimestamp, env report.ReportEnvironment, updateMessageTs *ResponseTimestamp) (err error) {
	opts := []slack.MsgOption{
		slack.MsgOptionTS(parentMessageTs.Ts),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionMetadata(slack.SlackMetadata{
			EventType: BRING_UP_HEALTHCHECK_ENVIRONMENT,
			EventPayload: map[string]interface{}{
				"environment": env.Name,
			},
		}),
		slack.MsgOptionUsername(c.username),
		slack.MsgOptionAttachments(buildEnvironmentReport(env)...),
	}

	if updateMessageTs != nil {
		log.WithField("env", env.Name).Debug("Updating existing environment report")
		_, _, _, err = c.client.UpdateMessageContext(
			context.TODO(),
			c.channel,
			updateMessageTs.Ts,
			opts...,
		)

	} else {
		log.WithField("env", env.Name).Debug("Creating new environment report")
		_, _, err = c.client.PostMessageContext(
			context.TODO(),
			c.channel,
			opts...,
		)
	}

	return err
}

//-----------------------------------------------------------------------------------------
// Debug

type debugNotifier struct {
	reportConfig report.ReportConfig
}

func NewDebugNotifier() *debugNotifier {
	return &debugNotifier{}
}

func (c *debugNotifier) SendSummaryReport(report report.ReportJson, updateMessageTs *ResponseTimestamp) (summaryReportTs ResponseTimestamp, err error) {
	blocks := buildSummaryReportBlocks(c.reportConfig, report)
	bytes, err := json.MarshalIndent(blocks, "", "  ")
	if err != nil {
		return NewResponseTimestamp(""), err
	}
	log.Debug(string(bytes) + "\n")

	return NewResponseTimestamp("placeholder"), nil
}

func (c *debugNotifier) SendEnvironmentReport(parentMessageTs ResponseTimestamp, env report.ReportEnvironment, updateMessageTs *ResponseTimestamp) error {
	attachments := buildEnvironmentReport(env)
	bytes, err := json.MarshalIndent(attachments, "", "  ")
	if err != nil {
		return err
	}
	log.Debug(string(bytes) + "\n")

	return nil
}
