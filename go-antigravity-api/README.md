# Go Antigravity API

A standalone Go implementation of the Antigravity API, providing a Gemini-compatible interface to access Google's internal Antigravity service with support for the latest models including Gemini 3 Pro, Claude 4.5 Opus, and more.

## Features

- **🔐 OAuth2 Authentication** - Automatic Google OAuth2 flow with token management and refresh
- **🌐 Multi-Environment Support** - Automatic failover between Daily and Autopush environments
- **🔄 Intelligent Retry Logic** - Exponential backoff with automatic environment switching
- **📊 Model Management** - Dynamic model discovery and alias mapping
- **💬 Streaming Support** - Server-Sent Events (SSE) for real-time responses
- **📈 Quota Monitoring** - Real-time quota tracking for all models
- **🔧 Flexible Configuration** - JSON config file or environment variables
- **⚡ High Performance** - Built with Go for superior performance

## Supported Models

The service provides access to the following cutting-edge models:

- **Gemini 3 Pro** (`gemini-3-pro-preview`) - Latest Gemini architecture
- **Gemini 3 Flash** (`gemini-3-flash-preview`) - Fast and efficient
- **Gemini 2.5 Flash** (`gemini-2.5-flash`) - Balanced performance
- **Gemini 2.5 Computer Use** (`gemini-2.5-computer-use-preview-10-2025`) - Computer interaction
- **Claude Sonnet 4.5** (`gemini-claude-sonnet-4-5`) - Anthropic's Claude via Antigravity
- **Claude Opus 4.5 Thinking** (`gemini-claude-opus-4-5-thinking`) - Advanced reasoning

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Google account with Antigravity access
- OAuth2 credentials (automatically handled)

### Installation

```bash
# Clone the repository
cd go-antigravity-api

# Install dependencies
go mod download

# Build the server
go build -o antigravity-api cmd/server/main.go
```

### Running the Server

#### Using Configuration File

```bash
# Copy and edit the example config
cp example-config.json config.json
# Edit config.json with your settings

# Start the server
./antigravity-api -config config.json
```

#### Using Environment Variables

```bash
export SERVER_PORT=3000
export SERVER_API_KEY=your-secret-key
export ANTIGRAVITY_CREDS_FILE=/path/to/oauth_creds.json
export ANTIGRAVITY_PROJECT_ID=your-project-id

./antigravity-api
```

#### Using Command Line Arguments

```bash
./antigravity-api -config config.json -port 8080 -project-id my-project
```

### First-Time OAuth Setup

When you run the server for the first time, it will guide you through the OAuth flow:

1. The server will print an authorization URL
2. Visit the URL in your browser
3. Sign in with your Google account
4. Authorize the application
5. Copy the authorization code
6. Paste it into the terminal
7. Credentials will be saved to `~/.antigravity/oauth_creds.json`

Subsequent runs will use the saved credentials automatically.

## Configuration

### Configuration File Format

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 3000,
    "apiKey": "your-api-key-here"
  },
  "antigravity": {
    "oauthCredsFilePath": "/path/to/oauth_creds.json",
    "projectId": "",
    "baseUrlDaily": "https://daily-cloudcode-pa.sandbox.googleapis.com",
    "baseUrlAutopush": "https://autopush-cloudcode-pa.sandbox.googleapis.com",
    "userAgent": "antigravity/1.11.5 windows/amd64",
    "requestMaxRetries": 3,
    "requestBaseDelay": 1000,
    "cronNearMinutes": 50
  }
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `SERVER_PORT` | Server port | `3000` |
| `SERVER_API_KEY` | API key for authentication | (none) |
| `ANTIGRAVITY_CREDS_FILE` | OAuth credentials file path | `~/.antigravity/oauth_creds.json` |
| `ANTIGRAVITY_PROJECT_ID` | Google Cloud Project ID | (auto-discovered) |
| `ANTIGRAVITY_BASE_URL_DAILY` | Daily environment URL | (default) |
| `ANTIGRAVITY_BASE_URL_AUTOPUSH` | Autopush environment URL | (default) |
| `ANTIGRAVITY_USER_AGENT` | User agent string | `antigravity/1.11.5 windows/amd64` |
| `ANTIGRAVITY_MAX_RETRIES` | Maximum retry attempts | `3` |
| `ANTIGRAVITY_BASE_DELAY` | Base delay for retries (ms) | `1000` |

## API Usage

The service provides a Gemini-compatible API interface:

### List Models

```bash
curl http://localhost:3000/v1beta/models \
  -H "Authorization: Bearer your-api-key"
```

### Generate Content (Non-Streaming)

```bash
curl http://localhost:3000/v1beta/models/gemini-3-pro-preview:generateContent \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Explain quantum computing in simple terms"}]
      }
    ]
  }'
```

### Stream Generate Content (Streaming)

