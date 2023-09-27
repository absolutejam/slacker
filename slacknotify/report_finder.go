package slacknotify

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type ReportFinder interface {
	FindReport(date string) (*ResponseTimestamp, error)
	FindEnvironmentReport(environment string, responseTs ResponseTimestamp) (*ResponseTimestamp, error)
}

// Interface assertions
var (
	_ ReportFinder = (*slackReportFinder)(nil)
	_ ReportFinder = (*noOpReportFinder)(nil)
)

//-----------------------------------------------------------------------------------------

type slackReportFinder struct {
	channel string
	client  *slack.Client
}

func NewSlackReportFinder(token string, channel string) *slackReportFinder {
	return &slackReportFinder{
		channel: channel,
		client:  slack.New(token),
	}
}

func (s *slackReportFinder) FindReport(date string) (*ResponseTimestamp, error) {
	msgs, err := s.client.GetConversationHistoryContext(context.Background(),
		&slack.GetConversationHistoryParameters{
			ChannelID:          s.channel,
			IncludeAllMetadata: true,
			Limit:              1000,
		})
	if err != nil {
		return nil, fmt.Errorf("error getting conversations: %s", err)
	}

	for _, msg := range msgs.Messages {
		if msg.Metadata.EventType == BRING_UP_HEALTHCHECK &&
			msg.Metadata.EventPayload != nil &&
			msg.Metadata.EventPayload["date"] == date {
			responseTs := NewResponseTimestamp(msg.Timestamp)

			return &responseTs, nil
		}
	}

	return nil, nil
}

func (s *slackReportFinder) FindEnvironmentReport(environment string, responseTs ResponseTimestamp) (*ResponseTimestamp, error) {
	msgs, _, _, err := s.client.GetConversationRepliesContext(context.Background(),
		&slack.GetConversationRepliesParameters{
			ChannelID:          s.channel,
			Timestamp:          responseTs.Ts,
			IncludeAllMetadata: true,
			Limit:              1000,
		})
	if err != nil {
		return nil, fmt.Errorf("error getting conversations: %s", err)
	}

	for _, msg := range msgs {
		if msg.Metadata.EventType == BRING_UP_HEALTHCHECK_ENVIRONMENT &&
			msg.Metadata.EventPayload != nil &&
			msg.Metadata.EventPayload["environment"] == environment {
			responseTs := NewResponseTimestamp(msg.Timestamp)

			return &responseTs, nil
		}
	}

	return nil, nil
}

//------------------------------------------------------------------------------

type noOpReportFinder struct{}

func NewNoOpReportFinder() *noOpReportFinder {
	return &noOpReportFinder{}
}

func (s *noOpReportFinder) FindReport(date string) (*ResponseTimestamp, error) {
	return nil, nil
}

func (s *noOpReportFinder) FindEnvironmentReport(environment string, responseTs ResponseTimestamp) (*ResponseTimestamp, error) {
	return nil, nil
}
