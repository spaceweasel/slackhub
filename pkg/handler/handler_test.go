package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/fluffyspangle/slackhub/pkg/handler"
)

func TestHandler_Handle(t *testing.T) {
	c := qt.New(t)

	poster := &MockPoster{}
	h := handler.New(poster)

	//ctx:= context.TODO()

	ec := createContext(c, "biscuits", "jeff", "pull_request")

	poster.PostFn = func(ctx context.Context, reader io.Reader) error {
		b, err := io.ReadAll(reader)
		c.Assert(err, qt.IsNil)
		c.Logf("%s", string(b))
		c.Assert(b, qt.HasLen, 3)
		return nil
	}
	err := h.Handle(ec)

	c.Assert(err, qt.IsNil)
}

func createContext(c *qt.C, channel, actor, eventname string) *testContext {
	f, err := os.Open(fmt.Sprintf("testdata/%s.json", eventname))
	c.Assert(err, qt.IsNil)
	defer f.Close()

	var event map[string]any
	err = json.NewDecoder(f).Decode(&event)
	c.Assert(err, qt.IsNil)

	return &testContext{
		channel:   channel,
		actor:     actor,
		eventName: eventname,
		event:     event,
	}
}

type testContext struct {
	channel   string
	actor     string
	eventName string // e.g. pull_request
	event     any    // ["action"] == "opened"
	branch    string
}

func (e *testContext) Channel() string {
	return e.channel
}
func (e *testContext) Actor() string {
	return e.actor
}
func (e *testContext) Name() string {
	return e.eventName
}
func (e *testContext) Action() string {
	if m, ok := e.event.(map[string]any); ok {
		if a, ok := m["action"]; ok {
			return a.(string)
		}
	}
	return "default"
}

func (e *testContext) Branch() string {
	return e.branch
}

func (e *testContext) Event() any {
	return e.event
}

type MockPoster struct {
	PostFn func(ctx context.Context, reader io.Reader) error
}

func (m *MockPoster) Post(ctx context.Context, reader io.Reader) error {
	if m.PostFn == nil {
		return nil
	}
	return m.PostFn(ctx, reader)
}
