// Package shortcut provides a minimal HTTP client for the Shortcut API.
package shortcut

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// GetWorkflows lists every workflow in the workspace (and its states).
func (c *Client) GetWorkflows(ctx context.Context) ([]Workflow, error) {
	var ws []Workflow
	if err := c.get(ctx, "/workflows", &ws); err != nil {
		return nil, err
	}
	return ws, nil
}

// GetEpicWorkflow returns the workspace's epic workflow.
func (c *Client) GetEpicWorkflow(ctx context.Context) (*EpicWorkflow, error) {
	var ew EpicWorkflow
	if err := c.get(ctx, "/epic-workflow", &ew); err != nil {
		return nil, err
	}
	return &ew, nil
}

// GetMembers lists every workspace member.
func (c *Client) GetMembers(ctx context.Context) ([]Member, error) {
	var ms []Member
	if err := c.get(ctx, "/members", &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

// GetGroups lists every workspace group (team).
func (c *Client) GetGroups(ctx context.Context) ([]Group, error) {
	var gs []Group
	if err := c.get(ctx, "/groups", &gs); err != nil {
		return nil, err
	}
	return gs, nil
}

// SearchStories runs a Shortcut /search/stories query. We use this to
// map a branch name to its linked story when the branch itself doesn't
// contain `sc-NNNNN` (e.g. `branch:chore/sc-new-story/foo`). page_size
// is capped at 1 because callers only need the first match.
func (c *Client) SearchStories(ctx context.Context, query string) ([]Story, error) {
	path := "/search/stories?page_size=1&query=" + url.QueryEscape(query)
	var resp struct {
		Data []Story `json:"data"`
	}
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
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
