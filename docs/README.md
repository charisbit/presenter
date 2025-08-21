# Intelligent Presenter for HTML Slides

> HTMLスライドのインテリジェントプレゼンター  
> HTML 幻灯片的智能演示者

基于 Backlog 项目数据自动生成智能化 HTML 幻灯片演示的系统，集成 MCP 协议、AI 内容生成和实时语音解说。

## 🎯 项目概述

**Intelligent Presenter** 是为 Nulab Backlog 平台设计的智能演示系统，能够：

- 📊 **自动数据获取**: 通过 MCP 协议从 Backlog 获取项目数据
- 🤖 **AI 内容生成**: 使用 LLM 自动生成幻灯片内容和解说词
- 🎙️ **实时语音合成**: 支持日语 TTS 语音解说
- 📱 **现代化展示**: 基于 Slidev 的响应式 HTML5 幻灯片

## 🏗️ 技术架构

```
TypeScript Frontend (Nginx) ←→ Go Backend ←→ MCP Servers
     (Port 3003)              (Port 8081)    (Internal)
```

### 核心组件

| 组件 | 技术栈 | 端口 | 职责 |
|------|--------|------|------|
| **前端** | TypeScript + Vue 3 + Slidev + Nginx | 3003 | 幻灯片编译、渲染和API代理 |
| **后端** | Go + Gin + WebSocket | 8081 | MCP 网关和业务逻辑 |
| **MCP Servers** | Backlog MCP + Speech TTS | 内部 | 数据获取和语音合成 |
| **Redis** | Redis 7 | 6379 | 缓存和会话存储 |

## 🚀 快速开始

### 环境要求

- Docker & Docker Compose
- Node.js 18+ (开发模式)
- Go 1.21+ (开发模式)
- Backlog 账户 + OAuth 2.0 应用

### 1. 克隆项目

```bash
git clone <repository-url>
cd intelligent-presenter
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```bash
# Backlog 配置
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=your-client-id
BACKLOG_CLIENT_SECRET=your-client-secret
BACKLOG_API_KEY=your-api-key

# OpenAI 配置
OPENAI_API_KEY=your-openai-key

# 前端基础URL (可选，默认使用localhost:3003)
FRONTEND_BASE_URL=http://localhost:3003
```

### 3. 启动服务

```bash
# 生产环境
docker-compose up -d

# 开发环境
cd frontend && npm run dev
cd backend && go run cmd/main.go
```

### 4. 访问应用

- **前端**: http://localhost:3003
- **后端健康检查**: http://localhost:8081/health
- **API 文档**: http://localhost:3003/api (通过前端代理)

## 📋 功能特性

### 🎨 10种幻灯片主题

1. **项目概况与基本信息** - 项目总览和团队介绍
2. **项目进度与状态** - 进度可视化和里程碑
3. **课题详情与管理** - 问题分析和处理状态
4. **项目风险与瓶颈预警** - 风险识别和预警
5. **团队成员与协作状态** - 团队效率分析
6. **文档与知识库管理** - 知识积累状态
7. **代码库与开发活动** - 开发进展和代码质量
8. **通知与沟通管理** - 沟通效率分析
9. **项目进度预测分析** - AI 驱动的预测分析
10. **总结与下一步计划** - 总结和规划

### 🤖 AI 智能化功能

- **内容生成**: LLM 自动生成幻灯片内容
- **语音解说**: 日语 TTS 自动生成解说词
- **数据可视化**: Mermaid + Chart.js 图表生成
- **实时更新**: 基于 WebSocket 的流式内容推送

## 🛠️ 开发指南

### 项目结构

```
intelligent-presenter/
├── backend/          # Go 后端服务 (Port 8081)
├── frontend/         # TypeScript 前端 (Port 3003)
├── backlog-server/   # Backlog MCP 服务器
├── speech-server/    # Speech TTS 服务器
├── docs/            # 项目文档
├── docker-compose.yml
└── README.md
```

### 端口配置

| 服务 | 容器内端口 | 宿主机端口 | 说明 |
|------|------------|------------|------|
| Frontend | 3000 | 3003 | 前端服务，包含Nginx代理 |
| Backend | 8080 | 8081 | 后端API服务 |
| Speech Server | 3001 | 3002 | 语音合成服务 |
| Redis | 6379 | 6379 | 缓存服务 |
| Backlog MCP | 3001 | - | 内部服务，不对外暴露 |

### 开发环境

```bash
# 后端开发
cd backend
go run cmd/main.go

# 前端开发
cd frontend
npm run dev

# 测试
npm run test
go test ./...
```

## 🌐 部署配置

### 本地开发

前端通过 Vite 代理访问后端：
```typescript
// vite.config.ts
proxy: {
  '/api': { target: 'http://localhost:8081' },
  '/ws': { target: 'ws://localhost:8081' }
}
```

### 生产环境

前端通过 Nginx 代理访问后端：
```nginx
# nginx.conf
location /api/ {
    proxy_pass http://intelligent-presenter-backend:8080;
}
location /ws/ {
    proxy_pass http://intelligent-presenter-backend:8080;
}
```

### 外网部署

1. **配置域名**: 设置 `FRONTEND_BASE_URL` 环境变量
2. **端口映射**: 将宿主机 3003 端口暴露到公网
3. **HTTPS**: 配置 SSL 证书（可选）

```bash
# 环境变量配置
export FRONTEND_BASE_URL=https://your-domain.com
docker-compose up -d
```

## 🧪 测试

### 运行测试

```bash
# 单元测试
npm run test:unit
go test ./... -v

# 集成测试
npm run test:integration

# E2E 测试
npm run test:e2e
```

### 测试覆盖率

- Go 后端: 80% 代码覆盖率
- TypeScript 前端: 70% 代码覆盖率

## 📚 文档

- [技术挑战总结](./nulab-tech-challenge-summary.md)
- [CLAUDE 开发记录](./CLAUDE.md)
- [部署指南](./deployment.md)
- [开发指南](./development.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支: `git checkout -b feature/amazing-feature`
3. 提交变更: `git commit -m 'Add some amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🏢 关于 Nulab

本项目是为 [Nulab Inc.](https://nulab.com/) 技术面试开发的演示项目，展示了对 Backlog 平台和现代 Web 技术的深度集成。

---

**盛偉 (Sei I)** - Nulab 技术面试项目  
📧 联系方式: [your-email@example.com]