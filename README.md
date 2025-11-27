# AI Service - Go

A Golang-based AI service that provides vector search and conversational AI capabilities using OpenAI and Qdrant. This is a Go port of the original Node.js AI service.

## Features

- **Vector Search**: Semantic search using OpenAI embeddings and Qdrant vector database
- **Conversational AI**: Context-aware responses using OpenAI GPT models
- **Layer Updates**: Dynamic information about data layer updates
- **Security**: Rate limiting, CORS, compression, and security headers
- **Logging**: Structured logging with logrus
- **Docker Support**: Containerized deployment with Docker Compose

## Project Structure

```
ai-service-go/
├── app/                    # Application setup
│   └── app.go             # Router and middleware configuration
├── controllers/           # Business logic
│   └── vector_store.go   # Vector store operations
├── data/                  # Data structures and prompts
│   ├── prompts.go        # AI prompts and instructions
│   ├── updates.go        # Layer update information
│   └── layer_matcher.go  # Layer name matching logic
├── middleware/            # HTTP middleware
│   ├── error_handler.go  # Error handling
│   └── rate_limiter.go   # Rate limiting
├── routers/              # HTTP routes
│   └── router.go         # API endpoints
├── utils/                # Utilities
│   ├── logger.go         # Logging configuration
│   ├── app_error.go      # Custom error types
│   ├── split_text.go     # Text chunking
│   └── pricing.go        # Token pricing calculation
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── Dockerfile            # Docker configuration
└── compose.yml           # Docker Compose configuration
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for containerized deployment)
- OpenAI API key
- Qdrant instance (included in docker-compose)

## Environment Variables

Create a `.env` file in the project root:

```env
PORT=8080
QDRANT_HOST=http://localhost:6333
OPENAI_API_KEY=your_openai_api_key_here
SECRET=your_secret_key_here
EMBEDMODEL=text-embedding-3-small
GENMODEL=gpt-4o-mini
N_CHUNKS=5
ENV=development
```

## Installation

### Local Development

1. **Clone and navigate to the Go service directory:**
   ```bash
   cd ai-service-go
   ```

2. **Copy the example environment file:**
   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

3. **Copy data.txt from parent directory:**
   ```bash
   cp ../data.txt .
   ```

4. **Initialize Go modules:**
   ```bash
   go mod download
   ```

5. **Run the application:**
   ```bash
   go run main.go
   ```

### Docker Deployment

1. **Make sure you have a `.env` file with required variables**

2. **Build and start services:**
   ```bash
   docker-compose up -d
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f ai-service-go
   ```

4. **Stop services:**
   ```bash
   docker-compose down
   ```

## API Endpoints

### GET `/status`
Check the status of the vector store.

**Response:**
```json
{
  "status": "success",
  "data": {
    "initialized": true,
    "collection": {
      "name": "documents"
    }
  }
}
```

### GET `/load-embeddings?secret=YOUR_SECRET`
Load embeddings from `data.txt` into the vector store.

**Query Parameters:**
- `secret` (required): Authentication secret

**Response:**
```json
{
  "status": "success",
  "message": "Embeddings loaded successfully"
}
```

### POST `/handle-query?platform=standard`
Process a user query and generate a response.

**Query Parameters:**
- `platform` (optional): `standard` or `trial` (default: `standard`)

**Request Body:**
```json
{
  "query": "What is EEHORIZON?",
  "previousConversation": "[{\"role\":\"user\",\"content\":\"Hi\"}]"
}
```

**Response:**
```json
{
  "status": "success",
  "content": {
    "result": "EEHORIZON is...",
    "followUps": ["What features are available?"],
    "updateCycleQueryLayers": []
  }
}
```

## Key Differences from Node.js Version

1. **Performance**: Go offers better performance and lower memory footprint
2. **Concurrency**: Native goroutines for concurrent operations
3. **Type Safety**: Strong static typing catches errors at compile time
4. **Deployment**: Single binary deployment (no runtime needed)
5. **Dependencies**: Managed via `go.mod` instead of `package.json`

## Dependencies

- **gin-gonic/gin**: Web framework
- **qdrant/go-client**: Qdrant vector database client
- **sashabaranov/go-openai**: OpenAI API client
- **sirupsen/logrus**: Structured logging
- **gin-contrib/cors**: CORS middleware
- **gin-contrib/gzip**: Compression middleware
- **golang.org/x/time/rate**: Rate limiting

## Development

### Run tests:
```bash
go test ./...
```

### Build binary:
```bash
go build -o ai-service-go
```

### Format code:
```bash
go fmt ./...
```

### Run linter:
```bash
go vet ./...
```

## Production Considerations

1. **Logging**: Logs are written to `/logs/ai-service/app.log`
2. **Rate Limiting**: 100 requests per 15 minutes per IP
3. **Body Size**: Limited to 10KB
4. **CORS**: Configured for all origins (adjust for production)
5. **Graceful Shutdown**: Handles SIGTERM and SIGINT signals

## Troubleshooting

### Port already in use
```bash
# Change PORT in .env file or kill the process using the port
lsof -ti:8080 | xargs kill -9
```

### Qdrant connection issues
```bash
# Make sure Qdrant is running
docker-compose ps
# Check Qdrant logs
docker-compose logs qdrant
```

### Module import errors
```bash
# Clean and re-download modules
go clean -modcache
go mod download
```

## License

Same as the original Node.js service

## Support

For issues or questions, refer to the main project documentation or create an issue in the repository.
