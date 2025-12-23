package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Prodro21/video-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools adds all tool handlers to the server
func RegisterTools(s *server.MCPServer, c *client.Client) {
	// Session tools
	s.AddTool(mcp.Tool{
		Name:        "list_sessions",
		Description: "List recording sessions with optional filters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by status",
					"enum":        []string{"scheduled", "active", "paused", "completed", "archived"},
				},
				"session_type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by session type",
					"enum":        []string{"game", "practice", "scrimmage", "training", "other"},
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default 20)",
				},
			},
		},
	}, makeListSessions(c))

	s.AddTool(mcp.Tool{
		Name:        "create_session",
		Description: "Create a new recording session",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Session name",
				},
				"session_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of session",
					"enum":        []string{"game", "practice", "scrimmage", "training", "other"},
				},
				"opponent": map[string]interface{}{
					"type":        "string",
					"description": "Opponent name (for games)",
				},
				"location": map[string]interface{}{
					"type":        "string",
					"description": "Location of the session",
				},
			},
			Required: []string{"name", "session_type"},
		},
	}, makeCreateSession(c))

	s.AddTool(mcp.Tool{
		Name:        "start_session",
		Description: "Start a scheduled session to begin recording",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the session to start",
				},
			},
			Required: []string{"session_id"},
		},
	}, makeStartSession(c))

	s.AddTool(mcp.Tool{
		Name:        "pause_session",
		Description: "Pause an active recording session",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the session to pause",
				},
			},
			Required: []string{"session_id"},
		},
	}, makePauseSession(c))

	s.AddTool(mcp.Tool{
		Name:        "complete_session",
		Description: "Complete and finalize a recording session",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the session to complete",
				},
			},
			Required: []string{"session_id"},
		},
	}, makeCompleteSession(c))

	// Clip tools
	s.AddTool(mcp.Tool{
		Name:        "list_clips",
		Description: "List video clips with optional filters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by session ID",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by clip status",
					"enum":        []string{"pending", "processing", "ready", "failed"},
				},
				"favorite": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by favorite status",
				},
				"search": map[string]interface{}{
					"type":        "string",
					"description": "Search clips by title",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default 20)",
				},
			},
		},
	}, makeListClips(c))

	s.AddTool(mcp.Tool{
		Name:        "favorite_clip",
		Description: "Toggle favorite status on a clip",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"clip_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the clip",
				},
			},
			Required: []string{"clip_id"},
		},
	}, makeFavoriteClip(c))

	// Channel tools
	s.AddTool(mcp.Tool{
		Name:        "list_channels",
		Description: "List all video input channels and their status",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}, makeListChannels(c))

	s.AddTool(mcp.Tool{
		Name:        "activate_channel",
		Description: "Activate a video input channel",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"channel_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the channel to activate",
				},
			},
			Required: []string{"channel_id"},
		},
	}, makeActivateChannel(c))

	s.AddTool(mcp.Tool{
		Name:        "deactivate_channel",
		Description: "Deactivate a video input channel",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"channel_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the channel to deactivate",
				},
			},
			Required: []string{"channel_id"},
		},
	}, makeDeactivateChannel(c))

	// Tag tools
	s.AddTool(mcp.Tool{
		Name:        "list_tags",
		Description: "List clip tags/annotations with filters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by session ID",
				},
				"clip_id": map[string]interface{}{
					"type":        "string",
					"description": "Filter by clip ID",
				},
				"play_type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by play type",
				},
				"is_important": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter important tags only",
				},
				"is_reviewed": map[string]interface{}{
					"type":        "boolean",
					"description": "Filter by review status",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default 50)",
				},
			},
		},
	}, makeListTags(c))

	s.AddTool(mcp.Tool{
		Name:        "create_tag",
		Description: "Create a new tag/annotation for a clip",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"clip_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the clip to tag",
				},
				"session_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the session",
				},
				"play_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of play (Run, Pass, Punt, etc.)",
				},
				"formation": map[string]interface{}{
					"type":        "string",
					"description": "Formation used",
				},
				"result": map[string]interface{}{
					"type":        "string",
					"description": "Result of the play",
				},
				"down": map[string]interface{}{
					"type":        "integer",
					"description": "Down number (1-4)",
				},
				"distance": map[string]interface{}{
					"type":        "integer",
					"description": "Yards to go",
				},
				"yards_gained": map[string]interface{}{
					"type":        "integer",
					"description": "Yards gained on the play",
				},
				"notes": map[string]interface{}{
					"type":        "string",
					"description": "Additional notes",
				},
			},
			Required: []string{"clip_id", "session_id"},
		},
	}, makeCreateTag(c))
}

// Tool handler factories

func makeListSessions(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := client.ListSessionsParams{Limit: 20}

		if status, ok := req.Params.Arguments["status"].(string); ok {
			params.Status = status
		}
		if sessionType, ok := req.Params.Arguments["session_type"].(string); ok {
			params.SessionType = sessionType
		}
		if limit, ok := req.Params.Arguments["limit"].(float64); ok {
			params.Limit = int(limit)
		}

		resp, err := c.ListSessions(ctx, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list sessions: %v", err)), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeCreateSession(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := req.Params.Arguments["name"].(string)
		sessionType, _ := req.Params.Arguments["session_type"].(string)

		if name == "" || sessionType == "" {
			return mcp.NewToolResultError("name and session_type are required"), nil
		}

		createReq := client.CreateSessionRequest{
			Name:        name,
			SessionType: sessionType,
		}

		if opponent, ok := req.Params.Arguments["opponent"].(string); ok {
			createReq.Opponent = &opponent
		}
		if location, ok := req.Params.Arguments["location"].(string); ok {
			createReq.Location = &location
		}

		session, err := c.CreateSession(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create session: %v", err)), nil
		}

		data, _ := json.MarshalIndent(session, "", "  ")
		return mcp.NewToolResultText(fmt.Sprintf("Session created successfully:\n%s", string(data))), nil
	}
}

