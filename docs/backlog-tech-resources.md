# Intelligent Presenter 技術リソース参考

## プロジェクト概要
**Intelligent Presenter for HTML Slides** の技術実装において参考となるBacklog関連技術リソース。
MCP協議とSlidevを活用したインテリジェント幻灯片生成システム。

### 核心機能
- Backlogプロジェクトデータから自動スライド生成
- 日語TTS による音声解説
- リアルタイム流式处理
- Vue 3 + Slidev による現代的プレゼンテーション体験

## 1. MCP (Model Context Protocol) 統合

### Backlog MCP Server 統合
- リポジトリ: https://github.com/nulab/backlog-mcp-server
- 公式 MCP Server 実装
- Docker コンテナ対応
- 非同期ストリーミング処理サポート

### Intelligent Presenter での利用
- Go バックエンドが MCP Client として接続
- プロジェクトデータの自動取得
- スライドテーマ別データ集約
- リアルタイム更新対応

### 対応 API エンドポイント (幻灯片生成用)
- getProject: プロジェクト概要取得
- getIssues: 課題データ取得  
- getUsers: チームメンバー情報
- getPullRequests: 開発活動状況
- getWikiPages: ドキュメント管理状況

### 使用例
```javascript
import { BacklogAPIClient } from 'backlog-js'

const client = new BacklogAPIClient({
  host: 'your-space.backlog.com',
  apiKey: 'your-api-key'
})

// プロジェクト一覧取得
const projects = await client.getProjects()

// 課題作成
const issue = await client.postIssue({
  projectId: 1,
  summary: 'New issue',
  issueTypeId: 1,
  priorityId: 3
})
```

### TypeScript支援
- 完全なTypeScript型定義
- IntelliSenseサポート
- 型安全なAPI呼び出し

## 2. backlog-mcp-server 詳細分析

### 概要
- Model Context Protocol (MCP) サーバー
- AI代理（Claude Desktop/Cline/Cursor等）とBacklog APIを接続
- 公式開発・保守
- TypeScript実装

### アーキテクチャ
```
AI Agent (Claude Desktop)
    ↕ MCP Protocol
backlog-mcp-server
    ↕ REST API
Backlog API
```

### コア機能

#### Toolset: space
- get_space: Backlogスペース情報取得
- get_users: ユーザー一覧取得
- get_myself: 認証ユーザー情報取得

#### Toolset: project
- get_project_list: プロジェクト一覧
- add_project: プロジェクト作成
- get_project: プロジェクト詳細
- update_project: プロジェクト更新
- delete_project: プロジェクト削除

#### Toolset: issue
- get_issue: 課題詳細取得
- get_issues: 課題一覧取得
- count_issues: 課題数カウント
- add_issue: 課題作成
- update_issue: 課題更新
- delete_issue: 課題削除
- get_issue_comments: 課題コメント取得
- add_issue_comment: 課題コメント追加
- get_priorities: 優先度一覧
- get_categories: カテゴリ一覧
- get_custom_fields: カスタムフィールド一覧
- get_issue_types: 課題種別一覧
- get_resolutions: 完了理由一覧
- get_watching_list_items: ウォッチ一覧
- get_watching_list_count: ウォッチ数

#### Toolset: wiki
- get_wiki_pages: Wiki一覧
- get_wikis_count: Wiki数カウント
- get_wiki: Wiki詳細
- add_wiki: Wiki作成

#### Toolset: git
- get_git_repositories: Gitリポジトリ一覧
- get_git_repository: Gitリポジトリ詳細
- get_pull_requests: プルリクエスト一覧
- get_pull_requests_count: プルリクエスト数
- get_pull_request: プルリクエスト詳細
- add_pull_request: プルリクエスト作成
- update_pull_request: プルリクエスト更新
- get_pull_request_comments: プルリクエストコメント
- add_pull_request_comment: プルリクエストコメント追加
- update_pull_request_comment: プルリクエストコメント更新

#### Toolset: notifications
- get_notifications: 通知一覧
- get_notifications_count: 通知数
- reset_unread_notification_count: 未読通知リセット
- mark_notification_as_read: 通知既読化

#### Toolset: document
- get_document_tree: ドキュメントツリー取得
- get_documents: ドキュメント一覧
- get_document: ドキュメント詳細

### 高度な機能

#### 設定可能なToolset
```bash
--enable-toolsets space,project,issue
```
環境変数: ENABLE_TOOLSETS="space,project,issue"

#### 動的Toolset発見
```bash
--dynamic-toolsets
```
環境変数: DYNAMIC_TOOLSETS=1

#### レスポンス最適化
```bash
--optimize-response
```
GraphQLスタイルのフィールド選択

#### トークン制限
```bash
--max-tokens=10000
```
デフォルト: 50,000トークン

#### ツール名プレフィックス
```bash
--prefix backlog_
```

### デプロイ方法

