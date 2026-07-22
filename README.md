# OneERP Enterprise AI Gateway

A production-ready, high-performance enterprise AI Platform written in Go 1.24+ that runs completely on-premises and exposes REST APIs for ERP applications.

> 📘 **Production Deployment**: See [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) for full guide on Docker Compose, Systemd daemon, Nginx SSL reverse proxy, and zero-downtime redeployment.

---

## Key Features

- **Multi-Server Ollama Load Balancing**: Includes a **Least-in-Flight** request distributor across multiple backend Ollama nodes to prevent any single GPU/server bottleneck.
- **YAML & Environment Configuration**: Configurable via `config/config.yaml` with environment variable overrides (`PORT`, `API_KEY`, `OLLAMA_URL`, etc.).
- **Plugin Architecture**:
  - **Profiles**: `email`, `support_ticket`, `inline_text`, `jira_story`.
  - **Actions**: `rewrite`, `summarize`, `translate`, `improve`, `expand`, `shorten`, `proofread`, `generate`, `create`.
- **Prompt Engine**: File-based markdown prompt renderer supporting dynamic placeholders (`{{TEXT}}`, `{{TITLE}}`, `{{CONVERSATION}}`, `{{TONE}}`, `{{LANGUAGE}}`, `{{SIGNATURE}}`, `{{CUSTOM_CONTEXT}}`).
- **Enterprise Middleware**: Bearer API Key authentication, IP rate limiting, Request ID tracing (`X-Request-ID`), Panic recovery, and structured logging.
- **Future Features Ready**: Interface stubs ready for Speech (Whisper/Piper), Document Summarization, OCR, and Qdrant RAG.

---

## High-Level Architecture

```text
ERP (OneERP Modules / Helpdesk)
  ↓
REST API (Bearer API Key Authentication)
  ↓
Gin Middleware (Request ID, Rate Limiter, Auth, Recovery)
  ↓
Write Controller / AI Service
  ↓
Profile Registry & Action Registry
  ↓
Markdown Prompt Engine (prompts/*/*.md)
  ↓
Multi-Server LLM Provider (Least-In-Flight Load Balancer)
  ┌───────────────┼───────────────┐
  ↓               ↓               ↓
Ollama Node 1   Ollama Node 2   Ollama Node N
 (port 11434)    (port 11435)    (...)
```

---

## Configuration (`config/config.example.yaml` → `config/config.yaml`)

Copy [`config/config.example.yaml`](config/config.example.yaml) to `config/config.yaml` to customize your environment:

```yaml
server:
  port: 8080
  request_timeout_seconds: 60
  max_payload_size_mb: 10
  log_level: "info"

security:
  api_key: "krea-secret-ai-key-2026"
  rate_limit:
    requests_per_minute: 100

llm:
  provider: "ollama" # "ollama" or "mock"
  default_model: "qwen3:8b"
  load_balancing_strategy: "least_in_flight"
  ollama_servers:
    - name: "ollama-node-1"
      url: "http://localhost:11434"
    - name: "ollama-node-2"
      url: "http://localhost:11435"
```

Environment variable overrides:
- `PORT`: HTTP listener port (e.g. `8080`)
- `API_KEY`: Secret Bearer API Key
- `OLLAMA_URL`: Comma-separated backend list (e.g. `http://ollama-1:11434,http://ollama-2:11434`)
- `MODEL`: Default LLM model (e.g. `qwen3:8b`)

---

## API Reference

### Health & Monitoring

- `GET /health` - Health check status and live multi-server backend statistics.
- `GET /version` - Engine version info.
- `GET /profiles` - List registered profiles (`["email", "support_ticket", "inline_text"]`).
- `GET /actions` - List registered actions (`["rewrite", "summarize", "translate", ...]`).
- `GET /models` - List available models and active backend node stats.

### Core Write API (`POST /api/v1/write`)

**Request**:

