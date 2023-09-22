package slacknotify

import "github.com/slack-go/slack"

func markdown(markdown string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.MarkdownType, markdown, false, false)
}

func plaintext(markdown string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.PlainTextType, markdown, true, false)
}

func linkButton(text string, url string) *slack.ButtonBlockElement {
	return &slack.ButtonBlockElement{
		Type: slack.METButton,
		Text: plaintext(text),
		URL:  url,
	}
}
