package config

import (
	"strings"

	"github.com/sethvargo/go-githubactions"
)

type Config struct {
	Slack struct {
		Token   string
		Channel string
	}
	FailOnError     bool
	DumpEvent       bool
	PretextOverride string
	IgnoreActions   map[string]bool
	Log             Logger
}

func New(action *githubactions.Action) *Config {
	action.AddMask("SLACK_BOT_TOKEN")
	cfg := &Config{
		Slack: struct {
			Token   string
			Channel string
		}{
			Token:   action.Getenv("SLACK_BOT_TOKEN"),
			Channel: action.GetInput("channel"),
		},
		FailOnError:   strings.EqualFold(action.GetInput("fail_on_error"), "true"),
		DumpEvent:     strings.EqualFold(action.GetInput("dump_event"), "true"),
		IgnoreActions: strToMap(action.GetInput("ignore_actions")),
		Log: logger{
			failOnErr: strings.EqualFold(action.GetInput("fail_on_error"), "true"),
			l:         action,
		},
	}

	return cfg
}

func strToMap(s string) map[string]bool {
	m := make(map[string]bool)
	if s == "" {
		return m
	}
	if !strings.HasPrefix(s, "[") {
		// single element
		m[strings.TrimSpace(s)] = true
		return m
	}
	s = strings.Trim(s, "[]")
	els := strings.Split(s, ",")
	for _, el := range els {
		m[strings.TrimSpace(el)] = true
	}

	return m
}

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Noticef(msg string, args ...any)
	Warningf(msg string, args ...any)
	Errorf(msg string, args ...any)
	Fatalf(msg string, args ...any)
}

type logger struct {
	l         Logger
	failOnErr bool
}

func (l logger) Debugf(msg string, args ...any) {
	l.l.Debugf(msg, args...)
}

func (l logger) Infof(msg string, args ...any) {
	l.l.Infof(msg, args...)
}

func (l logger) Noticef(msg string, args ...any) {
	l.l.Noticef(msg, args...)
}

func (l logger) Warningf(msg string, args ...any) {
	l.l.Warningf(msg, args...)
}

func (l logger) Errorf(msg string, args ...any) {
	l.l.Errorf(msg, args...)
}

func (l logger) Fatalf(msg string, args ...any) {
	if l.failOnErr {
		l.l.Fatalf(msg, args...)
	}
	l.l.Errorf(msg, args...)
}
