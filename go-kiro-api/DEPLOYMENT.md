# Deployment Guide

This guide covers various deployment options for the Go Kiro API.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Binary Deployment](#binary-deployment)
- [Docker Deployment](#docker-deployment)
- [Systemd Service](#systemd-service)
- [Cloud Deployment](#cloud-deployment)
- [Configuration](#configuration)

## Prerequisites

- Go 1.21 or later (for building from source)
- Valid Kiro OAuth credentials
- Optional: Docker for containerized deployment

## Binary Deployment

### Building from Source

```bash
# Clone the repository
git clone https://github.com/avldya/AIClient-2-API.git
cd AIClient-2-API/go-kiro-api

# Build for current platform
go build -o kiro-api cmd/server/main.go

# Or use the Makefile
make build
```

### Cross-Platform Builds

Build for multiple platforms at once:

```bash
make build-all
```

This creates:
- `kiro-api-linux-amd64` - Linux x86_64
- `kiro-api-windows-amd64.exe` - Windows x86_64
- `kiro-api-darwin-amd64` - macOS Intel
- `kiro-api-darwin-arm64` - macOS Apple Silicon

### Running the Binary

```bash
# With config file
./kiro-api -config config.json

# With environment variables
export KIRO_CREDS_BASE64="your-base64-credentials"
export SERVER_PORT=3000
./kiro-api

# With command line flags
./kiro-api -host 0.0.0.0 -port 3000 -api-key your-secret-key
```

## Docker Deployment

### Building Docker Image

```bash
# Build the image
docker build -t kiro-api:latest .

# Or use the Makefile
make docker-build
```

### Running with Docker

#### Basic Run

```bash
docker run -d \
  --name kiro-api \
  -p 3000:3000 \
  -e KIRO_CREDS_BASE64="your-base64-credentials" \
  -e KIRO_REGION="us-east-1" \
  kiro-api:latest
```

#### With Config File

```bash
docker run -d \
  --name kiro-api \
  -p 3000:3000 \
  -v $(pwd)/config.json:/root/config.json \
  kiro-api:latest -config /root/config.json
```

#### With Volume for Credentials

```bash
docker run -d \
  --name kiro-api \
  -p 3000:3000 \
  -v ~/.aws:/root/.aws:ro \
  -e KIRO_CRED_PATH=/root/.aws/sso/cache \
  kiro-api:latest
```

### Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  kiro-api:
    build: .
    image: kiro-api:latest
    container_name: kiro-api
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - KIRO_CREDS_BASE64=${KIRO_CREDS_BASE64}
      - KIRO_REGION=us-east-1
      - SERVER_API_KEY=${API_KEY}
    volumes:
      - ./config.json:/root/config.json:ro
    command: ["-config", "/root/config.json"]
```

Run with:

```bash
docker-compose up -d
```

## Systemd Service

For production Linux deployments, use systemd to manage the service.

Create `/etc/systemd/system/kiro-api.service`:

```ini
[Unit]
Description=Kiro API Server
After=network.target

[Service]
Type=simple
User=kiro
Group=kiro
WorkingDirectory=/opt/kiro-api
ExecStart=/opt/kiro-api/kiro-api -config /etc/kiro-api/config.json
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/kiro-api

# Environment
Environment="KIRO_CREDS_BASE64=your-credentials"
Environment="KIRO_REGION=us-east-1"

[Install]
WantedBy=multi-user.target
```

Setup and start:

```bash
# Create user
sudo useradd -r -s /bin/false kiro

# Create directories
sudo mkdir -p /opt/kiro-api /etc/kiro-api /var/log/kiro-api
sudo chown kiro:kiro /opt/kiro-api /var/log/kiro-api

# Copy binary and config
sudo cp kiro-api /opt/kiro-api/
sudo cp config.json /etc/kiro-api/
sudo chown kiro:kiro /opt/kiro-api/kiro-api /etc/kiro-api/config.json
sudo chmod 755 /opt/kiro-api/kiro-api
sudo chmod 600 /etc/kiro-api/config.json

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable kiro-api
sudo systemctl start kiro-api

# Check status
sudo systemctl status kiro-api

# View logs
sudo journalctl -u kiro-api -f
```

## Cloud Deployment

### AWS EC2

1. Launch an EC2 instance (t3.small recommended)
2. Install dependencies:

```bash
sudo yum update -y
sudo yum install -y git
```

3. Download and extract the binary:

```bash
wget https://github.com/avldya/AIClient-2-API/releases/download/v1.0.0/kiro-api-linux-amd64
chmod +x kiro-api-linux-amd64
sudo mv kiro-api-linux-amd64 /usr/local/bin/kiro-api
```

4. Create config file and systemd service (see above)
5. Configure security group to allow port 3000

### Google Cloud Run

Create `Dockerfile` (already included) and deploy:

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/PROJECT-ID/kiro-api

# Deploy to Cloud Run
gcloud run deploy kiro-api \
  --image gcr.io/PROJECT-ID/kiro-api \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars KIRO_CREDS_BASE64=your-credentials,KIRO_REGION=us-east-1
```

### Azure Container Instances

```bash
# Create resource group
az group create --name kiro-api-rg --location eastus

# Create container instance
az container create \
  --resource-group kiro-api-rg \
  --name kiro-api \
  --image your-registry/kiro-api:latest \
  --cpu 1 \
  --memory 1 \
  --ports 3000 \
  --environment-variables \
    KIRO_CREDS_BASE64=your-credentials \
    KIRO_REGION=us-east-1 \
  --dns-name-label kiro-api
```

### Kubernetes

Create deployment manifest `k8s-deployment.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kiro-api-secrets
type: Opaque
stringData:
  credentials: "your-base64-credentials"
  api-key: "your-api-key"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: kiro-api-config
data:
  config.json: |
    {
      "server": {
        "host": "0.0.0.0",
        "port": 3000
      },
      "kiro": {
        "region": "us-east-1",
        "requestMaxRetries": 3,
        "requestBaseDelay": 1000
      }
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kiro-api
spec:
  replicas: 2
  selector:
    matchLabels:
      app: kiro-api
  template:
    metadata:
      labels:
        app: kiro-api
    spec:
      containers:
      - name: kiro-api
        image: your-registry/kiro-api:latest
        ports:
        - containerPort: 3000
        env:
        - name: KIRO_CREDS_BASE64
          valueFrom:
            secretKeyRef:
              name: kiro-api-secrets
              key: credentials
        - name: SERVER_API_KEY
          valueFrom:
            secretKeyRef:
              name: kiro-api-secrets
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /root/config.json
          subPath: config.json
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: kiro-api-config
---
apiVersion: v1
kind: Service
metadata:
  name: kiro-api
spec:
  selector:
    app: kiro-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

Deploy:

```bash
kubectl apply -f k8s-deployment.yaml
```

## Configuration

### Environment Variables Priority

Configuration is loaded in the following priority (highest to lowest):

1. Command line flags (`-host`, `-port`, `-api-key`)
2. Environment variables (`SERVER_HOST`, `SERVER_PORT`, etc.)
3. Config file (`config.json`)
4. Default values

### Required Configuration

At minimum, you need to provide credentials via one of:

- `KIRO_CREDS_BASE64` environment variable
- `credsFilePath` pointing to a credentials file
- `credPath` pointing to a credentials directory

### Security Best Practices

1. **Never commit credentials** to version control
2. **Use environment variables** or secret management for production
3. **Restrict file permissions** on config files (chmod 600)
4. **Use HTTPS** when exposing to the internet (use reverse proxy like nginx)
5. **Set a strong API key** for authentication
6. **Regularly rotate** OAuth tokens
7. **Monitor** access logs for suspicious activity

### Reverse Proxy with Nginx

Example nginx configuration:

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/ssl/certs/yourdomain.crt;
    ssl_certificate_key /etc/ssl/private/yourdomain.key;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # For streaming responses
        proxy_buffering off;
        proxy_read_timeout 300s;
    }
}
```

## Monitoring and Logging

### Health Checks

The API provides a `/health` endpoint:

```bash
curl http://localhost:3000/health
```

### Logging

Logs are written to stdout/stderr. Capture them with:

```bash
# With systemd
journalctl -u kiro-api -f

# With Docker
docker logs -f kiro-api

# Direct binary
./kiro-api 2>&1 | tee kiro-api.log
```

### Metrics

Consider adding Prometheus metrics by extending the code with the `prometheus/client_golang` package.

## Troubleshooting

### Service Won't Start

1. Check credentials are valid
2. Verify port 3000 is not in use
3. Check file permissions
4. Review logs for error messages

### Connection Refused

1. Verify the service is running
2. Check firewall rules
3. Ensure correct host/port configuration

### Token Refresh Failures

1. Verify refresh token is still valid
2. Check network connectivity to AWS endpoints
3. Ensure system time is synchronized

### High Memory Usage

1. Check for memory leaks in streaming connections
2. Adjust connection pool settings
3. Monitor concurrent request count

## Performance Tuning

### System Limits

Increase open file limits for high traffic:

```bash
# /etc/security/limits.conf
kiro soft nofile 65536
kiro hard nofile 65536
```

### Go Runtime

Set GOMAXPROCS for optimal CPU usage:

```bash
export GOMAXPROCS=4  # Number of CPU cores
```

### Network Optimization

Tune TCP settings:

```bash
# /etc/sysctl.conf
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
net.ipv4.ip_local_port_range = 1024 65535
```

## Backup and Recovery

### Credentials Backup

Regularly backup your credentials directory:

```bash
tar -czf kiro-creds-backup-$(date +%Y%m%d).tar.gz ~/.aws/sso/cache
```

### Configuration Backup

Keep versioned copies of your configuration:

```bash
cp config.json config.json.backup-$(date +%Y%m%d)
```

## Support

For issues and questions:
- GitHub Issues: https://github.com/avldya/AIClient-2-API/issues
- Documentation: See README.md
