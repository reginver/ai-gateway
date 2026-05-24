# AI Gateway

A fork of [envoyproxy/ai-gateway](https://github.com/envoyproxy/ai-gateway) — an Envoy-based gateway for routing and managing AI/LLM API traffic.

## Overview

AI Gateway provides a unified interface for routing requests to multiple AI providers (OpenAI, Anthropic, Google Gemini, Ollama, etc.) with features like:

- **Unified API**: Single endpoint for multiple LLM backends
- **Load balancing**: Distribute requests across providers
- **Rate limiting**: Per-user and per-model rate controls
- **Cost tracking**: Monitor token usage and estimated costs
- **Failover**: Automatic fallback to secondary providers
- **Local models**: First-class support for Ollama and other local inference servers

## Prerequisites

- Go 1.22+
- Envoy Proxy (see [`.envoy-version`](.envoy-version) for the required version)
- Docker & Docker Compose (for local development)

## Quick Start

### Running with Ollama (local models)

1. Copy the Ollama environment file and adjust as needed:
   ```bash
   cp .env.ollama .env
   ```

2. Start the gateway with Ollama backend:
   ```bash
   docker compose --env-file .env.ollama up
   ```

3. Send a request:
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "llama3.2",
       "messages": [{"role": "user", "content": "Hello!"}]
     }'
   ```

> **Note:** I primarily use this with `llama3.2` and `mistral` locally. If Ollama isn't pulling the model automatically, run `ollama pull llama3.2` first.

### Running with cloud providers

1. Set your API keys in the environment:
   ```bash
   export OPENAI_API_KEY=sk-...
   export ANTHROPIC_API_KEY=sk-ant-...
   export GEMINI_API_KEY=...
   ```

2. Start the gateway:
   ```bash
   docker compose up
   ```

## Configuration

The gateway is configured via Envoy's xDS API or static configuration files. See the [`config/`](config/) directory for examples.

### Gemini

See [`.gemini/config.yaml`](.gemini/config.yaml) for Gemini-specific routing configuration.

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Running locally

```bash
go run ./cmd/ai-gateway/main.go --config config/default.yaml
```

## Architecture

```
Client → Envoy Proxy → AI Gateway Filter → Backend Provider
                              ↓
                    (routing, auth, rate limiting,
                     token counting, cost tracking)
```

The AI Gateway runs as an Envoy HTTP filter (via ext_proc or Lua) that intercepts requests, applies policies, rewrites headers/bodies for the target provider, and forwards the request.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Bug reports and feature requests are welcome via [GitHub Issues](../../issues).

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.
