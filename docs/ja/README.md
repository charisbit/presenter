# インテリジェント プレゼンター

BacklogプロジェクトデータからAIを活用してプロフェッショナルなスライドを自動生成する高度言語モデル統合システムです。

## 概要

インテリジェント プレゼンター システムは、Backlogプロジェクトデータを分析して包括的なプレゼンテーション スライドを自動生成します。AI技術を活用して、統合されたグラフ、図表、ナレーション機能を持つ経営陣向けスライドを作成します。

### 主要機能

- **AI駆動コンテンツ生成**: OpenAI GPTとAWS Bedrockを使用したインテリジェントなスライド内容作成
- **マルチテーマサポート**: プロジェクト管理の全側面をカバーする10の専門テーマ
- **リアルタイム生成**: WebSocketベースのスライド生成中のライブ更新
- **ビジュアル統合**: 自動Mermaid図表とChart.js可視化
- **音声ナレーション**: プレゼンテーション配信のためのテキスト読み上げ合成
- **多言語サポート**: 日本語と英語のコンテンツ生成
- **役員向けフォーマット**: 経営報告とステークホルダープレゼンテーション用に最適化

## アーキテクチャ

### システムコンポーネント

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   フロントエンド    │    │    バックエンド     │    │   外部API       │
│   (Vue.js)      │◄──►│    (Go)         │◄──►│                 │
│                 │    │                 │    │ • Backlog API   │
│ • スライド表示    │    │ • スライドサービス │    │ • OpenAI API    │
│ • ナビゲーション   │    │ • MCPサービス    │    │ • AWS Bedrock   │
│ • Chart.js      │    │ • AI統合        │    │ • TTSサービス    │
│ • Mermaid       │    │ • WebSocket     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 技術スタック

**フロントエンド:**
- Vue.js 3 with Composition API
- 状態管理用Pinia
- 型安全性のためのTypeScript
- データ可視化用Chart.js
- 図表レンダリング用Mermaid
- テスト用Vitest

**バックエンド:**
- Go 1.19+ with Ginフレームワーク
- リアルタイム通信用WebSocket
- AWS BedrockとOpenAI統合
- Backlogアクセス用MCP（Model Context Protocol）
- カスタムTTS統合

**インフラストラクチャ:**
- Dockerコンテナ化
- RESTful API設計
- リアルタイムWebSocket通信
- マルチプロバイダーAIフォールバックシステム

## スライドテーマ

システムは10の専門テーマをサポートします：

1. **プロジェクト概要** - 基本的なプロジェクト情報と目標
2. **プロジェクト進捗** - 完了率とマイルストーン追跡
3. **課題管理** - 課題追跡と解決メトリクス
4. **リスク分析** - リスク特定と軽減戦略
5. **チーム協力** - チーム活動とコミュニケーションパターン
6. **文書管理** - ドキュメンテーションと知識共有状況
7. **コードベース活動** - 開発メトリクスとコード品質指標
8. **通知管理** - コミュニケーション効率と情報フロー
9. **予測分析** - 予測と傾向分析
10. **総括・計画** - プロジェクト要約と将来の推奨事項

## はじめに

### 前提条件

- Node.js 18+
- Go 1.19+
- Docker（オプション）
- Backlog APIアクセス
- OpenAI APIキーまたはAWS認証情報

### インストール

1. **リポジトリをクローン**
   ```bash
   git clone <repository-url>
   cd intelligent-presenter
   ```

2. **バックエンドセットアップ**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # .envでAPIキーを設定
   go run main.go
   ```

3. **フロントエンドセットアップ**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **MCPサービスセットアップ**
   ```bash
   cd backlog-server
   npm install
   npm start
   
   cd ../speech-server
   npm install
   npm start
   ```

### 設定

backendディレクトリに`.env`ファイルを作成：

```env
# API設定
OPENAI_API_KEY=your_openai_key
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret
AWS_REGION=us-east-1

