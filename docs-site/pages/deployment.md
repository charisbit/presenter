<!--
.. title: Deployment Guide
.. slug: deployment
.. date: 2025-08-18
.. tags: deployment, docker, production
.. category: 
.. link: 
.. description: Deployment guide for the Intelligent Presenter
.. type: text
-->

# Deployment Guide

This guide covers deployment strategies for the Intelligent Presenter system in various environments.

## System Requirements

### Minimum Hardware Requirements
- **CPU**: 2 cores, 2.0 GHz
- **RAM**: 4 GB (8 GB recommended)
- **Storage**: 20 GB free space
- **Network**: Stable internet connection for AI services

### Software Dependencies
- **Docker**: 20.10+ and Docker Compose
- **Go**: 1.21+ (for native builds)
- **Node.js**: 18+ (for frontend builds)
- **Git**: For source code management

## Environment Configuration

### Required Environment Variables

Create a `.env` file with the following variables:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production
FRONTEND_URL=http://localhost:3000

# Backlog OAuth Configuration
BACKLOG_DOMAIN=yourspace.backlog.jp
BACKLOG_CLIENT_ID=your_client_id
BACKLOG_CLIENT_SECRET=your_client_secret
OAUTH_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback

# AI Provider Configuration
AI_PROVIDER=openai  # or bedrock
OPENAI_API_KEY=your_openai_api_key

# AWS Configuration (if using Bedrock)
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=us-east-1

# MCP Service URLs
MCP_BACKLOG_URL=http://localhost:3001
MCP_SPEECH_URL=http://localhost:3002

# Security
JWT_SECRET_KEY=your_secure_jwt_secret_key

# Storage
SPEECH_CACHE_DIR=./audio-cache
```

### Production Environment Variables

```bash
# Production Server Configuration
PORT=80
ENVIRONMENT=production
FRONTEND_URL=https://your-domain.com

# HTTPS Configuration
TLS_CERT_PATH=/etc/ssl/certs/cert.pem
TLS_KEY_PATH=/etc/ssl/private/key.pem

# Database Configuration (if using)
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# Monitoring
LOG_LEVEL=info
METRICS_ENABLED=true
```

## Docker Deployment

### Using Docker Compose (Recommended)

1. **Clone the repository:**
```bash
git clone https://github.com/your-org/intelligent-presenter.git
cd intelligent-presenter
```

2. **Configure environment:**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Build and start services:**
```bash
docker-compose up -d --build
```

4. **Verify deployment:**
```bash
docker-compose ps
docker-compose logs -f
```

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENVIRONMENT=production
    env_file:
      - .env
    volumes:
      - ./audio-cache:/app/audio-cache
    depends_on:
      - backlog-server
      - speech-server

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - VITE_API_URL=http://backend:8080

  backlog-server:
    build: ./backlog-server
    ports:
      - "3001:3001"
    env_file:
      - .env

  speech-server:
    build: ./speech-server
    ports:
      - "3002:3002"
    env_file:
      - .env
    volumes:
      - ./audio-cache:/app/cache

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl
    depends_on:
      - frontend
      - backend
```

### Individual Container Deployment

#### Backend Container
```bash
cd backend
docker build -t intelligent-presenter-backend .
docker run -d \
  --name presenter-backend \
  -p 8080:8080 \
  --env-file ../.env \
  -v $(pwd)/../audio-cache:/app/audio-cache \
  intelligent-presenter-backend
```

#### Frontend Container
```bash
cd frontend
docker build -t intelligent-presenter-frontend .
docker run -d \
  --name presenter-frontend \
  -p 3000:3000 \
  -e VITE_API_URL=http://localhost:8080 \
  intelligent-presenter-frontend
```

## Native Deployment

### Backend Deployment

1. **Build the backend:**
```bash
cd backend
go mod download
go build -o intelligent-presenter cmd/main.go
```

2. **Create systemd service:**
```bash
sudo tee /etc/systemd/system/intelligent-presenter.service > /dev/null <<EOF
[Unit]
Description=Intelligent Presenter Backend
After=network.target

[Service]
Type=simple
User=presenter
WorkingDirectory=/opt/intelligent-presenter
ExecStart=/opt/intelligent-presenter/intelligent-presenter
EnvironmentFile=/opt/intelligent-presenter/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

3. **Enable and start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable intelligent-presenter
sudo systemctl start intelligent-presenter
```

### Frontend Deployment

1. **Build the frontend:**
```bash
cd frontend
npm install
npm run build
```

2. **Deploy with Nginx:**
```bash
sudo cp -r dist/* /var/www/html/
```

