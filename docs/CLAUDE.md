# 技术挑战信息汇总

## 技术要求
- 编程语言: 推荐 Java/Kotlin/Scala/Python/Go
- 特别要求: TypeScript 或 Go
- 认证: 必须使用 OAuth 2.0（不可使用API密钥）
- 部署: 需要可在本地运行
- 代码仓库: 使用Backlog的Git仓库

## 选择的方向

### 改进「Backlog AI助手」

**Intelligent Presenter for HTML Slides**
HTMLスライドのインテリジェントプレゼンター
HTML 幻灯片的智能演示者

### 系统架构设计

**MCP Client 和 MCP Servers 之间采用异步流式输入输出**

#### Presentation MCP Client
- **TypeScript 实现**：浏览器端用户界面
- **Golang 实现**：提供 REST API 的后端服务

#### MCP Servers
- **Backlog MCP Server**：Backlog 数据获取服务
- **Speech MCP Server**：采用 RealtimeTTS 的语音合成服务

### Presentation MCP Client 模块设计

各个模块之间采用异步流式进行交互：

#### Markdown Slide Generator
- 根据从 Backlog MCP Server 获取的信息（调用 LLM API）
- 生成一页用 Markdown + Mermaid + Chart.js 描述的 Slide

#### Markdown Slide Narrator
- 根据用 Markdown + Mermaid + Chart.js 描述的 Slide（调用 LLM API）
- 生成口头解说用纯文本

#### HTML Slide Compiler
- 调用 Slidev TypeScript API
- 根据用 Markdown + Mermaid + Chart.js 描述的 Slide 生成用 HTML 描述的 Slide

#### HTML Slide Renderer
- 参考 Slidev 实现
- 把用 HTML 描述的 Slide 在浏览器前端展示出来

### backlog-mcp-server 生成单页 Slide 候选主题

1. **项目概况与基本信息**

   * 项目名称、负责人、开始时间、当前状态
   * 关联空间（space）与团队成员简介

2. **项目进度与状态**

   * 当前进行中的任务数量、已完成任务数量
   * 进度百分比、关键里程碑状态
   * 任务分类、优先级分布

3. **课题（Issue）详情与管理**

   * 课题列表及筛选（按状态、优先级等）
   * 课题创建与更新情况
   * 课题评论和讨论动态

4. **项目风险与瓶颈预警**

   * 延期任务统计
   * 未解决的高优先级问题
   * 资源分配不均或团队负载过重提示

5. **团队成员与协作状态**

   * 团队成员在线情况与活跃度
   * 角色分布、权限设置
   * 协作模式分析与推荐（结合评论、任务分配等数据）

6. **文档与知识库管理**

   * Wiki 页面概览
   * 文档树结构与最近更新文档
   * 关联文档与任务的链接情况

7. **代码库与开发活动**

   * Git 仓库列表与状态
   * Pull Request 统计与状态
   * 代码评审与合并动态

8. **通知与沟通管理**

   * 未读通知统计
   * 最近通知摘要
   * 通知处理效率与提醒机制

9. **智能化辅助与预测分析**

   * 预测风险与项目健康度分析

### 推荐的多页 Slides 组织顺序示例

| 页码 | 主题         | 说明              |
| -- | ---------- | --------------- |
| 1  | 项目概况与基本信息  | 给听众对项目的整体印象     |
| 2  | 项目进度与状态    | 展示项目当前整体进展和任务分布 |
| 3  | 课题详情与管理    | 细化到任务层面，重点任务介绍  |
| 4  | 项目风险与瓶颈预警  | 提前识别风险，促进决策     |
| 5  | 团队成员与协作状态  | 介绍团队与协作现状       |
| 6  | 文档与知识库管理   | 知识积累与共享情况       |
| 7  | 代码库与开发活动   | 展示技术开发状态        |
| 8  | 通知与沟通管理    | 沟通效率及信息流转       |
| 9  | 项目进度预测分析 | 预测风险与项目健康度分析     |
| 10 | 总结与下一步计划   | 汇报总结，未来规划       |

## 核心功能特性

### 1. **智能幻灯片生成**
   - 基于 Backlog 项目数据自动生成 10 种主题幻灯片
   - Markdown + Mermaid + Chart.js 多模态内容组合
   - LLM 驱动的内容智能化生成

### 2. **实时语音解说**
   - 日语 TTS 自动生成口头解说
   - 智能文本分句和语音合成
   - 流式音频播放支持

### 3. **异步流式处理**
   - MCP 协议的实时数据获取
   - WebSocket 流式幻灯片推送
   - 渐进式内容生成和展示

### 4. **现代化演示体验**
   - Slidev 驱动的 HTML5 幻灯片
   - Vue 3 组件化交互界面
   - 响应式设计和动画效果

### 5. **项目洞察可视化**
   - 项目健康度综合分析
   - 风险预警和瓶颈识别  
   - 团队协作状态展示
   - 预测性项目管理支援

