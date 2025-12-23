package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	c := New("http://localhost:8080")
	if c == nil {
		t.Fatal("New() returned nil client")
	}
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("New() baseURL = %v, want %v", c.baseURL, "http://localhost:8080")
	}
	if c.httpClient == nil {
		t.Error("New() httpClient is nil")
	}
}

func TestClient_ListSessions(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/sessions" {
			t.Errorf("Expected path /api/v1/sessions, got %s", r.URL.Path)
		}

		resp := PaginatedResponse[Session]{
			Data: []Session{
				{ID: "session-1", Name: "Game 1", SessionType: "game", Status: "scheduled"},
				{ID: "session-2", Name: "Practice 1", SessionType: "practice", Status: "active"},
			},
			Total:  2,
			Limit:  20,
			Offset: 0,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	result, err := c.ListSessions(context.Background(), ListSessionsParams{})
	if err != nil {
		t.Fatalf("ListSessions() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("ListSessions() returned nil")
	}
	if len(result.Data) != 2 {
		t.Errorf("ListSessions() returned %d sessions, want 2", len(result.Data))
	}
	if result.Total != 2 {
		t.Errorf("ListSessions() total = %d, want 2", result.Total)
	}
}

func TestClient_ListSessions_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("Expected status=active, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("session_type") != "game" {
			t.Errorf("Expected session_type=game, got %s", r.URL.Query().Get("session_type"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		resp := PaginatedResponse[Session]{
			Data:   []Session{},
			Total:  0,
			Limit:  10,
			Offset: 0,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	_, err := c.ListSessions(context.Background(), ListSessionsParams{
		Status:      "active",
		SessionType: "game",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("ListSessions() unexpected error: %v", err)
	}
}

func TestClient_GetSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/sessions/session-123" {
			t.Errorf("Expected path /api/v1/sessions/session-123, got %s", r.URL.Path)
		}

		session := Session{
			ID:          "session-123",
			Name:        "Test Game",
			SessionType: "game",
			Status:      "active",
			ClipCount:   5,
			TagCount:    10,
		}
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := New(server.URL)
	session, err := c.GetSession(context.Background(), "session-123")
	if err != nil {
		t.Fatalf("GetSession() unexpected error: %v", err)
	}
	if session.ID != "session-123" {
		t.Errorf("GetSession() ID = %v, want session-123", session.ID)
	}
	if session.ClipCount != 5 {
		t.Errorf("GetSession() ClipCount = %v, want 5", session.ClipCount)
	}
}

func TestClient_CreateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "New Game" {
			t.Errorf("Expected name 'New Game', got %s", req.Name)
		}
		if req.SessionType != "game" {
			t.Errorf("Expected session_type 'game', got %s", req.SessionType)
		}

		session := Session{
			ID:          "new-session-id",
			Name:        req.Name,
			SessionType: req.SessionType,
			Status:      "scheduled",
		}
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := New(server.URL)
	session, err := c.CreateSession(context.Background(), CreateSessionRequest{
		Name:        "New Game",
		SessionType: "game",
	})
	if err != nil {
		t.Fatalf("CreateSession() unexpected error: %v", err)
	}
	if session.ID != "new-session-id" {
		t.Errorf("CreateSession() ID = %v, want new-session-id", session.ID)
	}
}

func TestClient_SessionStateTransitions(t *testing.T) {
	t.Run("StartSession", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/sessions/session-1/start" {
				t.Errorf("Expected path /api/v1/sessions/session-1/start, got %s", r.URL.Path)
			}
			session := Session{ID: "session-1", Status: "active"}
			json.NewEncoder(w).Encode(session)
		}))
		defer server.Close()

		c := New(server.URL)
		session, err := c.StartSession(context.Background(), "session-1")
		if err != nil {
			t.Fatalf("StartSession() unexpected error: %v", err)
		}
		if session.Status != "active" {
			t.Errorf("StartSession() status = %v, want active", session.Status)
		}
	})

	t.Run("PauseSession", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/sessions/session-1/pause" {
				t.Errorf("Expected path /api/v1/sessions/session-1/pause, got %s", r.URL.Path)
			}
			session := Session{ID: "session-1", Status: "paused"}
			json.NewEncoder(w).Encode(session)
		}))
		defer server.Close()

		c := New(server.URL)
		session, err := c.PauseSession(context.Background(), "session-1")
		if err != nil {
			t.Fatalf("PauseSession() unexpected error: %v", err)
		}
		if session.Status != "paused" {
			t.Errorf("PauseSession() status = %v, want paused", session.Status)
		}
	})

	t.Run("CompleteSession", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/sessions/session-1/complete" {
				t.Errorf("Expected path /api/v1/sessions/session-1/complete, got %s", r.URL.Path)
			}
			session := Session{ID: "session-1", Status: "completed"}
			json.NewEncoder(w).Encode(session)
		}))
		defer server.Close()

		c := New(server.URL)
		session, err := c.CompleteSession(context.Background(), "session-1")
		if err != nil {
			t.Fatalf("CompleteSession() unexpected error: %v", err)
		}
		if session.Status != "completed" {
			t.Errorf("CompleteSession() status = %v, want completed", session.Status)
		}
	})
}

