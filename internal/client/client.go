package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client wraps the video-platform REST API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new video platform client
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Session represents a recording session
type Session struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	SessionType          string  `json:"session_type"`
	Status               string  `json:"status"`
	ScheduledStart       *string `json:"scheduled_start,omitempty"`
	ActualStart          *string `json:"actual_start,omitempty"`
	ActualEnd            *string `json:"actual_end,omitempty"`
	Opponent             *string `json:"opponent,omitempty"`
	Location             *string `json:"location,omitempty"`
	ClipCount            int     `json:"clip_count"`
	TagCount             int     `json:"tag_count"`
	TotalDurationSeconds int     `json:"total_duration_seconds"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// Clip represents a video clip
type Clip struct {
	ID              string  `json:"id"`
	SessionID       string  `json:"session_id"`
	ChannelID       string  `json:"channel_id"`
	Title           *string `json:"title,omitempty"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	DurationSeconds float64 `json:"duration_seconds"`
	Status          string  `json:"status"`
	IsFavorite      bool    `json:"is_favorite"`
	ViewCount       int     `json:"view_count"`
	CreatedAt       string  `json:"created_at"`
}

// Channel represents a video input channel
type Channel struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	InputType    *string `json:"input_type,omitempty"`
	InputURL     *string `json:"input_url,omitempty"`
	Resolution   *string `json:"resolution,omitempty"`
	Framerate    *int    `json:"framerate,omitempty"`
	Status       string  `json:"status"`
	LastSeenAt   *string `json:"last_seen_at,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// Tag represents a clip annotation
type Tag struct {
	ID          string   `json:"id"`
	ClipID      string   `json:"clip_id"`
	SessionID   string   `json:"session_id"`
	Quarter     *int     `json:"quarter,omitempty"`
	Down        *int     `json:"down,omitempty"`
	Distance    *int     `json:"distance,omitempty"`
	PlayType    *string  `json:"play_type,omitempty"`
	Formation   *string  `json:"formation,omitempty"`
	Result      *string  `json:"result,omitempty"`
	YardsGained *int     `json:"yards_gained,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	IsImportant bool     `json:"is_important"`
	IsReviewed  bool     `json:"is_reviewed"`
	CreatedAt   string   `json:"created_at"`
}