## 技术架构方案

### 核心系统架构
- **Presentation MCP Client**: 双端实现架构
  - **TypeScript 前端**: 浏览器端 UI 和交互
  - **Go 后端**: MCP 网关 + REST API 服务
- **Backlog MCP Server**: 使用官方 nulab/backlog-mcp-server
- **Speech MCP Server**: 基于 Go TTS 库实现语音合成

### MCP 交互设计
```
TypeScript Frontend ←→ Go Backend ←→ MCP Servers
     (REST API)        (MCP Protocol)
```

**设计原则**: 
- TypeScript 前端只通过 REST/WebSocket API 与 Go 后端交互
- Go 后端作为统一的 MCP 网关，处理所有 MCP 协议复杂性
- 避免前端直接访问 MCP Server，保持架构一致性

### 主要组件及技术分配

| 组件 | 实现语言 | 职责 | 交互方式 |
|------|----------|------|----------|
| **Markdown Slide Generator** | 🟦 Go | 数据聚合 + LLM 调用 | MCP → Backlog Server |
| **Markdown Slide Narrator** | 🟦 Go | 文本生成 + 日语处理 | MCP → Speech Server |
| **HTML Slide Compiler** | 🟨 TypeScript | Slidev API 调用 | REST API ← Go |
| **HTML Slide Renderer** | 🟨 TypeScript | 前端渲染 + 用户交互 | WebSocket ← Go |

### 数据流架构
```
Backlog Data → Go Backend (聚合+生成) → TypeScript Frontend (编译+渲染)
     ↓              ↓                          ↓
MCP Protocol    LLM API + TTS              Slidev + Vue 3
```

### 技术栈详细配置
- **前端**: TypeScript + Vue 3 + Slidev + Chart.js + Mermaid
- **后端**: Go + Gin + WebSocket + MCP Client Libraries
- **认证**: OAuth 2.0 + JWT (Backlog API 要求)
- **AI集成**: OpenAI API (LLM) + Go TTS (语音合成)
- **部署**: Docker + Docker Compose
- **协议**: REST API + WebSocket + MCP JSON-RPC 2.0

## MVP 实现计划

### 第一阶段 (天数 1-3): 基础架构搭建
**优先级 P0**
- [ ] Go 后端基础框架 (Gin + WebSocket)
- [ ] Backlog MCP Client 集成
- [ ] TypeScript 前端基础框架 (Vue 3 + Vite)
- [ ] OAuth 2.0 认证实现

### 第二阶段 (天数 4-6): 核心功能实现
**优先级 P0**
- [ ] Markdown Slide Generator (Go)
  - [ ] 项目概况主题实现
  - [ ] 项目进度主题实现
  - [ ] 基础 LLM API 集成
- [ ] HTML Slide Compiler (TypeScript)
  - [ ] Slidev API 集成
  - [ ] Mermaid + Chart.js 支持
- [ ] HTML Slide Renderer (TypeScript)
  - [ ] 基础渲染功能
  - [ ] 幻灯片切换逻辑

### 第三阶段 (天数 7-8): 语音和高级功能
**优先级 P1**
- [ ] Speech MCP Server (Go TTS)
- [ ] Markdown Slide Narrator (Go)
- [ ] 日语文本分句处理
- [ ] 完整的流式处理管道

### 第四阶段 (天数 9-10): 测试和优化
**优先级 P1**
- [ ] 单元测试覆盖
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档完善

## 技术风险评估

| 风险项 | 概率 | 影响 | 缓解策略 |
|--------|------|------|----------|
| Slidev API 集成复杂 | 中 | 中 | 提前技术验证，准备降级方案 |
| Go TTS 日语支持 | 高 | 中 | 使用云端 TTS API 作为备选 |
| MCP 协议实现 | 低 | 高 | 参考官方 backlog-mcp-server |
| OAuth 2.0 集成 | 低 | 中 | 使用成熟的 Go OAuth 库 |

## 下一步行动
1. ✅ 技术架构设计完成
2. 🔄 在 Backlog 项目中共享最新进展
3. 🔄 开始第一阶段开发 (基础架构)
4. 📝 准备技术 Demo 演示

## API 设计文档

### REST API 接口设计

#### 1. 项目数据获取
```http
GET /api/projects/{projectId}/overview
GET /api/projects/{projectId}/progress  
GET /api/projects/{projectId}/issues
GET /api/projects/{projectId}/team
GET /api/projects/{projectId}/risks
```

#### 2. 幻灯片生成
```http
POST /api/slides/generate
Content-Type: application/json

{
  "projectId": "PROJECT-123",
  "themes": ["overview", "progress", "issues"],
  "language": "ja"
}

Response: 
{
  "slideId": "slide-uuid",
  "status": "generating",
  "websocketUrl": "ws://localhost:8080/slides/slide-uuid"
}
```

