package slacknotify

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"dsab.slacker/report"
)

func buildNamespaceReportHeader(ns report.Namespace) slack.Block {
	text := fmt.Sprintf("*Namespace:* %s", ns.Name)

	return slack.NewSectionBlock(
		nil,
		[]*slack.TextBlockObject{markdown(text)},
		nil,
	)
}

// buildEnvironmentHealthMessage builds the health message used in replies, for each environment
func buildEnvironmentHealthMessage(env report.ReportEnvironment) string {
	if env.IsHealthy() {
		return fmt.Sprintf(":white_check_mark: Healthy")
	} else {
		return fmt.Sprintf(":rotating_light: Unhealthy - %d issues", env.Errors())
	}
}

func buildSectionReport(s report.Section) []slack.Block {
	var (
		header = fmt.Sprintf("%s %s", s.Icon, s.Name)
		items  = strings.Join(s.Failures, "\n")
	)

	if len(s.Failures) == 0 {
		return []slack.Block{}
	}

	return []slack.Block{
		slack.NewContextBlock("", markdown(header)),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(markdown(items), nil, nil),
	}
}

func buildEnvironmentReportHeader(env report.ReportEnvironment) slack.Attachment {
	return slack.Attachment{
		Color:         attachmentColour(env),
		AuthorName:    "Environment",
		AuthorSubname: env.Name,
		Text:          buildEnvironmentHealthMessage(env),
	}
}

func buildEnvironmentReport(env report.ReportEnvironment) []slack.Attachment {
	var (
		attachmentColor = attachmentColour(env)
		attachments     = []slack.Attachment{}
	)

	attachments = append(attachments, buildEnvironmentReportHeader(env))

	if env.Status != report.Completed {
		return attachments
	}

	logger := log.WithField("environment", env.Name)
	logger.Debugf("Generating blocks for %d namespaces", len(env.Namespaces))

	for _, ns := range env.Namespaces {
		blocks := []slack.Block{}

		blocks = append(blocks, buildNamespaceReportHeader(ns))

		logger = logger.WithField("namespace", ns.Name)
		logger.Debugf("Generating blocks for %d sections", len(ns.Sections))

		for _, section := range ns.Sections {
			logger = logger.WithField("section", section.Name)
			logger.Debugf("Generating blocks for %d failures", len(section.Failures))
			blocks = append(blocks, buildSectionReport(section)...)
		}

		attachments = append(attachments, slack.Attachment{
			Color: attachmentColor,
			Blocks: slack.Blocks{
				BlockSet: blocks,
			},
		})
	}

	return attachments
}
