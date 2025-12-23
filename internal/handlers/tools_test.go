package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Prodro21/video-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Helper to create a mock server with custom response
func mockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestListSessions(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			resp := client.PaginatedResponse[client.Session]{
				Data: []client.Session{
					{ID: "session-1", Name: "Game 1", SessionType: "game", Status: "active"},
					{ID: "session-2", Name: "Practice 1", SessionType: "practice", Status: "scheduled"},
				},
				Total:  2,
				Limit:  20,
				Offset: 0,
			}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListSessions(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success, got error")
		}

		// Check that response contains session data
		if len(result.Content) == 0 {
			t.Error("Expected content in result")
		}
	})

	t.Run("with filters", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			// Verify filters are passed
			if r.URL.Query().Get("status") != "active" {
				t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
			}
			if r.URL.Query().Get("session_type") != "game" {
				t.Errorf("Expected session_type=game, got %s", r.URL.Query().Get("session_type"))
			}

			resp := client.PaginatedResponse[client.Session]{Data: []client.Session{}}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListSessions(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"status":       "active",
			"session_type": "game",
			"limit":        float64(10),
		}

		_, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("API error", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListSessions(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result")
		}
	})
}

func TestCreateSession(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			var req client.CreateSessionRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.Name != "New Game" {
				t.Errorf("Expected name 'New Game', got %s", req.Name)
			}

			session := client.Session{
				ID:          "new-session-id",
				Name:        req.Name,
				SessionType: req.SessionType,
				Status:      "scheduled",
			}
			json.NewEncoder(w).Encode(session)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeCreateSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"name":         "New Game",
			"session_type": "game",
			"opponent":     "Team B",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success, got error")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		c := client.New("http://localhost:8080")
		handler := makeCreateSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"name": "Only Name",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for missing session_type")
		}
	})
}

func TestStartSession(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/start") {
				t.Errorf("Expected start endpoint, got %s", r.URL.Path)
			}

			session := client.Session{
				ID:     "session-123",
				Name:   "Test Session",
				Status: "active",
			}
			json.NewEncoder(w).Encode(session)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeStartSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id": "session-123",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}

		// Check result contains success message
		if len(result.Content) > 0 {
			content := result.Content[0].(mcp.TextContent)
			if !strings.Contains(content.Text, "started successfully") {
				t.Errorf("Expected success message, got: %s", content.Text)
			}
		}
	})

	t.Run("missing session_id", func(t *testing.T) {
		c := client.New("http://localhost:8080")
		handler := makeStartSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for missing session_id")
		}
	})
}