#### Docker（推奨）
```json
{
  "mcpServers": {
    "backlog": {
      "command": "docker",
      "args": [
        "run", "--pull", "always", "-i", "--rm",
        "-e", "BACKLOG_DOMAIN",
        "-e", "BACKLOG_API_KEY",
        "ghcr.io/nulab/backlog-mcp-server"
      ],
      "env": {
        "BACKLOG_DOMAIN": "your-domain.backlog.com",
        "BACKLOG_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### npx
```json
{
  "mcpServers": {
    "backlog": {
      "command": "npx",
      "args": ["backlog-mcp-server"],
      "env": {
        "BACKLOG_DOMAIN": "your-domain.backlog.com",
        "BACKLOG_API_KEY": "your-api-key"
      }
    }
  }
}
```

#### 手動インストール
```bash
git clone https://github.com/nulab/backlog-mcp-server.git
cd backlog-mcp-server
npm install
npm run build
```

### 多言語対応
- ホームディレクトリに.backlog-mcp-serverrc.json設置
- 環境変数でツール説明override
- 日本語設定例あり

### 使用例
```
「PROJECT-KEYプロジェクトの全課題を表示して」
「高優先度のバグ課題を作成して、担当者をuserに設定」
「mainブランチからfeature/new-featureでプルリクエスト作成」
```

## 3. 技術課題新方案建議

### 提案タイトル
「Intelligent Presenter for HTML Slides - MCP基盤のスライド生成システム」

### アーキテクチャ概要
```
TypeScript Frontend (Vue 3 + Slidev) ←→ Go Backend (MCP Gateway + REST API)
     (REST API + WebSocket)              ↕ MCP Protocol
                                    Backlog MCP Server + Speech MCP Server
                                         ↕ HTTP API
                                    Backlog API + OpenAI API + TTS
```

### コア機能設計

#### 1. Markdown Slide Generator (Go Backend)
**機能詳細:**
- MCPサーバー経由でBacklogデータ取得
- LLM APIを使用してMarkdown + Mermaid + Chart.jsスライド生成
- 10種類のスライドテーマ対応（プロジェクト概況、進度、課題など）
- 流式处理によるリアルタイム生成

**技術実装:**
- Go backend: MCP client実装
- OpenAI API: スライドコンテンツ生成
- Backlog MCP Server: プロジェクトデータ取得
- WebSocket: 流式スライド配信

#### 2. Markdown Slide Narrator (Go Backend)
**機能詳細:**
- Markdownスライドから口頭解説テキスト生成
- 日本語自然文処理
- Speech MCP Server連携
- RealtimeTTS による音声合成

**実装例:**
```
入力: プロジェクト概況スライド（Markdown + Chart.js）

出力: 「プロジェクトの概要を説明します。現在このプロジェクトには
      15個のタスクがあり、そのうち10個が完了しています。
      進捗率は67%で、予定通り今月末に完成予定です。」

音声: RealtimeTTS で日本語音声生成
```

#### 3. HTML Slide Compiler (TypeScript Frontend)
**機能詳細:**
- Slidev TypeScript API統合
- Markdown → HTML スライド変換
- Mermaid 図表レンダリング
- Chart.js グラフ埋め込み

**データ処理:**
- Slidev API: スライドコンパイル
- Mermaid: フローチャート・図表生成
- Chart.js: プロジェクトデータ可視化
- Vue 3: コンポーネント化UI

#### 4. HTML Slide Renderer (TypeScript Frontend)
**機能詳細:**
- ブラウザでのスライド表示
- Slidev参考のナビゲーション
- 音声同期再生
- レスポンシブデザイン対応

### 技術実装詳細

#### Go Backend設計
```go
package main

import (
    "github.com/gin-gonic/gin"
    "nulab-assistant/internal/mcp"
    "nulab-assistant/internal/ai"
)

type BacklogAssistant struct {
    mcpClient *mcp.Client
    aiService *ai.OpenAIService
    router    *gin.Engine
}

func (ba *BacklogAssistant) handleNaturalLanguageQuery(c *gin.Context) {
    // 1. 自然言語解析
    // 2. MCP tool呼び出し
    // 3. AI分析処理
    // 4. レスポンス生成
}
```

#### MCP Client実装
```go
type MCPClient struct {
    serverURL string
    tools     map[string]Tool
}

func (m *MCPClient) ExecuteTool(name string, params map[string]interface{}) (interface{}, error) {
    // MCP protocol implementation
}
```

#### TypeScript Frontend (Vue 3 + Slidev)
```typescript
interface SlideContent {
  markdown: string;
  theme: string;
  index: number;
  charts?: ChartConfig[];
  mermaid?: string;
}

interface SlideNarration {
  text: string;
  slideIndex: number;
  audioUrl?: string;
}

class PresentationAPI {
  async generateSlides(projectId: string, themes: string[]): Promise<string> {
    // Go backend API call for slide generation
  }
  