func TestClient_ListClips(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PaginatedResponse[Clip]{
			Data: []Clip{
				{ID: "clip-1", SessionID: "session-1", Status: "ready"},
				{ID: "clip-2", SessionID: "session-1", Status: "ready"},
			},
			Total: 2,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	result, err := c.ListClips(context.Background(), ListClipsParams{
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("ListClips() unexpected error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("ListClips() returned %d clips, want 2", len(result.Data))
	}
}

func TestClient_FavoriteClip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/clips/clip-1/favorite" {
			t.Errorf("Expected path /api/v1/clips/clip-1/favorite, got %s", r.URL.Path)
		}
		clip := Clip{ID: "clip-1", IsFavorite: true}
		json.NewEncoder(w).Encode(clip)
	}))
	defer server.Close()

	c := New(server.URL)
	clip, err := c.FavoriteClip(context.Background(), "clip-1")
	if err != nil {
		t.Fatalf("FavoriteClip() unexpected error: %v", err)
	}
	if !clip.IsFavorite {
		t.Error("FavoriteClip() IsFavorite should be true")
	}
}

func TestClient_ListChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PaginatedResponse[Channel]{
			Data: []Channel{
				{ID: "camera-1", Name: "Main Camera", Status: "active"},
				{ID: "camera-2", Name: "Secondary Camera", Status: "inactive"},
			},
			Total: 2,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	result, err := c.ListChannels(context.Background())
	if err != nil {
		t.Fatalf("ListChannels() unexpected error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("ListChannels() returned %d channels, want 2", len(result.Data))
	}
}

func TestClient_ActivateDeactivateChannel(t *testing.T) {
	t.Run("ActivateChannel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/channels/camera-1/activate" {
				t.Errorf("Expected path /api/v1/channels/camera-1/activate, got %s", r.URL.Path)
			}
			channel := Channel{ID: "camera-1", Status: "active"}
			json.NewEncoder(w).Encode(channel)
		}))
		defer server.Close()

		c := New(server.URL)
		channel, err := c.ActivateChannel(context.Background(), "camera-1")
		if err != nil {
			t.Fatalf("ActivateChannel() unexpected error: %v", err)
		}
		if channel.Status != "active" {
			t.Errorf("ActivateChannel() status = %v, want active", channel.Status)
		}
	})

	t.Run("DeactivateChannel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/channels/camera-1/deactivate" {
				t.Errorf("Expected path /api/v1/channels/camera-1/deactivate, got %s", r.URL.Path)
			}
			channel := Channel{ID: "camera-1", Status: "inactive"}
			json.NewEncoder(w).Encode(channel)
		}))
		defer server.Close()

		c := New(server.URL)
		channel, err := c.DeactivateChannel(context.Background(), "camera-1")
		if err != nil {
			t.Fatalf("DeactivateChannel() unexpected error: %v", err)
		}
		if channel.Status != "inactive" {
			t.Errorf("DeactivateChannel() status = %v, want inactive", channel.Status)
		}
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		c := New(server.URL)
		_, err := c.GetSession(context.Background(), "non-existent")
		if err == nil {
			t.Error("GetSession() expected error for 404")
		}
	})

	t.Run("500 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		c := New(server.URL)
		_, err := c.ListSessions(context.Background(), ListSessionsParams{})
		if err == nil {
			t.Error("ListSessions() expected error for 500")
		}
	})
}

func TestClient_ListTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PaginatedResponse[Tag]{
			Data: []Tag{
				{ID: "tag-1", ClipID: "clip-1", SessionID: "session-1", IsImportant: true},
				{ID: "tag-2", ClipID: "clip-2", SessionID: "session-1", IsReviewed: true},
			},
			Total: 2,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := New(server.URL)
	result, err := c.ListTags(context.Background(), ListTagsParams{
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("ListTags() unexpected error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("ListTags() returned %d tags, want 2", len(result.Data))
	}
}

func TestClient_CreateTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var req CreateTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.ClipID != "clip-1" {
			t.Errorf("Expected clip_id 'clip-1', got %s", req.ClipID)
		}

		tag := Tag{
			ID:        "new-tag-id",
			ClipID:    req.ClipID,
			SessionID: req.SessionID,
		}
		json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	c := New(server.URL)
	playType := "run"
	tag, err := c.CreateTag(context.Background(), CreateTagRequest{
		ClipID:    "clip-1",
		SessionID: "session-1",
		PlayType:  &playType,
	})
	if err != nil {
		t.Fatalf("CreateTag() unexpected error: %v", err)
	}
	if tag.ID != "new-tag-id" {
		t.Errorf("CreateTag() ID = %v, want new-tag-id", tag.ID)
	}
}