# サービスURL
MCP_BACKLOG_URL=http://localhost:8080
MCP_SPEECH_URL=http://localhost:8081

# サーバー設定
PORT=8000
CORS_ORIGINS=http://localhost:3000
```

## APIドキュメント

### RESTエンドポイント

#### スライド生成
```http
POST /api/slides/generate
Content-Type: application/json

{
  "projectId": "PROJECT_123",
  "themes": ["project_overview", "project_progress"],
  "language": "ja"
}
```

レスポンス:
```json
{
  "slideId": "uuid-string",
  "status": "generating",
  "websocketUrl": "ws://localhost:8000/ws/slides/uuid-string"
}
```

#### スライド状況取得
```http
GET /api/slides/{slideId}/status
```

### WebSocketイベント

接続先: `ws://localhost:8000/ws/slides/{slideId}`

**受信メッセージ:**
- `slide_content` - 新しいスライドが生成された
- `slide_narration` - ナレーションテキストが生成された
- `slide_audio` - 音声ファイルが生成された
- `presentation_complete` - すべてのスライドが完了
- `error` - 生成エラーが発生

## テスト

### バックエンドテスト（Go）
```bash
cd backend
go test ./tests/...
```

### フロントエンドテスト（Vitest）
```bash
cd frontend
npm run test
npm run test:coverage
```

### E2Eテスト
```bash
npm run test:e2e
```

## ドキュメント生成

### APIドキュメント生成
```bash
# Goドキュメント
cd backend
godoc -http=:6060

# TypeScriptドキュメント
cd frontend
npx typedoc --out docs src/
```

### 多言語ドキュメント
ドキュメントは3言語で利用可能：
- 英語: `/docs/en/`
- 日本語: `/docs/ja/`
- 中国語: `/docs/zh/`

## 開発

### プロジェクト構造
```
intelligent-presenter/
├── backend/                 # Goバックエンドサービス
│   ├── internal/           # 内部パッケージ
│   ├── pkg/               # 公開パッケージ
│   ├── tests/             # テストファイル
│   └── main.go
├── frontend/               # Vue.jsフロントエンド
│   ├── src/               # ソースコード
│   ├── tests/             # テストファイル
│   └── package.json
├── backlog-server/         # Backlog MCPサーバー
├── speech-server/          # 音声合成サーバー
└── docs/                  # ドキュメント
```

### 貢献

1. リポジトリをフォーク
2. 機能ブランチを作成
3. 新機能にテストを追加
4. すべてのテストが通ることを確認
5. プルリクエストを提出

### コードスタイル

- **Go**: 標準Go規約に従い`gofmt`を使用
- **TypeScript**: ESLintとPrettier設定を使用
- **コミット**: 従来のコミットメッセージを使用

## デプロイメント

### Dockerデプロイメント
```bash
docker-compose up -d
```

### 本番環境の考慮事項

- 適切な環境変数の設定
- CORSオリジンの設定
- SSL/TLS証明書の設定
- APIレート制限の監視
- 適切なログ実装
- ヘルスチェックの設定

## トラブルシューティング

### よくある問題

1. **APIレート制限**: システムはAIプロバイダー間の自動フォールバックを実装
2. **Mermaidレンダリング**: 三重バッククォートでの適切なコードブロック構文を確保
3. **Chart.js問題**: コードブロック内のJSON設定形式を検証
4. **WebSocket切断**: ネットワーク接続と認証を確認

### デバッグ

- バックエンドでデバッグログを有効化: `LOG_LEVEL=debug`
- フロントエンドデバッグにはブラウザ開発者ツールを使用
- データ取得問題についてはMCPサーバーログを確認

## ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。詳細はLICENSEファイルを参照してください。

## サポート

技術サポートや質問については：
- トラブルシューティングガイドを確認
- APIドキュメントを確認
- GitHubで問題を報告
- 開発チームに連絡