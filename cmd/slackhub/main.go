package main

import (
	"context"

	"github.com/sethvargo/go-githubactions"

	"github.com/spaceweasel/slackhub/pkg/handler"
	"github.com/spaceweasel/slackhub/pkg/sender"
)

type EventContext struct {
	channel   string
	actor     string
	eventName string // e.g. pull_request
	event     any    // ["action"] == "opened"
}

func (e *EventContext) Channel() string {
	return e.channel
}
func (e *EventContext) Actor() string {
	return e.actor
}
func (e *EventContext) Name() string {
	return e.eventName
}
func (e *EventContext) Action() string {
	if m, ok := e.event.(map[string]any); ok {
		if a, ok := m["action"]; ok {
			return a.(string)
		}
	}
	return "default"
}

func (e *EventContext) Event() any {
	return e.event
}

func run(action *githubactions.Action) error {
	action.AddMask("SLACK_BOT_TOKEN")
	token := action.Getenv("SLACK_BOT_TOKEN")
	poster := sender.NewPoster(token)
	hdlr := handler.New(poster)

	ctx := context.Background()
	_ = ctx
	c, err := action.Context()
	if err != nil {
		return err
	}

	ec := &EventContext{
		channel:   action.GetInput("channel"),
		actor:     c.Actor,
		eventName: c.EventName,
		event:     c.Event,
		//ctx
		// sha?
	}

	return hdlr.Handle(ec)
}

func main() {
	action := githubactions.New()

	err := run(action)
	if err != nil {
		action.Fatalf("%v", err)
	}
}
