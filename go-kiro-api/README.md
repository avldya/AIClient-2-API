# Go Kiro API

A high-performance Go implementation of the Kiro API client, providing OpenAI and Claude compatible interfaces for accessing CodeWhisperer/Amazon Q models.

## Features

- ✅ **Multiple Authentication Methods**: Support for Base64 credentials, file paths, and directory scanning
- ✅ **Automatic Token Refresh**: OAuth token management with automatic refresh before expiry
- ✅ **OpenAI Compatible**: Full support for `/v1/chat/completions` endpoint
- ✅ **Claude Compatible**: Full support for `/v1/messages` endpoint
- ✅ **Streaming Support**: Real-time streaming responses for both OpenAI and Claude formats
- ✅ **Tool Calling**: Complete support for function/tool calling
- ✅ **Retry Logic**: Automatic retry with exponential backoff for failed requests
- ✅ **Error Handling**: Comprehensive error handling with automatic token refresh on 403 errors
- ✅ **Usage Tracking**: Query usage limits and remaining quota

## Supported Models

- `claude-opus-4-5`
- `claude-opus-4-5-20251101`
- `claude-haiku-4-5`
- `claude-sonnet-4-5`
- `claude-sonnet-4-5-20250929`
- `claude-sonnet-4-20250514`
- `claude-3-7-sonnet-20250219`

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/avldya/AIClient-2-API.git
cd AIClient-2-API/go-kiro-api

# Build the binary
go build -o kiro-api cmd/server/main.go

# Run the server
./kiro-api -config config.json
```

### Configuration

Create a `config.json` file:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 3000,
    "apiKey": "your-api-key"
  },
  "kiro": {
    "credPath": "/path/to/.aws/sso/cache",
    "credsBase64": "",
    "credsFilePath": "",
    "region": "us-east-1",
    "requestMaxRetries": 3,
    "requestBaseDelay": 1000,
    "cronNearMinutes": 10
  }
}
```

### Environment Variables

You can also configure using environment variables:

```bash
export SERVER_HOST=0.0.0.0
export SERVER_PORT=3000
export SERVER_API_KEY=your-api-key
export KIRO_CREDS_BASE64=your-base64-encoded-credentials
export KIRO_REGION=us-east-1
```

### Running the Server

```bash
# Using config file
./kiro-api -config config.json

# Using command line arguments
./kiro-api -host 0.0.0.0 -port 3000 -api-key your-api-key

# Using environment variables
export KIRO_CREDS_BASE64="..."
export SERVER_PORT=3000
./kiro-api
```

## API Usage

### OpenAI Format (Non-Streaming)

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

### OpenAI Format (Streaming)

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "Tell me a story"}
    ],
    "stream": true
  }'
```

### Claude Format (Non-Streaming)

```bash
curl http://localhost:3000/v1/messages \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 1024
  }'
```

### Claude Format (Streaming)

```bash
curl http://localhost:3000/v1/messages \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "Write a poem"}
    ],
    "stream": true
  }'
```

### List Models

```bash
curl http://localhost:3000/v1/models \
  -H "Authorization: Bearer your-api-key"
```

### Count Tokens

```bash
curl http://localhost:3000/v1/messages/count_tokens \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ]
  }'
```

### Get Usage Limits

```bash
curl http://localhost:3000/getUsageLimits \
  -H "Authorization: Bearer your-api-key"
```

## Configuration Options

### Server Configuration

- `server.host`: Server listening address (default: `0.0.0.0`)
- `server.port`: Server listening port (default: `3000`)
- `server.apiKey`: API key for authentication (optional)

### Kiro Configuration

- `kiro.credPath`: Path to credentials directory (default: `~/.aws/sso/cache`)
- `kiro.credsBase64`: Base64-encoded credentials (highest priority)
- `kiro.credsFilePath`: Path to specific credentials file
- `kiro.region`: AWS region (default: `us-east-1`)
- `kiro.proxy`: HTTP proxy URL (optional)
- `kiro.requestMaxRetries`: Maximum number of retries (default: `3`)
- `kiro.requestBaseDelay`: Base delay for retry in milliseconds (default: `1000`)
- `kiro.cronNearMinutes`: Minutes before token expiry to trigger refresh (default: `10`)

## Authentication Methods

The client supports three methods for loading credentials (in priority order):

1. **Base64 Credentials**: Provide credentials as a Base64-encoded JSON string via `credsBase64`
2. **File Path**: Specify a direct path to a credentials file via `credsFilePath`
3. **Directory Scanning**: Automatically scan a directory (like AWS CLI's `.aws/sso/cache`) via `credPath`

Credentials are merged from all available sources, with earlier sources taking precedence.

## Tool Calling Support

The API supports function/tool calling in both OpenAI and Claude formats:

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "What is the weather in San Francisco?"}
    ],
    "tools": [
      {
        "name": "get_weather",
        "description": "Get current weather for a location",
        "input_schema": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "City name"
            }
          },
          "required": ["location"]
        }
      }
    ]
  }'
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o kiro-api cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kiro-api .
EXPOSE 3000
CMD ["./kiro-api"]
```

Build and run:

```bash
docker build -t kiro-api .
docker run -p 3000:3000 \
  -e KIRO_CREDS_BASE64=your-base64-creds \
  kiro-api
```

### Binary Deployment

Build for different platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o kiro-api-linux cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o kiro-api-windows.exe cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o kiro-api-macos cmd/server/main.go
```

## Performance

The Go implementation offers significant performance improvements over the Node.js version:

- **Lower Memory Usage**: ~10-20MB vs 50-100MB for Node.js
- **Faster Startup**: ~100ms vs 500ms for Node.js
- **Better Concurrency**: Native goroutines vs event loop
- **Smaller Binary**: ~10MB vs 50MB+ with node_modules

## Error Handling

The client implements comprehensive error handling:

- **403 Errors**: Automatically refresh token and retry
- **429 Errors**: Exponential backoff retry
- **5xx Errors**: Exponential backoff retry
- **Token Expiry**: Proactive refresh 10 minutes before expiry

## Architecture

```
go-kiro-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── kiro/
│   │   ├── client.go           # HTTP client with retry logic
│   │   ├── auth.go             # Authentication and token management
│   │   ├── converter.go        # Format conversion logic
│   │   ├── stream.go           # Stream response parsing
│   │   ├── constants.go        # Constants and model mappings
│   │   └── utils.go            # Utility functions
│   ├── api/
│   │   ├── handlers.go         # Common HTTP handlers
│   │   ├── openai.go           # OpenAI compatible endpoints
│   │   └── claude.go           # Claude compatible endpoints
│   └── config/
│       └── config.go            # Configuration management
└── pkg/
    └── models/
        └── types.go             # Type definitions
```

## Comparison with Node.js Version

| Feature | Go Version | Node.js Version |
|---------|-----------|----------------|
| Memory Usage | 10-20MB | 50-100MB |
| Startup Time | ~100ms | ~500ms |
| Binary Size | ~10MB | 50MB+ (with node_modules) |
| Concurrency | Native goroutines | Event loop |
| Type Safety | Strong static typing | Dynamic typing |
| Dependencies | Minimal (uuid only) | Multiple npm packages |

## License

This project is licensed under the GPL v3 License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.
