package handler

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"text/template"
	"time"

	"github.com/fluffyspangle/slackhub/pkg/markdown"
)

//go:embed templates
var templates embed.FS

type Poster interface {
	Post(ctx context.Context, reader io.Reader) error
}

type Handler struct {
	p Poster
}

func New(poster Poster) *Handler {
	return &Handler{
		p: poster,
	}
}

type EventContext interface {
	Channel() string
	Actor() string
	Name() string
	Action() string
	Event() any
	Branch() string
}

func (h *Handler) Handle(ec EventContext) error {
	tpl, err := template.New("").
		Delims("««", "»»").
		Funcs(template.FuncMap{
			"AsTimestamp":   AsTimestamp,
			"SlackMarkdown": SlackMarkdown,
			"ShortSHA":      ShortSHA,
		}).
		ParseFS(templates, fmt.Sprintf("templates/%s/%s.tmpl", ec.Name(), ec.Action()))
	if err != nil {
		return fmt.Errorf("could not instantiate template, %w", err)
	}

	out := bytes.NewBuffer(nil)

	if err := tpl.ExecuteTemplate(out, ec.Action()+".tmpl", ec); err != nil {
		return fmt.Errorf("could not execute template, %w", err)
	}

	ctx := context.Background()
	return h.p.Post(ctx, out)
}

func AsTimestamp(s string) int64 {
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		ts = time.Now().UTC()
	}
	return ts.Unix()
}

func SlackMarkdown(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	md, err := markdown.Parse(s)
	if err != nil {
		log.Println(">>SlackMarkdown:", err)
		return ""
	}
	return md
}

func ShortSHA(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}
