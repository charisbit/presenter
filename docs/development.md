# 开发指南

> Intelligent Presenter 开发环境搭建和开发流程指南

## 🛠️ 开发环境搭建

### 基础要求

- **操作系统**: macOS, Linux, Windows (WSL2)
- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: 20.10+
- **Git**: 2.30+

### 环境检查

```bash
# 检查 Go 版本
go version

# 检查 Node.js 版本
node --version
npm --version

# 检查 Docker 版本
docker --version
docker-compose --version

# 检查 Git 版本
git --version
```

## 🚀 项目初始化

### 1. 克隆项目

```bash
git clone <repository-url>
cd intelligent-presenter

# 安装 Git hooks (可选)
git config core.hooksPath .githooks
```

### 2. 环境配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env
```

**开发环境配置示例**:
```bash
# 开发环境配置
NODE_ENV=development
FRONTEND_BASE_URL=http://localhost:3003

# Backlog 配置 (需要真实值)
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=your-oauth-client-id
BACKLOG_CLIENT_SECRET=your-oauth-client-secret
BACKLOG_API_KEY=your-api-key

# OpenAI 配置 (需要真实值)
OPENAI_API_KEY=your-openai-api-key

# JWT 密钥 (开发环境可以使用默认值)
JWT_SECRET=dev-jwt-secret-key
```

### 3. 依赖安装

```bash
# 后端依赖
cd backend
go mod tidy
go mod download

# 前端依赖
cd ../frontend
npm install

# 返回项目根目录
cd ..
```

## 🔧 开发模式启动

### 方式一：Docker Compose (推荐)

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 方式二：本地开发

```bash
# 终端 1: 启动后端
cd backend
go run cmd/main.go

# 终端 2: 启动前端
cd frontend
npm run dev

# 终端 3: 启动 Redis (可选)
docker run -d -p 6379:6379 redis:7-alpine
```

### 方式三：混合模式

```bash
# 启动依赖服务
docker-compose up -d redis backlog-mcp-server speech-mcp-server

# 本地启动前后端
cd backend && go run cmd/main.go &
cd frontend && npm run dev &
```

## 📁 项目结构

```
intelligent-presenter/
├── backend/                    # Go 后端服务
│   ├── cmd/
│   │   └── main.go           # 主程序入口
│   ├── internal/
│   │   ├── api/              # API 层
│   │   │   ├── handlers/     # 请求处理器
│   │   │   └── routes.go     # 路由配置
│   │   ├── auth/             # 认证中间件
│   │   ├── mcp/              # MCP 客户端
│   │   ├── models/           # 数据模型
│   │   └── services/         # 业务逻辑
│   ├── pkg/
│   │   └── config/           # 配置管理
│   ├── Dockerfile            # 后端容器配置
│   └── go.mod                # Go 模块文件
├── frontend/                  # TypeScript 前端
│   ├── src/
│   │   ├── components/       # Vue 组件
│   │   ├── services/         # API 服务
│   │   ├── stores/           # 状态管理
│   │   ├── types/            # TypeScript 类型定义
│   │   ├── views/            # 页面视图
│   │   ├── App.vue           # 主应用组件
│   │   └── main.ts           # 前端入口
│   ├── Dockerfile            # 前端容器配置
│   ├── nginx.conf            # Nginx 配置
│   ├── package.json          # 前端依赖
│   └── vite.config.ts        # Vite 配置
├── backlog-server/            # Backlog MCP 服务器
├── speech-server/             # Speech TTS 服务器
├── docs/                     # 项目文档
├── docker-compose.yml         # 容器编排配置
└── README.md                 # 项目说明
```

## 🔌 API 开发

### 后端 API 结构

#### 路由配置

```go
// backend/internal/api/routes.go
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
    api := router.Group("/api/v1")
    
    // 认证路由
    authHandler := handlers.NewAuthHandler(cfg)
    api.GET("/auth/login", authHandler.InitiateOAuth)
    api.GET("/auth/callback", authHandler.HandleCallback)
    api.GET("/auth/me", authHandler.GetUserInfo)
    
    // 幻灯片路由
    slideHandler := handlers.NewSlideHandler(cfg)
    api.POST("/slides/generate", slideHandler.GenerateSlides)
    api.GET("/slides/:id/status", slideHandler.GetSlideStatus)
    
    // 项目路由
    projectHandler := handlers.NewProjectHandler(cfg)
    api.GET("/projects", projectHandler.GetProjects)
    api.GET("/projects/:id/overview", projectHandler.GetProjectOverview)
}
```

#### 处理器开发

```go
// backend/internal/api/handlers/slide.go
type SlideHandler struct {
    config *config.Config
    mcpService *services.MCPService
}

