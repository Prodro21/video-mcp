package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Prodro21/video-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterResources adds all resource handlers to the server
func RegisterResources(s *server.MCPServer, c *client.Client) {
	// Sessions list
	s.AddResource(mcp.Resource{
		URI:         "video://sessions",
		Name:        "All Sessions",
		Description: "List of all recording sessions",
		MIMEType:    "application/json",
	}, makeSessionsResource(c))

	// Clips list
	s.AddResource(mcp.Resource{
		URI:         "video://clips",
		Name:        "All Clips",
		Description: "List of all video clips",
		MIMEType:    "application/json",
	}, makeClipsResource(c))

	// Channels list
	s.AddResource(mcp.Resource{
		URI:         "video://channels",
		Name:        "All Channels",
		Description: "List of all video input channels and their status",
		MIMEType:    "application/json",
	}, makeChannelsResource(c))

	// Tags list
	s.AddResource(mcp.Resource{
		URI:         "video://tags",
		Name:        "All Tags",
		Description: "List of all clip annotations/tags",
		MIMEType:    "application/json",
	}, makeTagsResource(c))
}

func makeSessionsResource(c *client.Client) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		resp, err := c.ListSessions(ctx, client.ListSessionsParams{Limit: 100})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sessions: %w", err)
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
				},
				Text: string(data),
			},
		}, nil
	}
}


func makeClipsResource(c *client.Client) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		resp, err := c.ListClips(ctx, client.ListClipsParams{Limit: 100})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch clips: %w", err)
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
				},
				Text: string(data),
			},
		}, nil
	}
}


func makeChannelsResource(c *client.Client) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		resp, err := c.ListChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channels: %w", err)
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
				},
				Text: string(data),
			},
		}, nil
	}
}

func makeTagsResource(c *client.Client) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]interface{}, error) {
		resp, err := c.ListTags(ctx, client.ListTagsParams{Limit: 100})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return []interface{}{
			mcp.TextResourceContents{
				ResourceContents: mcp.ResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
				},
				Text: string(data),
			},
		}, nil
	}
}

