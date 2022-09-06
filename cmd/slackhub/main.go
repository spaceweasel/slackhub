package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sethvargo/go-githubactions"

	"github.com/spaceweasel/slackhub/pkg/config"
	"github.com/spaceweasel/slackhub/pkg/handler"
	"github.com/spaceweasel/slackhub/pkg/sender"
)

type EventContext struct {
	ctx       context.Context
	channel   string
	actor     string
	eventName string // e.g. pull_request
	event     any    // ["action"] == "opened"
	sha       string
}

func (e *EventContext) Context() context.Context {
	return e.ctx
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

func (e *EventContext) QualifiedAction() string {
	if m, ok := e.event.(map[string]any); ok {
		if a, ok := m["action"]; ok {
			return e.eventName + "." + a.(string)
		}
	}
	return e.eventName
}

func (e *EventContext) Action() string {
	action := e.Get("action")
	if action != nil {
		return action.(string)
	}
	// if m, ok := e.event.(map[string]any); ok {
	// 	if a, ok := m["action"]; ok {
	// 		return a.(string)
	// 	}
	// }
	return "default"
}

func (e *EventContext) Event() any {
	return e.event
}

func (e *EventContext) Branch() string {
	ref, _ := e.Get("ref").(string)
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref[11:]
	}

	return ""
}

func (e *EventContext) SHA() any {
	return e.sha
}

func (e *EventContext) Get(key string) any {
	return get(e.event, key)
}

func get(v any, key string) any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	ks := strings.Split(key, ".")
	mv := m[ks[0]]

	if len(ks) == 1 {
		return mv
	}

	return get(mv, strings.Join(ks[1:], "."))
}

func run(action *githubactions.Action) (err error) {
	cfg := config.New(action)
	poster := sender.NewPoster(cfg.Slack.Token)
	hdlr := handler.New(poster)

	defer func() {
		if err != nil {
			if cfg.FailOnError {
				action.Fatalf("%v", err)
			}
			action.Errorf("%v", err)
		}
	}()

	c, err := action.Context()
	if err != nil {
		return fmt.Errorf("failed to get action context, %v", err)
	}

	ec := &EventContext{
		channel:   cfg.Slack.Channel,
		actor:     c.Actor,
		eventName: c.EventName,
		event:     c.Event,
		ctx:       context.Background(),
		sha:       c.SHA,
	}

	// ignore any marshalling errors
	event, err := json.MarshalIndent(ec.event, "", "  ")
	if err == nil {
		action.Debugf("Event: %s", string(event))
	}

	if cfg.IgnoreActions[ec.QualifiedAction()] {
		action.Infof("Ignoring action: %s", ec.QualifiedAction())
		return nil
	}

	if NewEventFilter().Ignore(ec) {
		action.Infof("Filtering action: %s", ec.QualifiedAction())
		return nil
	}

	// Add pull_request.review_requested?

	err = hdlr.Handle(ec)
	return err
}

func main() {
	run(githubactions.New())
}

type EventFilter struct {
	cond []func(*EventContext) bool
}

func NewEventFilter() EventFilter {
	f := EventFilter{
		cond: []func(*EventContext) bool{
			func(ec *EventContext) bool {
				return ec.QualifiedAction() == "pull_request.opened" && ec.Get("draft") == true
			},
			func(ec *EventContext) bool {
				return ec.QualifiedAction() == "pull_request.closed" && ec.Get("merged") == false
			},
			func(ec *EventContext) bool {
				if ec.QualifiedAction() != "pull_request_review.submitted" {
					return false
				}
				if ec.Get("review.state") == "approved" {
					return false
				}

				body := ec.Get("review.body")
				return (body == nil || body == "")
			},
			func(ec *EventContext) bool {
				// ignore pushes unless to a branch (e.g. ignore tags)
				return ec.Name() != "push" && ec.Branch() == ""
			},
		},
	}

	return f
}

func (f EventFilter) Ignore(ec *EventContext) bool {
	for _, filter := range f.cond {
		if filter(ec) {
			return true
		}
	}

	return false
}
