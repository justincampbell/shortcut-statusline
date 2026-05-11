// Package shortcut provides a minimal HTTP client for the Shortcut API.
package shortcut

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the public Shortcut API v3 base URL.
const DefaultBaseURL = "https://api.app.shortcut.com/api/v3"

// DefaultTimeout keeps the statusline snappy when the API is slow.
const DefaultTimeout = 3 * time.Second

// Client talks to the Shortcut API.
type Client struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

// New returns a client with sensible defaults.
func New(token string) *Client {
	return &Client{
		Token:   token,
		BaseURL: DefaultBaseURL,
		HTTP:    &http.Client{Timeout: DefaultTimeout},
	}
}

// GetStory fetches a story by ID.
func (c *Client) GetStory(ctx context.Context, id int) (*Story, error) {
	var s Story
	if err := c.get(ctx, fmt.Sprintf("/stories/%d", id), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetEpic fetches an epic by ID.
func (c *Client) GetEpic(ctx context.Context, id int) (*Epic, error) {
	var e Epic
	if err := c.get(ctx, fmt.Sprintf("/epics/%d", id), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// GetObjective fetches an objective (milestone) by ID.
func (c *Client) GetObjective(ctx context.Context, id int) (*Objective, error) {
	var o Objective
	if err := c.get(ctx, fmt.Sprintf("/objectives/%d", id), &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Shortcut-Token", c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("shortcut API %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