func TestPauseSession(t *testing.T) {
	t.Run("successful pause", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			session := client.Session{
				ID:     "session-123",
				Name:   "Test Session",
				Status: "paused",
			}
			json.NewEncoder(w).Encode(session)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makePauseSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id": "session-123",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})
}

func TestCompleteSession(t *testing.T) {
	t.Run("successful complete", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			session := client.Session{
				ID:     "session-123",
				Name:   "Test Session",
				Status: "completed",
			}
			json.NewEncoder(w).Encode(session)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeCompleteSession(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id": "session-123",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})
}

func TestListClips(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			resp := client.PaginatedResponse[client.Clip]{
				Data: []client.Clip{
					{ID: "clip-1", SessionID: "session-1", Status: "ready"},
					{ID: "clip-2", SessionID: "session-1", Status: "ready"},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListClips(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id": "session-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})
}

func TestFavoriteClip(t *testing.T) {
	t.Run("add to favorites", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			clip := client.Clip{
				ID:         "clip-1",
				IsFavorite: true,
			}
			json.NewEncoder(w).Encode(clip)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeFavoriteClip(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"clip_id": "clip-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}

		// Check message
		if len(result.Content) > 0 {
			content := result.Content[0].(mcp.TextContent)
			if !strings.Contains(content.Text, "added to") {
				t.Errorf("Expected 'added to' message, got: %s", content.Text)
			}
		}
	})

	t.Run("remove from favorites", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			clip := client.Clip{
				ID:         "clip-1",
				IsFavorite: false,
			}
			json.NewEncoder(w).Encode(clip)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeFavoriteClip(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"clip_id": "clip-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check message
		if len(result.Content) > 0 {
			content := result.Content[0].(mcp.TextContent)
			if !strings.Contains(content.Text, "removed from") {
				t.Errorf("Expected 'removed from' message, got: %s", content.Text)
			}
		}
	})

	t.Run("missing clip_id", func(t *testing.T) {
		c := client.New("http://localhost:8080")
		handler := makeFavoriteClip(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for missing clip_id")
		}
	})
}

func TestListChannels(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			resp := client.PaginatedResponse[client.Channel]{
				Data: []client.Channel{
					{ID: "camera-1", Name: "Main Camera", Status: "active"},
					{ID: "camera-2", Name: "Secondary", Status: "inactive"},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListChannels(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})
}

func TestActivateChannel(t *testing.T) {
	t.Run("successful activation", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			channel := client.Channel{
				ID:     "camera-1",
				Name:   "Main Camera",
				Status: "active",
			}
			json.NewEncoder(w).Encode(channel)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeActivateChannel(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"channel_id": "camera-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}

		// Check message
		if len(result.Content) > 0 {
			content := result.Content[0].(mcp.TextContent)
			if !strings.Contains(content.Text, "activated") {
				t.Errorf("Expected 'activated' message, got: %s", content.Text)
			}
		}
	})

	t.Run("missing channel_id", func(t *testing.T) {
		c := client.New("http://localhost:8080")
		handler := makeActivateChannel(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for missing channel_id")
		}
	})
}

func TestDeactivateChannel(t *testing.T) {
	t.Run("successful deactivation", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			channel := client.Channel{
				ID:     "camera-1",
				Name:   "Main Camera",
				Status: "inactive",
			}
			json.NewEncoder(w).Encode(channel)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeDeactivateChannel(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"channel_id": "camera-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}

		// Check message
		if len(result.Content) > 0 {
			content := result.Content[0].(mcp.TextContent)
			if !strings.Contains(content.Text, "deactivated") {
				t.Errorf("Expected 'deactivated' message, got: %s", content.Text)
			}
		}
	})
}

func TestListTags(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			resp := client.PaginatedResponse[client.Tag]{
				Data: []client.Tag{
					{ID: "tag-1", ClipID: "clip-1", SessionID: "session-1"},
					{ID: "tag-2", ClipID: "clip-2", SessionID: "session-1"},
				},
				Total: 2,
			}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListTags(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id": "session-1",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})

	t.Run("with all filters", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			// Verify filters are passed
			if r.URL.Query().Get("session_id") != "session-1" {
				t.Errorf("Expected session_id=session-1, got %s", r.URL.Query().Get("session_id"))
			}
			if r.URL.Query().Get("play_type") != "run" {
				t.Errorf("Expected play_type=run, got %s", r.URL.Query().Get("play_type"))
			}

			resp := client.PaginatedResponse[client.Tag]{Data: []client.Tag{}}
			json.NewEncoder(w).Encode(resp)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeListTags(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"session_id":   "session-1",
			"play_type":    "run",
			"is_important": true,
			"is_reviewed":  false,
			"limit":        float64(25),
		}

		_, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

func TestCreateTag(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			var req client.CreateTagRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.ClipID != "clip-1" {
				t.Errorf("Expected clip_id 'clip-1', got %s", req.ClipID)
			}

			tag := client.Tag{
				ID:        "new-tag-id",
				ClipID:    req.ClipID,
				SessionID: req.SessionID,
			}
			json.NewEncoder(w).Encode(tag)
		})
		defer server.Close()

		c := client.New(server.URL)
		handler := makeCreateTag(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"clip_id":      "clip-1",
			"session_id":   "session-1",
			"play_type":    "run",
			"formation":    "I-Formation",
			"result":       "first_down",
			"down":         float64(1),
			"distance":     float64(10),
			"yards_gained": float64(5),
			"notes":        "Good blocking",
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success")
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		c := client.New("http://localhost:8080")
		handler := makeCreateTag(c)

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"clip_id": "clip-1",
			// missing session_id
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error for missing session_id")
		}
	})
}

// Verify error result is properly set
func verifyError(t *testing.T, result *mcp.CallToolResult, expectedMsg string) {
	t.Helper()
	if !result.IsError {
		t.Error("Expected IsError to be true")
	}
	if len(result.Content) > 0 {
		content := result.Content[0].(mcp.TextContent)
		if !strings.Contains(content.Text, expectedMsg) {
			t.Errorf("Expected error message containing '%s', got: %s", expectedMsg, content.Text)
		}
	}
}

// Suppress unused import error
var _ = errors.New
