package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SubscriberUpdate struct {
	Email  string `json:"email"`
	Status string `json:"status"`
}

type Statement struct {
	Creator           string             `json:"creator"`
	Period            string             `json:"period"`
	DigitalAssets     []string           `json:"digital_assets"`
	SubscriberUpdates []SubscriberUpdate `json:"subscriber_updates"`
	ContentItems      int                `json:"content_items"`
}

func (s Statement) Exportable() bool {
	return s.Creator != "" && s.Period != "" && len(s.DigitalAssets) > 0 && len(s.SubscriberUpdates) > 0 && s.ContentItems > 0
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiError struct {
	Status int
	Detail string
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("infrai request rejected (%d): %s", e.Status, e.Detail)
}

type Client struct {
	baseURL string
	key     string
	http    *http.Client
}

func NewClient(baseURL, key string, transport http.RoundTripper) *Client {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), key: key, http: &http.Client{Transport: transport}}
}

func (c *Client) post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.requestURL(path), strings.NewReader(string(payload)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		if err := json.Unmarshal(bodyBytes, &env); err != nil {
			return nil, fmt.Errorf("decode envelope: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return nil, &InfraiError{Status: resp.StatusCode, Detail: string(env.Error)}
		}
		if resp.StatusCode >= 500 {
			return nil, &InfraiError{Status: resp.StatusCode, Detail: string(env.Error)}
		}
		return env.Data, nil
	}
	return nil, errors.New("request retry budget exhausted")
}

func (c *Client) Generate(ctx context.Context, markdown string) (json.RawMessage, error) {
	return c.post(ctx, "/v1/pdf/generate", map[string]any{"markdown": markdown})
}

func (c *Client) UsageTimeseries(ctx context.Context, period string) (json.RawMessage, error) {
	return c.get(ctx, "/v1/account/usage/timeseries?period="+period)
}

func (c *Client) get(ctx context.Context, path string) (json.RawMessage, error) {
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.requestURL(path), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		bytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var env envelope
		if err := json.Unmarshal(bytes, &env); err != nil {
			return nil, fmt.Errorf("decode envelope: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			continue
		}
		if !env.OK {
			return nil, &InfraiError{Status: resp.StatusCode, Detail: string(env.Error)}
		}
		return env.Data, nil
	}
	return nil, errors.New("request retry budget exhausted")
}

func (c *Client) requestURL(path string) string {
	if strings.HasSuffix(c.baseURL, "/v1") && strings.HasPrefix(path, "/v1/") {
		return c.baseURL + strings.TrimPrefix(path, "/v1")
	}
	return c.baseURL + path
}

func renderMarkdown(s Statement, usage json.RawMessage) string {
	return fmt.Sprintf("# Creator account statement\n\nCreator: %s\nPeriod: %s\n\n## Digital assets\n- %s\n\n## Subscriber updates\n- %d updates\n\n## Content processing\n- %d content items processed\n\n## Usage\n%s\n", s.Creator, s.Period, strings.Join(s.DigitalAssets, "\n- "), len(s.SubscriberUpdates), s.ContentItems, string(usage))
}
