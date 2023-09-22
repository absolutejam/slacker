package slacknotify

import (
	"fmt"

	"github.com/slack-go/slack"

	"dsab.slacker/report"
)

// buildSummaryHealthMessage builds the message that is used in the top-level summary message,
// describing each environment and its health & errors
func buildSummaryHealthMessage(env report.ReportEnvironment) *slack.TextBlockObject {
	var msg string

	switch env.Status {
	case report.Pending:
		msg = fmt.Sprintf(":hourglass: *%s* | Pending", env.Name)
	case report.Completed:
		if env.IsHealthy() {
			msg = fmt.Sprintf(":white_check_mark: *%s* | Healthy", env.Name)
		} else {
			msg = fmt.Sprintf(":rotating_light: *%s* | Unhealthy - %d issues", env.Name, env.Errors())
		}
	default:
		msg = fmt.Sprintf(":x: *%s* | Unknown failure", env.Name)
	}

	return markdown(msg)
}

// attachmentColour builds a colour Hex code used to colour Slack attachment messages based
// on the health of the environment
func attachmentColour(env report.ReportEnvironment) string {
	if env.IsHealthy() {
		return "#00FF00"
	} else {
		return "#FF0000"
	}
}

func buildEnvironmentSummarySection(reportConfig report.ReportConfig, env report.ReportEnvironment) *slack.SectionBlock {
	var button *slack.Accessory

	button = slack.NewAccessory(
		&slack.ButtonBlockElement{
			Type: slack.METButton,
			Text: plaintext(":clipboard: See report"),
			URL:  buildReportUrl(reportConfig, env),
		},
	)

	return slack.NewSectionBlock(
		buildSummaryHealthMessage(env),
		nil,
		button,
	)
}