```json
POST /api/v1/write
Authorization: Bearer krea-secret-ai-key-2026
Content-Type: application/json

{
  "profile": "support_ticket",
  "action": "rewrite",
  "tone": "professional",
  "language": "english",
  "text": "issue fixed check now",
  "context": {
    "title": "Unable to Login",
    "conversation": [
      { "role": "user", "message": "Cannot login" },
      { "role": "support", "message": "Share username" },
      { "role": "user", "message": "abc123" }
    ]
  },
  "options": {
    "signature": "IT Support",
    "length": "medium"
  },
  "metadata": {
    "application": "OneERP",
    "module": "Helpdesk",
    "tenant_id": "KREA",
    "user_id": "1001"
  }
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "result": "Dear User,\n\nWe have resolved the login issue associated with your account (abc123). Please verify and let us know if you encounter any further difficulties.\n\nBest regards,\nIT Support",
    "profile": "support_ticket",
    "action": "rewrite",
    "model": "qwen3:8b",
    "processing_ms": 340
  }
}
```

### Jira Story Generation (`profile: "jira_story"`, `action: "generate"`)

**Request**:

```json
POST /api/v1/write
Authorization: Bearer krea-secret-ai-key-2026
Content-Type: application/json

{
  "profile": "jira_story",
  "action": "generate",
  "text": "Allow users to reset their password via SMS OTP verification"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "result": "Title: Password Reset via SMS OTP Verification\n\nDescription:\nAs a registered user,\nI want to reset my password using an SMS OTP code,\nSo that I can securely regain access to my account without email verification.\n\nBackground & Context:\nUsers often lose access to their registered email address. SMS OTP provides a secure alternative authentication channel.\n\nAcceptance Criteria:\n1. Given a user on the forgot password page, When they enter their registered mobile number, Then an 8-digit OTP is sent via SMS.\n2. Given the user receives the OTP, When they enter valid OTP within 5 minutes, Then they are prompted to set a new password.\n3. Given an invalid or expired OTP is entered, Then an appropriate error message is displayed.",
    "profile": "jira_story",
    "action": "generate",
    "model": "qwen2.5:0.5b",
    "processing_ms": 480
  }
}
```

---

## 🛠️ Management & Deployment Scripts

The repository includes automation scripts under `scripts/`:

### 1. `scripts/build.sh` (Clean Build)
Runs unit tests and compiles the executable binary to `bin/ai-gateway` with error checking:
```bash
./scripts/build.sh
```

### 2. `scripts/restart.sh` (Kill, Rebuild & Restart)
Kills any running `ai-gateway` process, cleans old binaries, verifies config, compiles a fresh binary, and launches the app in the background while checking for start errors:
```bash
./scripts/restart.sh
```

### 3. `scripts/setup_config.sh` (Config Generator)
Generates `config/config.yaml` from template `config/config.example.yaml` with interactive or automated (`-y`) parameter prompts and auto-random API Key generation:
```bash
./scripts/setup_config.sh -y
```

### 4. `scripts/benchmark_10_concurrent.sh` (10 Concurrent Requests Benchmark)
Fires 10 parallel HTTP requests against the gateway to measure multi-server load balancing performance and response latency for a target model:
```bash
./scripts/benchmark_10_concurrent.sh qwen2.5:0.5b
```

### 5. `scripts/benchmark_50_concurrent.sh` (50 Concurrent Requests Stress Test)
Stress tests the AI Gateway by sending 50 simultaneous parallel requests with 50 unique ERP text payloads across `email`, `support_ticket`, and `inline_text` profiles to verify zero-error throughput:
```bash
./scripts/benchmark_50_concurrent.sh qwen2.5:0.5b
```

### 6. `scripts/install_startup_service.sh` (Automatic Startup Service Installer)
Installs and enables the gateway to automatically launch on system boot (Systemd service daemon on Linux, LaunchAgent on macOS):
```bash
./scripts/install_startup_service.sh
```

### 7. `scripts/redeploy.sh` (Full Production Pull, Rebuild & Redeploy)
Pulls latest Git code, stops running service, cleans old binaries, runs unit tests, compiles new binary, restarts systemd service or background process, and checks `/health`:
```bash
./scripts/redeploy.sh
```

---

## 🔐 Git & Configuration Security (`.gitignore`)

To protect your API keys and production node URLs:
- **`config/config.example.yaml`**: Committed as the reference template.
- **`config/config.yaml`**: Excluded from git commits via `.gitignore`.
- **`bin/`**: Compiled binaries are excluded from git commits via `.gitignore`.

---

## Running Locally

```bash
# 1. Download dependencies & setup config
go mod tidy
./scripts/setup_config.sh

# 2. Run build and start
./scripts/build.sh
./bin/ai-gateway
```

## Running with Docker Compose

```bash
docker-compose up --build
```
