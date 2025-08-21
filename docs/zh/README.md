# 智能演示器

一个基于AI的演示文稿生成系统，使用先进的语言模型和实时可视化技术从Backlog项目数据创建专业幻灯片。

## 概述

智能演示器系统通过分析Backlog项目数据自动生成全面的演示幻灯片。它利用AI技术创建集成图表、图解和解说功能的高管级幻灯片。

### 核心功能

- **AI驱动内容生成**: 使用OpenAI GPT和AWS Bedrock进行智能幻灯片内容创建
- **多主题支持**: 涵盖项目管理各方面的10个专业主题
- **实时生成**: 基于WebSocket的幻灯片生成过程实时更新
- **可视化集成**: 自动Mermaid图表和Chart.js可视化
- **音频解说**: 用于演示传播的文本转语音合成
- **多语言支持**: 日语和英语内容生成
- **高管格式**: 针对管理报告和利益相关者演示进行优化

## 架构

### 系统组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     前端        │    │     后端        │    │    外部API      │
│   (Vue.js)      │◄──►│    (Go)         │◄──►│                 │
│                 │    │                 │    │ • Backlog API   │
│ • 幻灯片显示     │    │ • 幻灯片服务     │    │ • OpenAI API    │
│ • 导航          │    │ • MCP服务       │    │ • AWS Bedrock   │
│ • Chart.js      │    │ • AI集成        │    │ • TTS服务       │
│ • Mermaid       │    │ • WebSocket     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 技术栈

**前端:**
- Vue.js 3 with Composition API
- Pinia状态管理
- TypeScript类型安全
- Chart.js数据可视化
- Mermaid图表渲染
- Vitest测试

**后端:**
- Go 1.19+ with Gin框架
- WebSocket实时通信
- AWS Bedrock和OpenAI集成
- MCP（模型上下文协议）用于Backlog访问
- 自定义TTS集成

**基础设施:**
- Docker容器化
- RESTful API设计
- 实时WebSocket通信
- 多提供商AI故障转移系统

## 幻灯片主题

系统支持10个专业主题：

1. **项目概述** - 基本项目信息和目标
2. **项目进度** - 完成率和里程碑跟踪
3. **问题管理** - 问题跟踪和解决指标
4. **风险分析** - 风险识别和缓解策略
5. **团队协作** - 团队活动和沟通模式
6. **文档管理** - 文档和知识共享状态
7. **代码库活动** - 开发指标和代码质量指标
8. **通知管理** - 沟通效率和信息流
9. **预测分析** - 预测和趋势分析
10. **总结与规划** - 项目总结和未来建议

## 入门指南

### 先决条件

- Node.js 18+
- Go 1.19+
- Docker（可选）
- Backlog API访问权限
- OpenAI API密钥或AWS凭证

### 安装

1. **克隆仓库**
   ```bash
   git clone <repository-url>
   cd intelligent-presenter
   ```

2. **后端设置**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # 在.env中配置您的API密钥
   go run main.go
   ```

3. **前端设置**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **MCP服务设置**
   ```bash
   cd backlog-server
   npm install
   npm start
   
   cd ../speech-server
   npm install
   npm start
   ```

### 配置

在backend目录中创建`.env`文件：

```env
# API配置
OPENAI_API_KEY=your_openai_key
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret
AWS_REGION=us-east-1

# 服务URL
MCP_BACKLOG_URL=http://localhost:8080
MCP_SPEECH_URL=http://localhost:8081

# 服务器配置
PORT=8000
CORS_ORIGINS=http://localhost:3000
```

## API文档

### REST端点

#### 生成幻灯片
```http
POST /api/slides/generate
Content-Type: application/json

{
  "projectId": "PROJECT_123",
  "themes": ["project_overview", "project_progress"],
  "language": "ja"
}
```

响应:
```json
{
  "slideId": "uuid-string",
  "status": "generating",
  "websocketUrl": "ws://localhost:8000/ws/slides/uuid-string"
}
```

#### 获取幻灯片状态
```http
GET /api/slides/{slideId}/status
```

### WebSocket事件

连接到: `ws://localhost:8000/ws/slides/{slideId}`

**传入消息:**
- `slide_content` - 新幻灯片已生成
- `slide_narration` - 解说文本已生成
- `slide_audio` - 音频文件已生成
- `presentation_complete` - 所有幻灯片完成
- `error` - 生成错误发生

## 测试

### 后端测试（Go）
```bash
cd backend
go test ./tests/...
```

### 前端测试（Vitest）
```bash
cd frontend
npm run test
npm run test:coverage
```

### E2E测试
```bash
npm run test:e2e
```

## 文档生成

### 生成API文档
```bash
# Go文档
cd backend
godoc -http=:6060

# TypeScript文档
cd frontend
npx typedoc --out docs src/
```

### 多语言文档
文档提供三种语言版本：
- 英语: `/docs/en/`
- 日语: `/docs/ja/`
- 中文: `/docs/zh/`

## 开发

### 项目结构
```
intelligent-presenter/
├── backend/                 # Go后端服务
│   ├── internal/           # 内部包
│   ├── pkg/               # 公共包
│   ├── tests/             # 测试文件
│   └── main.go
├── frontend/               # Vue.js前端
│   ├── src/               # 源代码
│   ├── tests/             # 测试文件
│   └── package.json
├── backlog-server/         # Backlog MCP服务器
├── speech-server/          # 语音合成服务器
└── docs/                  # 文档
```

### 贡献

1. Fork仓库
2. 创建功能分支
3. 为新功能添加测试
4. 确保所有测试通过
5. 提交pull request

### 代码风格

- **Go**: 遵循标准Go约定并使用`gofmt`
- **TypeScript**: 使用ESLint和Prettier配置
- **提交**: 使用约定式提交消息

## 部署

### Docker部署
```bash
docker-compose up -d
```

### 生产环境考虑事项

- 设置适当的环境变量
- 配置CORS源
- 设置SSL/TLS证书
- 监控API速率限制
- 实现适当的日志记录
- 设置健康检查

## 故障排除

### 常见问题

1. **API速率限制**: 系统实现AI提供商之间的自动故障转移
2. **Mermaid渲染**: 确保使用三重反引号的正确代码块语法
3. **Chart.js问题**: 验证代码块中的JSON配置格式
4. **WebSocket断开**: 检查网络连接和身份验证

### 调试

- 在后端启用调试日志: `LOG_LEVEL=debug`
- 使用浏览器开发者工具进行前端调试
- 检查MCP服务器日志以解决数据检索问题

## 许可证

本项目基于MIT许可证。详情请参阅LICENSE文件。

## 支持

技术支持或问题：
- 查看故障排除指南
- 查看API文档
- 在GitHub上提交问题
- 联系开发团队