func (h *SlideHandler) GenerateSlides(c *gin.Context) {
    var req models.SlideGenerationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 业务逻辑处理
    slides, err := h.mcpService.GenerateSlides(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, slides)
}
```

#### 中间件开发

```go
// backend/internal/auth/middleware.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // JWT 验证逻辑
        claims, err := auth.ValidateToken(token, config.JWTSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### 前端 API 调用

#### API 服务配置

```typescript
// frontend/src/services/api.ts
const api: AxiosInstance = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || '',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// 请求拦截器
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// 响应拦截器
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('auth_token')
            window.location.href = '/login'
        }
        return Promise.reject(error)
    }
)
```

#### API 调用示例

```typescript
// frontend/src/services/slidev.ts
export const slideApi = {
    async generateSlides(request: SlideGenerationRequest): Promise<SlideGenerationResponse> {
        const response = await api.post('/api/v1/slides/generate', request)
        return response.data
    },
    
    async getSlideStatus(slideId: string): Promise<any> {
        const response = await api.get(`/api/v1/slides/${slideId}/status`)
        return response.data
    }
}
```

## 🎨 前端开发

### Vue 组件开发

#### 组件结构

```vue
<!-- frontend/src/components/ChartComponent.vue -->
<template>
    <div class="chart-container">
        <canvas ref="chartCanvas"></canvas>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import Chart from 'chart.js/auto'

interface Props {
    config: any
}

const props = defineProps<Props>()
const chartCanvas = ref<HTMLCanvasElement>()

let chart: Chart | null = null

onMounted(() => {
    if (chartCanvas.value) {
        createChart()
    }
})

watch(() => props.config, () => {
    updateChart()
}, { deep: true })

function createChart() {
    if (!chartCanvas.value) return
    
    const ctx = chartCanvas.value.getContext('2d')
    if (!ctx) return
    
    chart = new Chart(ctx, {
        type: 'line',
        data: props.config.data,
        options: props.config.options
    })
}

function updateChart() {
    if (chart) {
        chart.data = props.config.data
        chart.update()
    }
}
</script>

<style scoped>
.chart-container {
    width: 100%;
    height: 400px;
    position: relative;
}
</style>
```

#### 状态管理

```typescript
// frontend/src/stores/slides.ts
import { defineStore } from 'pinia'
import { slideApi } from '@/services/slidev'
import type { SlideContent, SlideGenerationRequest } from '@/types/slides'

export const useSlidesStore = defineStore('slides', {
    state: () => ({
        slides: [] as SlideContent[],
        loading: false,
        error: null as string | null
    }),
    
    actions: {
        async generateSlides(request: SlideGenerationRequest) {
            this.loading = true
            this.error = null
            
            try {
                const response = await slideApi.generateSlides(request)
                this.slides = response.slides
                return response
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Unknown error'
                throw error
            } finally {
                this.loading = false
            }
        },
        
        async getSlideStatus(slideId: string) {
            try {
                const response = await slideApi.getSlideStatus(slideId)
                return response
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Unknown error'
                throw error
            }
        }
    }
})
```

### 样式和主题

#### CSS 变量

```css
/* frontend/src/style.css */
:root {
    /* 颜色主题 */
    --primary-color: #3b82f6;
    --secondary-color: #64748b;
    --success-color: #10b981;
    --warning-color: #f59e0b;
    --error-color: #ef4444;
    
    /* 字体 */
    --font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    --font-size-base: 16px;
    --font-size-sm: 14px;
    --font-size-lg: 18px;
    
    /* 间距 */
    --spacing-xs: 4px;
    --spacing-sm: 8px;
    --spacing-md: 16px;
    --spacing-lg: 24px;
    --spacing-xl: 32px;
    
    /* 圆角 */
    --border-radius-sm: 4px;
    --border-radius-md: 8px;
    --border-radius-lg: 12px;
}

/* 响应式断点 */
@media (max-width: 768px) {
    :root {
        --font-size-base: 14px;
        --spacing-md: 12px;
        --spacing-lg: 18px;
    }
}
```

#### 组件样式

```vue
<template>
    <div class="slide-card">
        <div class="slide-header">
            <h3 class="slide-title">{{ title }}</h3>
            <div class="slide-meta">
                <span class="slide-theme">{{ theme }}</span>
                <span class="slide-duration">{{ duration }}s</span>
            </div>
        </div>
        <div class="slide-content">
            <slot />
        </div>
    </div>
</template>

<style scoped>
.slide-card {
    background: white;
    border-radius: var(--border-radius-lg);
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    padding: var(--spacing-lg);
    margin-bottom: var(--spacing-md);
}

.slide-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-md);
}

.slide-title {
    font-size: var(--font-size-lg);
    font-weight: 600;
    color: var(--primary-color);
    margin: 0;
}

.slide-meta {
    display: flex;
    gap: var(--spacing-sm);
}

.slide-theme,
.slide-duration {
    font-size: var(--font-size-sm);
    color: var(--secondary-color);
    padding: var(--spacing-xs) var(--spacing-sm);
    background: #f1f5f9;
    border-radius: var(--border-radius-sm);
}
</style>
```

## 🧪 测试开发

### 后端测试

#### 单元测试

```go
// backend/internal/services/slide_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockMCPService struct {
    mock.Mock
}

func (m *MockMCPService) GenerateSlides(req models.SlideGenerationRequest) ([]models.SlideContent, error) {
    args := m.Called(req)
    return args.Get(0).([]models.SlideContent), args.Error(1)
}

func TestSlideService_GenerateSlides(t *testing.T) {
    mockMCP := new(MockMCPService)
    service := NewSlideService(mockMCP)
    
    req := models.SlideGenerationRequest{
        ProjectID: "test-project",
        Theme:     "project_overview",
    }
    
    expectedSlides := []models.SlideContent{
        {
            Title:   "项目概览",
            Content: "项目基本信息",
            Theme:   "project_overview",
        },
    }
    
    mockMCP.On("GenerateSlides", req).Return(expectedSlides, nil)
    
    slides, err := service.GenerateSlides(req)
    
    assert.NoError(t, err)
    assert.Equal(t, expectedSlides, slides)
    mockMCP.AssertExpectations(t)
}
```

#### 集成测试

```go
// backend/internal/api/handlers/slide_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestSlideHandler_GenerateSlides(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    handler := NewSlideHandler(&config.Config{})
    router.POST("/api/v1/slides/generate", handler.GenerateSlides)
    
    req := models.SlideGenerationRequest{
        ProjectID: "test-project",
        Theme:     "project_overview",
    }
    
    body, _ := json.Marshal(req)
    request := httptest.NewRequest("POST", "/api/v1/slides/generate", bytes.NewBuffer(body))
    request.Header.Set("Content-Type", "application/json")
    
    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)
    
    assert.Equal(t, http.StatusOK, recorder.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(recorder.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response, "slides")
}
```

### 前端测试

#### 组件测试

```typescript
// frontend/src/components/__tests__/ChartComponent.test.ts
import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import ChartComponent from '../ChartComponent.vue'

// Mock Chart.js
vi.mock('chart.js/auto', () => ({
    default: vi.fn().mockImplementation(() => ({
        update: vi.fn(),
        destroy: vi.fn()
    }))
}))

describe('ChartComponent', () => {
    it('renders chart with config', () => {
        const config = {
            data: {
                labels: ['Jan', 'Feb', 'Mar'],
                datasets: [{
                    label: 'Sales',
                    data: [10, 20, 30]
                }]
            },
            options: {
                responsive: true
            }
        }
        
        const wrapper = mount(ChartComponent, {
            props: { config }
        })
        
        expect(wrapper.find('canvas').exists()).toBe(true)
    })
    
    it('updates chart when config changes', async () => {
        const wrapper = mount(ChartComponent, {
            props: {
                config: {
                    data: { labels: ['A'], datasets: [] },
                    options: {}
                }
            }
        })
        
        await wrapper.setProps({
            config: {
                data: { labels: ['B'], datasets: [] },
                options: {}
            }
        })
        
        // 验证图表更新逻辑
        expect(wrapper.vm.chart).toBeDefined()
    })
})
```

#### 服务测试

```typescript
// frontend/src/services/__tests__/api.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { slideApi } from '../slidev'
import api from '../api'

// Mock axios
vi.mock('../api', () => ({
    default: {
        post: vi.fn(),
        get: vi.fn()
    }
}))

describe('slideApi', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })
    
    it('generates slides successfully', async () => {
        const mockResponse = {
            data: {
                slides: [
                    { title: 'Slide 1', content: 'Content 1' },
                    { title: 'Slide 2', content: 'Content 2' }
                ]
            }
        }
        
        vi.mocked(api.post).mockResolvedValue(mockResponse)
        
        const request = {
            projectId: 'test-project',
            theme: 'project_overview'
        }
        
        const result = await slideApi.generateSlides(request)
        
        expect(api.post).toHaveBeenCalledWith('/api/v1/slides/generate', request)
        expect(result).toEqual(mockResponse.data)
    })
    
    it('handles API errors', async () => {
        const error = new Error('API Error')
        vi.mocked(api.post).mockRejectedValue(error)
        
        const request = {
            projectId: 'test-project',
            theme: 'project_overview'
        }
        
        await expect(slideApi.generateSlides(request)).rejects.toThrow('API Error')
    })
})
```

## 🔄 开发工作流

### Git 工作流

```bash
# 1. 创建功能分支
git checkout -b feature/new-slide-theme

# 2. 开发功能
# ... 编写代码 ...

# 3. 提交代码
git add .
git commit -m "feat: add new slide theme for risk analysis"

# 4. 推送分支
git push origin feature/new-slide-theme

# 5. 创建 Pull Request
# 在 GitHub/GitLab 上创建 PR

# 6. 代码审查和合并
# 等待审查通过后合并到 main 分支
```

### 代码规范

#### Go 代码规范

```bash
# 格式化代码
go fmt ./...

# 运行 linter
golangci-lint run

# 运行测试
go test ./... -v

# 检查测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### TypeScript 代码规范

```bash
# 格式化代码
npm run format

# 运行 linter
npm run lint

# 运行类型检查
npm run type-check

# 运行测试
npm run test

# 检查测试覆盖率
npm run test:coverage
```

### 环境配置

#### 开发环境变量

```bash
# .env.development
NODE_ENV=development
FRONTEND_BASE_URL=http://localhost:3003
BACKLOG_DOMAIN=dev-domain.backlog.com
BACKLOG_CLIENT_ID=dev-client-id
BACKLOG_CLIENT_SECRET=dev-client-secret
BACKLOG_API_KEY=dev-api-key
OPENAI_API_KEY=dev-openai-key
JWT_SECRET=dev-jwt-secret
```

#### 生产环境变量

```bash
# .env.production
NODE_ENV=production
FRONTEND_BASE_URL=https://your-domain.com
BACKLOG_DOMAIN=your-domain.backlog.com
BACKLOG_CLIENT_ID=prod-client-id
BACKLOG_CLIENT_SECRET=prod-client-secret
BACKLOG_API_KEY=prod-api-key
OPENAI_API_KEY=prod-openai-key
JWT_SECRET=prod-jwt-secret
```

## 🚨 常见问题

### 后端问题

#### 1. 端口冲突

```bash
# 检查端口占用
lsof -i :8081
netstat -tulpn | grep :8081

# 修改端口
# backend/pkg/config/config.go
Port: getEnv("PORT", "8082"),

# 或者修改 docker-compose.yml
ports:
  - "8082:8080"
```

#### 2. 依赖问题

```bash
# 清理 Go 模块缓存
go clean -modcache
go mod tidy

# 更新依赖
go get -u all
go mod tidy
```

### 前端问题

#### 1. 构建失败

```bash
# 清理依赖
rm -rf node_modules package-lock.json
npm install

# 清理构建缓存
npm run clean
npm run build
```

#### 2. 热重载不工作

```bash
# 检查 Vite 配置
# frontend/vite.config.ts
server: {
    port: 3000,
    host: true,
    watch: {
        usePolling: true
    }
}
```

## 📚 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [Vue 3 文档](https://vuejs.org/)
- [TypeScript 文档](https://www.typescriptlang.org/)
- [Vite 文档](https://vitejs.dev/)
- [Gin 框架文档](https://gin-gonic.com/)
- [Chart.js 文档](https://www.chartjs.org/)
- [Mermaid 文档](https://mermaid-js.github.io/)

---

**最后更新**: 2024年12月
