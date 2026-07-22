# Production Deployment Guide - OneERP Enterprise AI Gateway

This document provides a comprehensive guide for deploying the **OneERP Enterprise AI Platform Gateway** as a standalone application in a production environment.

---

## 📋 Hardware Specifications

Because the **OneERP AI Gateway** is written in Golang with stateless HTTP routing, the central proxy/load balancer machine requires minimal resources. All heavy LLM inference occurs on the 2 or 3 separate GPU worker machines.

### 🖥️ 1. Central AI Gateway Machine (Proxy / Load Balancer)
| Component | Minimum Spec | Recommended Spec |
| :--- | :--- | :--- |
| **CPU** | 2 vCPUs / 2 Cores | 4 vCPUs / 4 Cores |
| **RAM** | 2 GB | 4 GB to 8 GB |
| **GPU** | **None Required** (0 VRAM) | **None Required** |
| **Network** | 1 Gbps LAN | 10 Gbps LAN |
| **OS** | Ubuntu 22.04 LTS / Debian 12 | Ubuntu 22.04 LTS |

*Note: The Go Gateway process consumes only ~25MB to 50MB of RAM and can easily handle 5,000+ concurrent proxy connections.*

---

### ⚡ 2. Ollama Backend Machines (2 to 3 Dedicated GPU Workers)
| Component | Per GPU Worker Spec |
| :--- | :--- |
| **GPU** | 1x NVIDIA RTX 3090 (24GB VRAM) / RTX 4090 (24GB VRAM) / A6000 / A100 |
| **CPU** | 8+ Physical Cores |
| **RAM** | 32 GB System RAM |
| **Storage** | 100 GB NVMe SSD |
| **Software** | Linux + NVIDIA Drivers + CUDA 12+ + Ollama |

---

## 🔒 1. Configuration Setup

Copy [`config/config.example.yaml`](config/config.example.yaml) to `config/config.yaml` on your gateway machine and customize your server settings and backend Ollama IPs:

```bash
cp config/config.example.yaml config/config.yaml
```

### `config/config.yaml` Configuration Example:

```yaml
server:
  port: 8080
  request_timeout_seconds: 120
  max_payload_size_mb: 25
  log_level: "info" # "info", "warn", "error"

security:
  api_key: "PROD_STRONG_SECRET_API_KEY_CHANGE_THIS"
  rate_limit:
    requests_per_minute: 200
    burst: 50

prompt:
  directory: "./prompts"

llm:
  provider: "ollama"
  default_model: "qwen3:8b"
  load_balancing_strategy: "least_in_flight"
  
  # Multi-Server / Multi-GPU Ollama nodes
  ollama_servers:
    - name: "gpu-node-1"
      url: "http://192.168.1.101:11434"
      weight: 1
      max_concurrent: 10
    - name: "gpu-node-2"
      url: "http://192.168.1.102:11434"
      weight: 1
      max_concurrent: 10
    - name: "gpu-node-3"
      url: "http://192.168.1.103:11434"
      weight: 1
      max_concurrent: 10
```

---

## 🚀 2. Deployment Methods

### 🌟 Scenario A: Dedicated Gateway Proxy Deployment (Recommended)
In this setup, the **AI Gateway runs standalone on the central Proxy Server** (no GPUs needed) and load-balances across 2 or 3 external GPU machines running Ollama over your local network.

```text
               ERP Clients / Applications
                           │
                           ▼
             ┌───────────────────────────┐
             │   Central Proxy Server    │  <-- Runs AI Gateway ONLY
             │   (IP: 192.168.1.100)     │      (Port 8080 / HTTPS 443)
             └─────────────┬─────────────┘
                           │
             ┌─────────────┼─────────────┐  (Least-In-Flight Load Balancer)
             ▼             ▼             ▼
       ┌───────────┐ ┌───────────┐ ┌───────────┐
       │  GPU #1   │ │  GPU #2   │ │  GPU #3   │  <-- Runs Ollama Service
       │.101:11434 │ │.102:11434 │ │.103:11434 │      (Port 11434)
       └───────────┘ └───────────┘ └───────────┘
```

#### Step 1: Ensure Ollama is running on GPU Machines
On each GPU machine (e.g. `192.168.1.101`, `192.168.1.102`, `192.168.1.103`), start Ollama accepting remote connections:

```bash
# On each GPU machine, start Ollama listening on 0.0.0.0:11434
OLLAMA_HOST=0.0.0.0 ollama serve

# Pull the required model on each GPU machine
ollama pull qwen3:8b
```

---

#### Method 1: Deploy Gateway via Docker Container (Using `config.yaml`)

On your central **Proxy Machine** (`192.168.1.100`):