// PaginatedResponse wraps paginated API responses
type PaginatedResponse[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ListSessionsParams for filtering sessions
type ListSessionsParams struct {
	Status      string
	SessionType string
	Limit       int
	Offset      int
}

// ListSessions returns all sessions
func (c *Client) ListSessions(ctx context.Context, params ListSessionsParams) (*PaginatedResponse[Session], error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.SessionType != "" {
		query.Set("session_type", params.SessionType)
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", params.Offset))
	}

	var resp PaginatedResponse[Session]
	if err := c.get(ctx, "/api/v1/sessions", query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSession returns a single session
func (c *Client) GetSession(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := c.get(ctx, "/api/v1/sessions/"+id, nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// CreateSessionRequest for creating a session
type CreateSessionRequest struct {
	Name           string  `json:"name"`
	SessionType    string  `json:"session_type"`
	ScheduledStart *string `json:"scheduled_start,omitempty"`
	Opponent       *string `json:"opponent,omitempty"`
	Location       *string `json:"location,omitempty"`
}

// CreateSession creates a new session
func (c *Client) CreateSession(ctx context.Context, req CreateSessionRequest) (*Session, error) {
	var session Session
	if err := c.post(ctx, "/api/v1/sessions", req, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// StartSession starts a session
func (c *Client) StartSession(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := c.post(ctx, "/api/v1/sessions/"+id+"/start", nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// PauseSession pauses a session
func (c *Client) PauseSession(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := c.post(ctx, "/api/v1/sessions/"+id+"/pause", nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// CompleteSession completes a session
func (c *Client) CompleteSession(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := c.post(ctx, "/api/v1/sessions/"+id+"/complete", nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// ListClipsParams for filtering clips
type ListClipsParams struct {
	SessionID string
	ChannelID string
	Status    string
	Favorite  *bool
	Search    string
	Limit     int
	Offset    int
}

// ListClips returns clips with filters
func (c *Client) ListClips(ctx context.Context, params ListClipsParams) (*PaginatedResponse[Clip], error) {
	query := url.Values{}
	if params.SessionID != "" {
		query.Set("session_id", params.SessionID)
	}
	if params.ChannelID != "" {
		query.Set("channel_id", params.ChannelID)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Favorite != nil {
		query.Set("favorite", fmt.Sprintf("%v", *params.Favorite))
	}
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", params.Offset))
	}

	var resp PaginatedResponse[Clip]
	if err := c.get(ctx, "/api/v1/clips", query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetClip returns a single clip
func (c *Client) GetClip(ctx context.Context, id string) (*Clip, error) {
	var clip Clip
	if err := c.get(ctx, "/api/v1/clips/"+id, nil, &clip); err != nil {
		return nil, err
	}
	return &clip, nil
}

// FavoriteClip toggles favorite status
func (c *Client) FavoriteClip(ctx context.Context, id string) (*Clip, error) {
	var clip Clip
	if err := c.post(ctx, "/api/v1/clips/"+id+"/favorite", nil, &clip); err != nil {
		return nil, err
	}
	return &clip, nil
}

// ListChannels returns all channels
func (c *Client) ListChannels(ctx context.Context) (*PaginatedResponse[Channel], error) {
	var resp PaginatedResponse[Channel]
	if err := c.get(ctx, "/api/v1/channels", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ActivateChannel activates a channel
func (c *Client) ActivateChannel(ctx context.Context, id string) (*Channel, error) {
	var channel Channel
	if err := c.post(ctx, "/api/v1/channels/"+id+"/activate", nil, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// DeactivateChannel deactivates a channel
func (c *Client) DeactivateChannel(ctx context.Context, id string) (*Channel, error) {
	var channel Channel
	if err := c.post(ctx, "/api/v1/channels/"+id+"/deactivate", nil, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

// ListTagsParams for filtering tags
type ListTagsParams struct {
	SessionID   string
	ClipID      string
	PlayType    string
	IsImportant *bool
	IsReviewed  *bool
	Limit       int
	Offset      int
}

// ListTags returns tags with filters
func (c *Client) ListTags(ctx context.Context, params ListTagsParams) (*PaginatedResponse[Tag], error) {
	query := url.Values{}
	if params.SessionID != "" {
		query.Set("session_id", params.SessionID)
	}
	if params.ClipID != "" {
		query.Set("clip_id", params.ClipID)
	}
	if params.PlayType != "" {
		query.Set("play_type", params.PlayType)
	}
	if params.IsImportant != nil {
		query.Set("is_important", fmt.Sprintf("%v", *params.IsImportant))
	}
	if params.IsReviewed != nil {
		query.Set("is_reviewed", fmt.Sprintf("%v", *params.IsReviewed))
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	var resp PaginatedResponse[Tag]
	if err := c.get(ctx, "/api/v1/tags", query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateTagRequest for creating a tag
type CreateTagRequest struct {
	ClipID      string   `json:"clip_id"`
	SessionID   string   `json:"session_id"`
	Quarter     *int     `json:"quarter,omitempty"`
	Down        *int     `json:"down,omitempty"`
	Distance    *int     `json:"distance,omitempty"`
	PlayType    *string  `json:"play_type,omitempty"`
	Formation   *string  `json:"formation,omitempty"`
	Result      *string  `json:"result,omitempty"`
	YardsGained *int     `json:"yards_gained,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
}

// CreateTag creates a new tag
func (c *Client) CreateTag(ctx context.Context, req CreateTagRequest) (*Tag, error) {
	var tag Tag
	if err := c.post(ctx, "/api/v1/tags", req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// HTTP helpers

func (c *Client) get(ctx context.Context, path string, query url.Values, result interface{}) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