#### 3. 流式幻灯片接收
```javascript
// WebSocket 连接
const ws = new WebSocket('ws://localhost:8080/slides/slide-uuid')

ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  
  switch(data.type) {
    case 'slide_content':
      // { markdown: "...", theme: "overview" }
      break
    case 'slide_narration':  
      // { text: "...", slideIndex: 1 }
      break
    case 'slide_audio':
      // { audioUrl: "/api/audio/...", slideIndex: 1 }
      break
    case 'presentation_complete':
      // { totalSlides: 5, duration: "120s" }
      break
  }
}
```

#### 4. 语音合成
```http
POST /api/speech/synthesize
Content-Type: application/json

{
  "text": "プロジェクトの概要を説明します",
  "language": "ja",
  "voice": "female",
  "streaming": true
}

Response: audio/wav stream
```

### WebSocket 事件协议

| 事件类型 | 数据格式 | 说明 |
|----------|----------|------|
| `slide_content` | `{markdown, theme, index}` | 生成的幻灯片内容 |
| `slide_narration` | `{text, slideIndex}` | 解说文本 |
| `slide_audio` | `{audioUrl, slideIndex}` | 语音文件 URL |
| `presentation_complete` | `{totalSlides, duration}` | 生成完成通知 |
| `error` | `{message, code}` | 错误信息 |

## 参考资料和API信息

### Backlog API
- 官方文档: https://developer.nulab.com/ja/docs/backlog/
- OAuth 2.0认证: https://developer.nulab.com/ja/docs/backlog/auth/#oauth-2-0
- MCP Server: https://github.com/nulab/backlog-mcp-server

### 技术参考
- Slidev 文档: https://sli.dev/guide/
- Go TTS 库: https://github.com/go-ego/gse (日语分词)
- Vue 3 + TypeScript: https://vuejs.org/guide/typescript/
- MCP 协议: https://modelcontextprotocol.io/

### 可获取数据
- 项目信息、问题、评论
- 用户活动、Git历史  
- Wiki内容、文件、里程碑
- 时间记录、自定义字段

## 部署和测试方案

### Docker 容器化部署

#### docker-compose.yml
```yaml
version: '3.8'

services:
  # Go 后端服务
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - BACKLOG_DOMAIN=${BACKLOG_DOMAIN}
      - BACKLOG_CLIENT_ID=${BACKLOG_CLIENT_ID}
      - BACKLOG_CLIENT_SECRET=${BACKLOG_CLIENT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      - backlog-mcp-server
    volumes:
      - ./logs:/app/logs

  # TypeScript 前端
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - VITE_API_BASE_URL=http://localhost:8080
    depends_on:
      - backend

  # Backlog MCP Server
  backlog-mcp-server:
    image: ghcr.io/nulab/backlog-mcp-server:latest
    environment:
      - BACKLOG_DOMAIN=${BACKLOG_DOMAIN}
      - BACKLOG_API_KEY=${BACKLOG_API_KEY}
    ports:
      - "3001:3000"

networks:
  default:
    name: presentation-network
```

#### 环境变量配置 (.env)
```bash
# Backlog 配置
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=your-client-id
BACKLOG_CLIENT_SECRET=your-client-secret
BACKLOG_API_KEY=your-api-key

# OpenAI 配置
OPENAI_API_KEY=your-openai-key

# 开发环境
NODE_ENV=development
GO_ENV=development
```

### 测试策略

#### 1. 单元测试
```bash
# Go 后端测试
cd backend
go test ./... -v -cover

# TypeScript 前端测试  
cd frontend
npm run test:unit
```

#### 2. 集成测试
```bash
# API 集成测试
go test ./tests/integration -v

# E2E 测试
npm run test:e2e
```

#### 3. 测试覆盖率目标
- **Go 后端**: 80% 代码覆盖率
- **TypeScript 前端**: 70% 代码覆盖率
- **API 集成**: 100% 关键路径覆盖

### 本地开发环境

#### 快速启动
```bash
# 1. 克隆项目
git clone <repository-url>
cd intelligent-presenter

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 3. 启动服务
docker-compose up -d

# 4. 访问应用
# 前端: http://localhost:3000
# 后端 API: http://localhost:8080
# API 文档: http://localhost:8080/swagger
```

#### 开发调试
```bash
# 后端开发模式
cd backend
go run main.go --dev

# 前端开发模式
cd frontend  
npm run dev

# 实时日志监控
docker-compose logs -f backend
```

### 期限管理
- 开发期间: ~8月17日（约10天）
- 夏季休假: 8月9日-17日

## 提交信息格式要求

### Git 提交信息规范
- 不要包含 "Generated with" 信息
- 不要包含 "Co-Author" 信息
- 使用简洁的描述性标题
- 使用项目符号列表描述具体修改内容
- 保持专业和技术性，避免不必要的元信息