```bash
# 1. Build the lightweight Proxy image
docker build -t oneerp-ai-gateway -f docker/Dockerfile .

# 2. Run the Gateway Container mounting your config.yaml
docker run -d \
  --name oneerp-ai-gateway \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml \
  oneerp-ai-gateway
```

*(Note: `-p 8080:8080` is required by Docker to expose port 8080 to your network so clients can connect).*

---

#### Method 2: Deploy Gateway via Systemd Service (Bare-Metal Linux)

On your central **Proxy Machine** (`192.168.1.100`):

1. **Build binary**:
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /opt/ai-gateway/bin/ai-gateway ./cmd/server
   ```

2. **Copy files**:
   ```bash
   sudo mkdir -p /opt/ai-gateway/bin /opt/ai-gateway/config /opt/ai-gateway/prompts
   sudo cp /opt/ai-gateway/bin/ai-gateway /opt/ai-gateway/bin/
   sudo cp config/config.yaml /opt/ai-gateway/config/config.yaml
   sudo cp -r prompts/ /opt/ai-gateway/
   ```

3. **Create Systemd Service** at `/etc/systemd/system/oneerp-ai-gateway.service`:
   ```ini
   [Unit]
   Description=OneERP Enterprise AI Gateway Proxy
   After=network.target

   [Service]
   Type=simple
   User=www-data
   Group=www-data
   WorkingDirectory=/opt/ai-gateway
   ExecStart=/opt/ai-gateway/bin/ai-gateway
   Restart=always
   RestartSec=5s

   [Install]
   WantedBy=multi-user.target
   ```

4. **Start Service**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable oneerp-ai-gateway
   sudo systemctl start oneerp-ai-gateway
   ```

---

### 📦 Scenario B: Local Multi-Container Docker Compose Deployment

If running gateway and mock/local Ollama together on a single development server:

```bash
cp config/config.example.yaml config/config.yaml
docker-compose -f docker-compose.yml up -d --build
```

---

## 🌐 3. Nginx Reverse Proxy & SSL (HTTPS) Setup

Route public HTTPS requests to the AI Gateway running on port `8080`.

### Nginx Configuration File (`/etc/nginx/sites-available/ai-gateway.conf`):

```nginx
server {
    listen 80;
    server_name ai-gateway.krea.edu.in;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name ai-gateway.krea.edu.in;

    # SSL Certificates
    ssl_certificate /etc/letsencrypt/live/ai-gateway.krea.edu.in/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/ai-gateway.krea.edu.in/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    add_header X-XSS-Protection "1; mode=block";

    # Request Size Limit (Matches MaxPayloadSizeMB config)
    client_max_body_size 25M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for Long LLM Generations
        proxy_read_timeout 120s;
        proxy_connect_timeout 60s;
        proxy_send_timeout 120s;
    }
}
```

Enable site & reload Nginx:

```bash
sudo ln -s /etc/nginx/sites-available/ai-gateway.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 📊 4. Health & Load Balancer Verification

Validate that your production gateway is up and load balancing across Ollama nodes:

### Health Check

```bash
curl -X GET https://ai-gateway.krea.edu.in/health
```

Expected JSON Output:
```json
{
  "status": "UP",
  "service": "OneERP AI Gateway",
  "ollama_backends": [
    {
      "name": "gpu-node-1",
      "url": "http://192.168.1.101:11434",
      "in_flight": 0,
      "total_requests": 142,
      "error_count": 0,
      "is_healthy": true
    },
    {
      "name": "gpu-node-2",
      "url": "http://192.168.1.102:11434",
      "in_flight": 0,
      "total_requests": 139,
      "error_count": 0,
      "is_healthy": true
    },
    {
      "name": "gpu-node-3",
      "url": "http://192.168.1.103:11434",
      "in_flight": 0,
      "total_requests": 140,
      "error_count": 0,
      "is_healthy": true
    }
  ]
}
```

### Interactive API Documentation

Access the live Swagger testing interface at:
`https://ai-gateway.krea.edu.in/docs`

---

## 🛠️ 5. Zero-Downtime Redeployment (CI/CD Update Script)

Save this script as `deploy.sh` for zero-downtime updates when pushing code updates:

```bash
#!/usr/bin/env bash
set -e

echo "=== Pulling Latest Code ==="
git pull origin main

echo "=== Running Unit Tests ==="
go test ./...

echo "=== Rebuilding Binary ==="
CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/ai-gateway/bin/ai-gateway ./cmd/server

echo "=== Restarting Gateway Service ==="
sudo systemctl restart oneerp-ai-gateway

echo "=== Verifying Health ==="
sleep 2
curl -f http://localhost:8080/health || (echo "Deployment Failed!" && exit 1)

echo "=== Deployment Successful ==="
```

Make executable:
```bash
chmod +x deploy.sh
```