3. **Nginx configuration:**
```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

## Cloud Deployment

### AWS Deployment

#### Using ECS (Elastic Container Service)
1. **Create ECR repositories:**
```bash
aws ecr create-repository --repository-name intelligent-presenter-backend
aws ecr create-repository --repository-name intelligent-presenter-frontend
```

2. **Build and push images:**
```bash
$(aws ecr get-login --no-include-email)
docker tag intelligent-presenter-backend:latest $ECR_URI/intelligent-presenter-backend:latest
docker push $ECR_URI/intelligent-presenter-backend:latest
```

3. **Create ECS task definition:**
```json
{
  "family": "intelligent-presenter",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "containerDefinitions": [
    {
      "name": "backend",
      "image": "$ECR_URI/intelligent-presenter-backend:latest",
      "portMappings": [{"containerPort": 8080}],
      "environment": [
        {"name": "PORT", "value": "8080"},
        {"name": "ENVIRONMENT", "value": "production"}
      ]
    }
  ]
}
```

#### Using Lambda (Serverless)
For lighter workloads, deploy as AWS Lambda functions:

```bash
# Install serverless framework
npm install -g serverless

# Deploy backend as Lambda
cd backend
serverless deploy
```

### Google Cloud Platform

#### Using Cloud Run
```bash
# Build and deploy backend
gcloud builds submit --tag gcr.io/$PROJECT_ID/intelligent-presenter-backend
gcloud run deploy --image gcr.io/$PROJECT_ID/intelligent-presenter-backend --platform managed
```

#### Using App Engine
```yaml
# app.yaml
runtime: go121
env: standard

automatic_scaling:
  min_instances: 1
  max_instances: 10

env_variables:
  PORT: 8080
  ENVIRONMENT: production
```

### Kubernetes Deployment

#### Deployment Configuration
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: intelligent-presenter-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: presenter-backend
  template:
    metadata:
      labels:
        app: presenter-backend
    spec:
      containers:
      - name: backend
        image: intelligent-presenter-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        envFrom:
        - secretRef:
            name: presenter-secrets
```

#### Service Configuration
```yaml
apiVersion: v1
kind: Service
metadata:
  name: presenter-backend-service
spec:
  selector:
    app: presenter-backend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Security Configuration

### SSL/TLS Setup
```bash
# Using Let's Encrypt with Certbot
sudo certbot --nginx -d your-domain.com

# Manual certificate installation
sudo cp cert.pem /etc/ssl/certs/
sudo cp private.key /etc/ssl/private/
sudo chmod 600 /etc/ssl/private/private.key
```

### Firewall Configuration
```bash
# UFW configuration
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw --force enable
```

### Security Headers
```nginx
# Add to Nginx configuration
add_header X-Frame-Options DENY;
add_header X-Content-Type-Options nosniff;
add_header X-XSS-Protection "1; mode=block";
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
```

## Monitoring and Logging

### Application Monitoring
```yaml
# Prometheus configuration
scrape_configs:
  - job_name: 'intelligent-presenter'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Log Management
```bash
# Configure log rotation
sudo tee /etc/logrotate.d/intelligent-presenter > /dev/null <<EOF
/var/log/intelligent-presenter/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 presenter presenter
}
EOF
```

### Health Checks
```bash
# Backend health check
curl -f http://localhost:8080/health || exit 1

# Frontend health check
curl -f http://localhost:3000/ || exit 1
```

## Backup and Recovery

### Data Backup
```bash
# Backup audio cache
tar -czf audio-cache-backup-$(date +%Y%m%d).tar.gz audio-cache/

# Backup configuration
cp .env config-backup-$(date +%Y%m%d).env
```

### Database Backup (if applicable)
```bash
# PostgreSQL backup
pg_dump -h localhost -U user -d intelligent_presenter > backup.sql

# Restore from backup
psql -h localhost -U user -d intelligent_presenter < backup.sql
```

## Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Find process using port
sudo lsof -i :8080
sudo kill -9 <PID>
```

#### Memory Issues
```bash
# Monitor memory usage
free -h
docker stats

# Increase container memory limits
docker run -m 4g intelligent-presenter-backend
```

#### SSL Certificate Issues
```bash
# Verify certificate
openssl x509 -in cert.pem -text -noout

# Test SSL configuration
curl -vI https://your-domain.com
```

### Log Analysis
```bash
# View application logs
docker-compose logs -f backend
sudo journalctl -u intelligent-presenter -f

# Check error logs
grep ERROR /var/log/intelligent-presenter/app.log
```

## Performance Optimization

### Production Optimizations
- Enable gzip compression
- Configure CDN for static assets
- Implement Redis caching
- Optimize database queries
- Use connection pooling

### Scaling Considerations
- Horizontal scaling with load balancers
- Auto-scaling based on metrics
- Database read replicas
- Microservice architecture
- Caching layers