  connectWebSocket(slideId: string): WebSocket {
    // Real-time slide content streaming
  }
}
```

### 評価基準対応

#### 1. 視認性・可読性
- TypeScript型定義による型安全性
- Go構造体による明確なデータモデル
- React Componentの階層化設計
- 統一的なコーディングスタイル

#### 2. 保守性・拡張性
- MCP Client抽象化による疎結合
- プラグインアーキテクチャ設計
- 設定ファイルによる動的設定
- Docker Composeによる環境統一

#### 3. テスト意識
```
├── backend/
│   ├── internal/
│   │   ├── mcp/
│   │   │   ├── client.go
│   │   │   └── client_test.go
│   │   └── ai/
│   │       ├── service.go
│   │       └── service_test.go
│   └── integration_test.go
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   └── __tests__/
│   │   └── services/
│   │       └── __tests__/
│   └── e2e/
└── docker-compose.test.yml
```

#### 4. セキュリティ意識
- OAuth 2.0 + JWT認証
- CORS適切設定
- 入力値検証・サニタイゼーション
- APIレート制限実装
- secrets管理（Docker secrets）

### デプロイ構成

#### Docker Compose設定
```yaml
version: '3.8'
services:
  # TypeScript 前端 (Vue 3 + Slidev)
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    environment:
      - VITE_API_BASE_URL=http://localhost:8080
    depends_on:
      - backend
  
  # Go 後端服務 (MCP Gateway)
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
      - speech-mcp-server
  
  # Backlog MCP Server
  backlog-mcp-server:
    image: ghcr.io/nulab/backlog-mcp-server:latest
    environment:
      - BACKLOG_DOMAIN=${BACKLOG_DOMAIN}
      - BACKLOG_API_KEY=${BACKLOG_API_KEY}
    ports:
      - "3001:3000"
  
  # Speech MCP Server (Go TTS)
  speech-mcp-server:
    build: ./speech-mcp-server
    environment:
      - TTS_ENGINE=realtime-tts
      - LANGUAGE=ja
    ports:
      - "3002:3000"
```

### 独自価値提案

#### 1. 技術的優位性
- **MCP Protocol**: 最新AI-API連携標準の採用
- **Official Integration**: 公式ツール活用
- **Full-stack TypeScript/Go**: モダンな技術スタック
- **Real-time AI**: リアルタイムAI分析機能

#### 2. ビジネス価値
- **生産性向上**: 自然言語でのプロジェクト操作
- **リスク軽減**: 予測的問題検出
- **意思決定支援**: データドリブンな洞察
- **学習機能**: チームパフォーマンス継続改善

#### 3. 技術課題適合性
- **実力証明**: フルスタック開発能力
- **革新性**: AI × MCP × Backlogの組み合わせ
- **実用性**: 実際のプロジェクト管理課題解決
- **拡張性**: 将来的な機能追加容易

### 開発スケジュール（~8月17日）

#### 第一阶段 (天数 1-3): 基础架构搭建
**优先级 P0**
- Go 後端基础框架 (Gin + WebSocket)
- Backlog MCP Client 集成
- TypeScript 前端基础框架 (Vue 3 + Vite + Slidev)
- OAuth 2.0 認証実装

#### 第二阶段 (天数 4-6): 核心功能实现  
**优先级 P0**
- Markdown Slide Generator (Go)
  - プロジェクト概况主题实现
  - プロジェクト進度主题实现
  - 基础 LLM API 集成
- HTML Slide Compiler (TypeScript)
  - Slidev API 集成
  - Mermaid + Chart.js サポート
- HTML Slide Renderer (TypeScript)
  - 基础渲染功能
  - スライド切换逻辑

#### 第三阶段 (天数 7-8): 语音和高级功能
**优先级 P1**
- Speech MCP Server (Go TTS)
- Markdown Slide Narrator (Go)
- 日语文本分句処理
- 完整流式処理管道

#### 第四阶段 (天数 9-10): 测试和优化
**优先级 P1**
- 単体テスト覆盖
- 統合テスト
- パフォーマンス最適化
- ドキュメント完善

### 提出時のアピールポイント

#### 1. 技術選択の合理性
「最新公式MCPサーバーを活用し、SlidevによるHTML5スライド生成とAI統合における最先端アプローチを採用」

#### 2. 実用的価値  
「Backlogプロジェクトデータから自動的にプレゼンテーション資料を生成し、日本語音声解説付きで業務効率を大幅向上」

#### 3. 将来性
「MCPプロトコルの採用により、Claude DesktopやBacklog AI助手との統合が容易で、将来的な機能拡張に対応」

#### 4. イノベーション
「従来の静的レポートではなく、リアルタイム流式処理による動的プレゼンテーション生成システム」

#### 5. 技術スタック調和
「TypeScript(Vue 3) + Go の組み合わせで、フロントエンド・バックエンド両方の実力を demonstrate」

この提案は、技術方向性と課題要求を満たしつつ、候補者の技術力と創造性を総合的に demonstrate する内容となっています。