```bash
curl http://localhost:3000/v1beta/models/gemini-3-pro-preview:streamGenerateContent \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Write a short story about AI"}]
      }
    ]
  }'
```

### Get Usage Limits

```bash
curl http://localhost:3000/usage \
  -H "Authorization: Bearer your-api-key"
```

### Health Check

```bash
curl http://localhost:3000/health
```

## Architecture

### Project Structure

```
go-antigravity-api/
├── cmd/
│   └── server/
│       └── main.go              # Main entry point
├── internal/
│   ├── antigravity/
│   │   ├── auth.go              # OAuth2 authentication
│   │   ├── client.go            # API client with retry logic
│   │   ├── constants.go         # Constants and model mappings
│   │   ├── converter.go         # Format conversion
│   │   ├── discovery.go         # Project discovery
│   │   ├── models.go            # Model management
│   │   ├── quota.go             # Quota tracking
│   │   ├── service.go           # Main service orchestration
│   │   ├── stream.go            # SSE streaming
│   │   └── utils.go             # Utility functions
│   ├── api/
│   │   └── handlers.go          # HTTP handlers
│   └── config/
│       └── config.go            # Configuration management
├── pkg/
│   └── models/
│       └── types.go             # Shared type definitions
├── go.mod                       # Go module definition
├── example-config.json          # Example configuration
├── .gitignore                   # Git ignore rules
└── README.md                    # This file
```

### Key Components

1. **Authentication Manager** - Handles OAuth2 flow, token refresh, and credential storage
2. **API Client** - Makes HTTP requests with intelligent retry and failover
3. **Discovery Manager** - Discovers or creates Google Cloud Project ID
4. **Models Manager** - Manages available models and their capabilities
5. **Quota Manager** - Tracks and reports usage quotas
6. **Converter** - Transforms between Gemini and Antigravity formats
7. **Stream Handler** - Parses SSE streams for real-time responses

### Multi-Environment Failover

The service automatically tries multiple environments in order:

1. **Daily Environment** - Primary endpoint
2. **Autopush Environment** - Fallback endpoint

Switches occur on:
- 429 (Rate Limit) errors
- Network failures
- 5xx server errors

### Error Handling

- **400/401** - Refresh OAuth token and retry
- **429** - Switch environment or exponential backoff
- **500-599** - Exponential backoff with retry
- **Network errors** - Immediate environment switch

## Advanced Features

### Token Management

- Automatic token refresh 50 minutes before expiry
- Persistent credential storage
- Graceful handling of expired tokens

### Project ID Management

The service handles Project ID in this priority order:

1. Configuration file
2. Command line argument
3. Auto-discovery via `loadCodeAssist` API
4. Auto-creation via `onboardUser` API
5. Random generation as fallback

### Model Aliasing

Friendly model names are automatically mapped to internal names:

- `gemini-3-pro-preview` → `gemini-3-pro-high`
- `gemini-claude-sonnet-4-5` → `claude-sonnet-4-5`

## Comparison with Node.js Version

| Feature | Go Version | Node.js Version |
|---------|-----------|-----------------|
| Performance | ⚡ Faster | Standard |
| Memory Usage | 💾 Lower | Higher |
| Concurrency | 🔀 Native goroutines | Event loop |
| Compilation | ✅ Single binary | Runtime required |
| Dependencies | Minimal | Multiple packages |
| Type Safety | ✅ Strong typing | Dynamic typing |

## Troubleshooting

### OAuth Issues

**Problem**: "Failed to get access token"
- **Solution**: Delete `~/.antigravity/oauth_creds.json` and re-authenticate

**Problem**: "Token refresh failed"
- **Solution**: Ensure your Google account has Antigravity access

### API Errors

**Problem**: "All Antigravity base URLs failed"
- **Solution**: Check network connectivity and firewall settings

**Problem**: "Rate limit exceeded"
- **Solution**: The service will automatically retry with backoff

### Project ID Issues

**Problem**: "Failed to discover Project ID"
- **Solution**: Specify `projectId` in config or use `-project-id` flag

## Development

### Building

```bash
# Build for current platform
go build -o antigravity-api cmd/server/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o antigravity-api-linux cmd/server/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o antigravity-api.exe cmd/server/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o antigravity-api-macos cmd/server/main.go
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...
```

## Security Considerations

- Store OAuth credentials securely
- Use strong API keys
- Run behind HTTPS in production
- Restrict network access appropriately
- Regularly rotate credentials

## License

This project follows the same license as the parent AIClient-2-API project.

## Contributing

Contributions are welcome! Please ensure:

- Code follows Go best practices
- Tests are included for new features
- Documentation is updated
- Commits are well-described

## Support

For issues and questions:
- Open an issue in the GitHub repository
- Check existing issues for solutions
- Review the troubleshooting section

## Acknowledgments

This implementation is based on the Node.js Antigravity service in AIClient-2-API, with enhancements for Go's performance and concurrency models.
