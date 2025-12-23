package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterPrompts adds all prompt templates to the server
func RegisterPrompts(s *server.MCPServer) {
	s.AddPrompt(mcp.Prompt{
		Name:        "analyze_session",
		Description: "Analyze a game or practice session to identify patterns, key plays, and coaching insights",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "session_id",
				Description: "ID of the session to analyze",
				Required:    true,
			},
			{
				Name:        "focus",
				Description: "Analysis focus: offense, defense, special_teams, or all",
				Required:    false,
			},
		},
	}, handleAnalyzeSession)

	s.AddPrompt(mcp.Prompt{
		Name:        "review_clips",
		Description: "Review and provide feedback on clips from a session",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "session_id",
				Description: "ID of the session to review clips from",
				Required:    true,
			},
			{
				Name:        "play_type",
				Description: "Filter by play type (Run, Pass, etc.)",
				Required:    false,
			},
		},
	}, handleReviewClips)

	s.AddPrompt(mcp.Prompt{
		Name:        "game_report",
		Description: "Generate a comprehensive game report with statistics and analysis",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "session_id",
				Description: "ID of the game session",
				Required:    true,
			},
		},
	}, handleGameReport)

	s.AddPrompt(mcp.Prompt{
		Name:        "system_status",
		Description: "Check the status of the video platform including channels and active sessions",
		Arguments:   []mcp.PromptArgument{},
	}, handleSystemStatus)
}

func handleAnalyzeSession(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	sessionID := req.Params.Arguments["session_id"]
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	focus := "all"
	if f := req.Params.Arguments["focus"]; f != "" {
		focus = f
	}

	prompt := fmt.Sprintf(`Analyze the video session with ID: %s

Focus area: %s

Please follow these steps:

1. First, read the session summary resource:
   video://sessions/%s/summary

2. Review the session metadata including:
   - Session type (game/practice)
   - Duration and clip count
   - Opponent and location (if applicable)

3. Analyze the play distribution:
   - Break down plays by type (Run/Pass/etc.)
   - Identify successful vs unsuccessful plays
   - Note any patterns in formations or results

4. Highlight key moments:
   - Look for important tagged plays
   - Identify any scoring plays or turnovers
   - Note teaching moments

5. Provide coaching recommendations:
   - What worked well?
   - What needs improvement?
   - Specific drills or focus areas for practice

Format your analysis with clear sections and actionable insights.`, sessionID, focus, sessionID)

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

func handleReviewClips(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	sessionID := req.Params.Arguments["session_id"]
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	playType := req.Params.Arguments["play_type"]

	var playTypeFilter string
	if playType != "" {
		playTypeFilter = fmt.Sprintf("Filtered by play type: %s", playType)
	}

	prompt := fmt.Sprintf(`Review clips from session: %s
%s

Please help me review and analyze the clips:

1. Use the list_clips tool to get clips for this session:
   - Filter by session_id: %s
   %s

2. For each clip, examine:
   - Duration and status
   - Whether it's marked as favorite
   - Associated tags and annotations

3. Group clips by:
   - Play type (if tagged)
   - Formation used
   - Result (success/failure)

4. Identify clips that:
   - Should be marked as favorites (if not already)
   - Need additional tagging
   - Could be used for teaching

5. Provide a summary of:
   - Total clips reviewed
   - Highlights worth keeping
   - Clips that may need re-recording

Let me know if you need me to create tags for any clips or mark specific ones as favorites.`, sessionID, playTypeFilter, sessionID, playTypeFilter)

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

func handleGameReport(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	sessionID := req.Params.Arguments["session_id"]
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	prompt := fmt.Sprintf(`Generate a comprehensive game report for session: %s

1. First, fetch the complete session summary:
   video://sessions/%s/summary

2. Create a structured game report with these sections:

## Game Overview
- Date and opponent
- Final result (if available)
- Total plays and duration

## Offensive Summary
- Run/Pass ratio
- Successful plays vs unsuccessful
- Key formations used
- Yards gained breakdown

## Defensive Summary
- Stops and tackles
- Turnovers forced
- Coverage breakdowns

## Special Teams
- Punts, kickoffs, field goals
- Return yards

## Key Plays
- Touchdowns
- Turnovers
- Big plays (15+ yards)
- Critical third/fourth down conversions

## Areas for Improvement
- Identify weaknesses
- Suggest practice focus areas

## Player Highlights
- Notable performances
- Players who need additional coaching

Format the report professionally with clear headers and bullet points.`, sessionID, sessionID)

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

func handleSystemStatus(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	prompt := `Check the current status of the video platform system:

1. Use the list_channels tool to check all video input channels:
   - Which channels are active?
   - Are there any channels in error state?
   - When were channels last seen?

2. Use the list_sessions tool to check active sessions:
   - Are there any sessions currently recording (status: active)?
   - Are there scheduled sessions that should have started?
   - Recent completed sessions

3. Provide a status summary:
   - Overall system health
   - Active channel count
   - Current recording status
   - Any issues that need attention

If there are any problems, suggest corrective actions like:
- Activating inactive channels
- Starting scheduled sessions
- Investigating error states`

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}
