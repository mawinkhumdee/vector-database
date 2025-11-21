# Vector Database Demo

End-to-end Go service demonstrating how to build a semantic search API backed by MongoDB Atlas Vector Search.

## Prerequisites

- Go 1.21+
- Docker + Docker Compose
- MongoDB Atlas Local image (pulled automatically via compose)

## Setup

1. **Clone and install Go modules**

   ```bash
   git clone <repo-url> vector-database && cd vector-database
   go mod download
   ```

2. **Create your runtime config**

   ```bash
   cp config.template.yml config.yml
   # edit config.yml to point at your MongoDB instance (defaults target docker-compose)
   ```

3. **Bootstrap MongoDB Atlas Local**

   ```bash
   docker compose up -d mongo
   ```

   The Mongo container exposes `mongodb://xxx:xxxxx@localhost:27017/?authSource=admin&directConnection=true`.

4. **Run the API server**
   ```bash
   go run ./...
   ```
   The service ensures the vector index exists and then starts listening on `http://localhost:8080`.

## API

All routes are prefixed with `/api`.

| Method | Path             | Description                                   |
| ------ | ---------------- | --------------------------------------------- |
| POST   | `/messages`      | Insert a message + auto-generate vector       |
| GET    | `/messages`      | Semantic search over stored messages          |
| POST   | `/images`        | Insert an image + textual description         |
| POST   | `/images/search` | Find images whose embeddings are the closest  |

### Insert a Message

```bash
curl -X POST http://localhost:8080/api/messages \
  -H 'Content-Type: application/json' \
  -d '{
        "content": "Embeddings let search understand meaning.",
        "metadata": {"topic": "demo", "language": "en"}
      }'
```

Response:

```json
{
  "document": {
    "id": "67009e42b3f629343e58802a",
    "content": "Embeddings let search understand meaning.",
    "metadata": { "topic": "demo", "language": "en" }
  }
}
```

### Search for Messages

```bash
curl "http://localhost:8080/api/messages?q=semantic%20search&limit=3"
```

Response:

```json
{
  "query": "semantic search",
  "results": [
    {
      "id": "67009e42b3f629343e58802a",
      "content": "Embeddings let search understand meaning.",
      "metadata": { "topic": "demo", "language": "en" },
      "score": 0.91
    }
  ]
}
```

### Insert an Image

Upload JPEG or PNG files via `multipart/form-data`. Optional metadata must be a JSON string.

```bash
curl -X POST http://localhost:8080/api/images \
  -F "description=A beach sunset with orange sky." \
  -F 'metadata={"location":"Phuket","time":"18:30"}' \
  -F "image=@/path/to/sunset.jpg"
```

Response:

```json
{
  "image": {
    "id": "67009e42b3f629343e58802a",
    "description": "A beach sunset with orange sky.",
    "metadata": { "location": "Phuket", "time": "18:30" }
  }
}
```

### Search for Similar Images

```bash
curl -X POST http://localhost:8080/api/images/search \
  -F "image=@/path/to/sunset.jpg" \
  -F "limit=3"
```

Response:

```json
{
  "results": [
    {
      "id": "67009e42b3f629343e58802a",
      "description": "A beach sunset with orange sky.",
      "metadata": { "location": "Phuket", "time": "18:30" },
      "score": 0.89
    }
  ]
}
```

## Development

- `go test ./...` – compile/test all packages.
- `docker compose logs -f mongo` – inspect MongoDB for troubleshooting.
- Update `config.template.yml` when introducing new config fields; local `.gitignore` keeps `config.yml` untracked.
