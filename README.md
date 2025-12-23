# Video Platform MCP Server

MCP (Model Context Protocol) server that exposes the video streaming platform API to AI assistants like Claude Desktop.

## Features

### Tools
- **list_sessions** - List all recording sessions with optional filters
- **create_session** - Create a new recording session
- **start_session** - Start a scheduled session
- **pause_session** - Pause an active session
- **complete_session** - Complete/end a session
- **list_clips** - List video clips with filters (session, favorites, etc.)
- **favorite_clip** - Toggle favorite status on a clip
- **list_channels** - List all video input channels
- **activate_channel** - Activate a channel for recording
- **deactivate_channel** - Deactivate a channel
- **list_tags** - List clip annotations/tags
- **create_tag** - Create a new tag annotation

### Resources
- `video://sessions` - List of all recording sessions
- `video://clips` - List of all video clips
- `video://channels` - Channel status information
- `video://tags` - List of all tags

### Prompts
- **analyze_session** - Analyze a game/practice session for patterns and insights
- **review_clips** - Review and provide feedback on clips from a session
- **game_report** - Generate a comprehensive game report
- **system_status** - Check system health and active channels

## Installation

```bash
# Build
go build -o video-mcp ./cmd/server

# Install globally
sudo cp video-mcp /usr/local/bin/
```

## Usage

### Command Line

```bash
# Default (connects to localhost:8080)
./video-mcp

# Custom API URL
./video-mcp -api-url http://192.168.1.100:8080

# Using environment variable
VIDEO_PLATFORM_URL=http://myserver:8080 ./video-mcp
```

### Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "video-platform": {
      "command": "/usr/local/bin/video-mcp",
      "args": ["-api-url", "http://localhost:8080"]
    }
  }
}
```

Or with environment variable:

```json
{
  "mcpServers": {
    "video-platform": {
      "command": "/usr/local/bin/video-mcp",
      "env": {
        "VIDEO_PLATFORM_URL": "http://localhost:8080"
      }
    }
  }
}
```

## Development

```bash
# Run locally
go run ./cmd/server -api-url http://localhost:8080

# Test with MCP inspector
npx @modelcontextprotocol/inspector go run ./cmd/server
```

## Requirements

- Go 1.23+
- Running video-platform API server