func makeStartSession(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, _ := req.Params.Arguments["session_id"].(string)
		if sessionID == "" {
			return mcp.NewToolResultError("session_id is required"), nil
		}

		session, err := c.StartSession(ctx, sessionID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to start session: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Session '%s' started successfully. Status: %s", session.Name, session.Status)), nil
	}
}

func makePauseSession(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, _ := req.Params.Arguments["session_id"].(string)
		if sessionID == "" {
			return mcp.NewToolResultError("session_id is required"), nil
		}

		session, err := c.PauseSession(ctx, sessionID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to pause session: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Session '%s' paused. Status: %s", session.Name, session.Status)), nil
	}
}

func makeCompleteSession(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID, _ := req.Params.Arguments["session_id"].(string)
		if sessionID == "" {
			return mcp.NewToolResultError("session_id is required"), nil
		}

		session, err := c.CompleteSession(ctx, sessionID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to complete session: %v", err)), nil
		}

		data, _ := json.MarshalIndent(session, "", "  ")
		return mcp.NewToolResultText(fmt.Sprintf("Session completed:\n%s", string(data))), nil
	}
}

func makeListClips(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := client.ListClipsParams{Limit: 20}

		if sessionID, ok := req.Params.Arguments["session_id"].(string); ok {
			params.SessionID = sessionID
		}
		if status, ok := req.Params.Arguments["status"].(string); ok {
			params.Status = status
		}
		if favorite, ok := req.Params.Arguments["favorite"].(bool); ok {
			params.Favorite = &favorite
		}
		if search, ok := req.Params.Arguments["search"].(string); ok {
			params.Search = search
		}
		if limit, ok := req.Params.Arguments["limit"].(float64); ok {
			params.Limit = int(limit)
		}

		resp, err := c.ListClips(ctx, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list clips: %v", err)), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeFavoriteClip(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		clipID, _ := req.Params.Arguments["clip_id"].(string)
		if clipID == "" {
			return mcp.NewToolResultError("clip_id is required"), nil
		}

		clip, err := c.FavoriteClip(ctx, clipID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to toggle favorite: %v", err)), nil
		}

		status := "added to"
		if !clip.IsFavorite {
			status = "removed from"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Clip %s favorites", status)), nil
	}
}

func makeListChannels(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := c.ListChannels(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list channels: %v", err)), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeActivateChannel(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, _ := req.Params.Arguments["channel_id"].(string)
		if channelID == "" {
			return mcp.NewToolResultError("channel_id is required"), nil
		}

		channel, err := c.ActivateChannel(ctx, channelID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to activate channel: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Channel '%s' activated. Status: %s", channel.Name, channel.Status)), nil
	}
}

func makeDeactivateChannel(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelID, _ := req.Params.Arguments["channel_id"].(string)
		if channelID == "" {
			return mcp.NewToolResultError("channel_id is required"), nil
		}

		channel, err := c.DeactivateChannel(ctx, channelID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to deactivate channel: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Channel '%s' deactivated. Status: %s", channel.Name, channel.Status)), nil
	}
}

func makeListTags(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := client.ListTagsParams{Limit: 50}

		if sessionID, ok := req.Params.Arguments["session_id"].(string); ok {
			params.SessionID = sessionID
		}
		if clipID, ok := req.Params.Arguments["clip_id"].(string); ok {
			params.ClipID = clipID
		}
		if playType, ok := req.Params.Arguments["play_type"].(string); ok {
			params.PlayType = playType
		}
		if isImportant, ok := req.Params.Arguments["is_important"].(bool); ok {
			params.IsImportant = &isImportant
		}
		if isReviewed, ok := req.Params.Arguments["is_reviewed"].(bool); ok {
			params.IsReviewed = &isReviewed
		}
		if limit, ok := req.Params.Arguments["limit"].(float64); ok {
			params.Limit = int(limit)
		}

		resp, err := c.ListTags(ctx, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tags: %v", err)), nil
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func makeCreateTag(c *client.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		clipID, _ := req.Params.Arguments["clip_id"].(string)
		sessionID, _ := req.Params.Arguments["session_id"].(string)

		if clipID == "" || sessionID == "" {
			return mcp.NewToolResultError("clip_id and session_id are required"), nil
		}

		createReq := client.CreateTagRequest{
			ClipID:    clipID,
			SessionID: sessionID,
		}

		if playType, ok := req.Params.Arguments["play_type"].(string); ok {
			createReq.PlayType = &playType
		}
		if formation, ok := req.Params.Arguments["formation"].(string); ok {
			createReq.Formation = &formation
		}
		if result, ok := req.Params.Arguments["result"].(string); ok {
			createReq.Result = &result
		}
		if down, ok := req.Params.Arguments["down"].(float64); ok {
			d := int(down)
			createReq.Down = &d
		}
		if distance, ok := req.Params.Arguments["distance"].(float64); ok {
			d := int(distance)
			createReq.Distance = &d
		}
		if yardsGained, ok := req.Params.Arguments["yards_gained"].(float64); ok {
			y := int(yardsGained)
			createReq.YardsGained = &y
		}
		if notes, ok := req.Params.Arguments["notes"].(string); ok {
			createReq.Notes = &notes
		}

		tag, err := c.CreateTag(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create tag: %v", err)), nil
		}

		data, _ := json.MarshalIndent(tag, "", "  ")
		return mcp.NewToolResultText(fmt.Sprintf("Tag created:\n%s", string(data))), nil
	}
}
