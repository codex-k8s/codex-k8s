package service

import "context"

// CallbackSink forwards normalized Telegram callbacks to the platform-owned semantic layer.
type CallbackSink interface {
	Submit(context.Context, CallbackEnvelope) (CallbackOutcome, error)
}

type messageLinks struct {
	RunURL         string
	IssueURL       string
	PullRequestURL string
}

type notifyMessageData struct {
	Summary         string
	DetailsMarkdown string
	Links           messageLinks
}

type decisionMessageData struct {
	Question         string
	DetailsMarkdown  string
	ReplyInstruction string
	Links            messageLinks
}
