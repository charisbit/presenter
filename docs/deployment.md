# 部署指南

> Intelligent Presenter 部署和配置指南

## 📋 系统要求

### 基础环境
- **操作系统**: Linux, macOS, Windows (WSL2)
- **Docker**: 20.10+ 
- **Docker Compose**: 2.0+
- **内存**: 最少 4GB RAM
- **存储**: 最少 10GB 可用空间

### 网络要求
- **端口**: 3003 (前端), 8081 (后端), 6379 (Redis), 3002 (语音服务)
- **防火墙**: 确保所需端口未被阻止
- **域名**: 生产环境建议配置域名和 SSL 证书

## 🚀 快速部署

### 1. 环境准备

```bash
# 克隆项目
git clone <repository-url>
cd intelligent-presenter

# 检查 Docker 环境
docker --version
docker-compose --version
```

### 2. 环境配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env
```

**必需的环境变量**:
```bash
# Backlog 配置
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=your-oauth-client-id
BACKLOG_CLIENT_SECRET=your-oauth-client-secret
BACKLOG_API_KEY=your-api-key

# OpenAI 配置
OPENAI_API_KEY=your-openai-api-key

# 可选配置
FRONTEND_BASE_URL=http://localhost:3003
JWT_SECRET=your-jwt-secret-key
```

### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 检查前端服务
curl http://localhost:3003/health

# 检查后端服务
curl http://localhost:8081/health

# 检查 Redis 服务
docker exec intelligent-presenter-redis redis-cli ping
```

## 🌐 生产环境部署

### 1. 域名配置

```bash
# 设置生产环境前端URL
export FRONTEND_BASE_URL=https://your-domain.com

# 更新 .env 文件
echo "FRONTEND_BASE_URL=https://your-domain.com" >> .env
```

### 2. SSL 证书配置

#### 使用 Let's Encrypt (推荐)

```bash
# 安装 certbot
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d your-domain.com

# 证书位置
/etc/letsencrypt/live/your-domain.com/fullchain.pem
/etc/letsencrypt/live/your-domain.com/privkey.pem
```

#### 配置 Nginx SSL

```nginx
# nginx/ssl/nginx.conf
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # 其他配置...
}
```

### 3. 反向代理配置

#### 使用 Nginx (推荐)

```nginx
# /etc/nginx/sites-available/intelligent-presenter
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL 配置...
    
    # 前端代理
    location / {
        proxy_pass http://localhost:3003;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # API 代理
    location /api/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket 代理
    location /ws/ {
        proxy_pass http://localhost:8081/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

#### 使用 Traefik

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  traefik:
    image: traefik:v2.10
    command:
      - --api.insecure=true
      - --providers.docker=true
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.email=your-email@example.com
      - --certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json
      - --certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    labels:
      - "traefik.enable=true"

  frontend:
    # ... 其他配置
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`your-domain.com`)"
      - "traefik.http.routers.frontend.entrypoints=websecure"
      - "traefik.http.routers.frontend.tls.certresolver=letsencrypt"
```

### 4. 环境变量管理

#### 使用 Docker Secrets

```bash
# 创建 secrets
echo "your-secret-value" | docker secret create backlog_domain -
echo "your-secret-value" | docker secret create openai_api_key -

# 在 docker-compose.yml 中使用
secrets:
  - backlog_domain
  - openai_api_key
```

#### 使用 .env 文件

```bash
# 生产环境 .env
NODE_ENV=production
FRONTEND_BASE_URL=https://your-domain.com
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=your-client-id
BACKLOG_CLIENT_SECRET=your-client-secret
BACKLOG_API_KEY=your-api-key
OPENAI_API_KEY=your-openai-key
JWT_SECRET=your-super-secret-jwt-key
```

## 🔧 配置调优

### 1. 性能优化

#### 前端优化

```nginx
# nginx.conf
# Gzip 压缩
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/javascript application/javascript application/json;

# 缓存配置
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

#### 后端优化

```go
// backend/cmd/main.go
// 连接池配置
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)

// Gin 配置
gin.SetMode(gin.ReleaseMode)
```

### 2. 安全配置

#### 防火墙设置

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# iptables
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -P INPUT DROP
```

#### 环境变量安全

```bash
# 设置文件权限
chmod 600 .env
chown root:root .env

# 定期轮换密钥
JWT_SECRET=$(openssl rand -base64 32)
```

## 📊 监控和维护

### 1. 健康检查

```bash
# 服务状态检查
docker-compose ps

# 日志监控
docker-compose logs -f backend
docker-compose logs -f frontend

# 资源使用情况
docker stats
```

### 2. 备份策略

```bash
# Redis 数据备份
docker exec intelligent-presenter-redis redis-cli BGSAVE

# 配置文件备份
tar -czf config-backup-$(date +%Y%m%d).tar.gz .env docker-compose.yml

# 日志备份
tar -czf logs-backup-$(date +%Y%m%d).tar.gz logs/
```

### 3. 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建并部署
docker-compose down
docker-compose up -d --build

# 清理旧镜像
docker image prune -f
```

## 🚨 故障排除

### 常见问题

#### 1. 端口冲突

```bash
# 检查端口占用
netstat -tulpn | grep :3003
netstat -tulpn | grep :8081

# 修改端口映射
# docker-compose.yml
ports:
  - "3004:3000"  # 改为 3004
```

#### 2. 服务启动失败

```bash
# 查看详细日志
docker-compose logs backend
docker-compose logs frontend

# 检查环境变量
docker-compose config

# 重启单个服务
docker-compose restart backend
```

#### 3. 网络连接问题

```bash
# 检查容器网络
docker network ls
docker network inspect presenter_intelligent-presenter-network

# 测试容器间通信
docker exec intelligent-presenter-frontend ping intelligent-presenter-backend
```

### 日志分析

```bash
# 实时日志监控
docker-compose logs -f --tail=100

# 错误日志过滤
docker-compose logs backend | grep ERROR
docker-compose logs frontend | grep error

# 日志文件位置
logs/
├── backend.log
├── frontend.log
└── speech.log
```

## 📚 参考资源

- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Nginx 配置指南](https://nginx.org/en/docs/)
- [Let's Encrypt 文档](https://letsencrypt.org/docs/)
- [Traefik 文档](https://doc.traefik.io/traefik/)

---

**最后更新**: 2024年12月  
**维护者**: 盛偉 (Sei I)
