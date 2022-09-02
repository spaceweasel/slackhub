package sender

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MessagePoster interface {
	PostMessage(ctx context.Context, channelID string, message []byte) error
}

type Poster struct {
	hc    *http.Client
	token string
}

type Option func(*Poster)

func WithHTTPClient(hc *http.Client) Option {
	return func(p *Poster) {
		p.hc = hc
	}
}

func NewPoster(token string, opts ...Option) *Poster {
	p := &Poster{
		hc: &http.Client{
			Timeout: 15 * time.Second,
		},
		token: token,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Poster) Post(ctx context.Context, reader io.Reader) error {
	const url = "https://slack.com/api/chat.postMessage"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := p.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Status response:", resp.StatusCode, resp.Status)
		return fmt.Errorf("bad code")
	}

	rb, err := io.ReadAll(resp.Body)
	fmt.Println("Response:", string(rb), err)

	// TODO: check for error response
	// {
	//     "ok": false,
	//     "error": "too_many_attachments"
	// }

	return nil
}
