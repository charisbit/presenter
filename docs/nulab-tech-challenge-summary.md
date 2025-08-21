# Nulab技術課題提出物: Intelligent Presenter

## 提出プロジェクト情報
- 候補者: 盛偉 (Sei I)
- 職種: 【Backlog】ソフトウェアエンジニア（生成AI）
- 選択課題: 課題②「自由なアイデアでBacklogのAPIを使ったアプリケーションを作成」
- プロジェクト名: **Intelligent Presenter for HTML Slides**
- 使用技術: TypeScript + Go + MCP Protocol
- 期限: 8月17日（日）まで
- 評価期間: 提出後3営業日

## プロジェクト概要
BacklogプロジェクトデータからAI自動生成でHTML幻灯片演示システムを構築。
MCP協議、Slidev、日本語TTS統合による次世代プレゼンテーション体験。

## 課題要件

### 技術要件
- 言語: Java/Kotlin/Scala/Python/Go推奨
- 特別希望: TypeScript または Go
- 認証: OAuth 2.0 必須（APIキー不可）
- デプロイ: ローカルで動作可能
- リポジトリ: BacklogのGitリポジトリ使用

### 評価基準
1. 視認性・可読性を意識しているか
2. 保守性・拡張性を意識しているか  
3. テストを意識しているか
4. セキュリティを意識しているか

### 求められる「実力」
- ヌーラボで長期的にサービス開発・運用していくことを想定
- 書き捨てではなく実力がわかる内容
- 手を抜く箇所は明確に理由提示

## 会社情報

### Nulab概要
- 設立: 2004年3月
- 本社: 福岡県福岡市
- 上場: 東京証券取引所グロース市場（5033）
- 従業員: 全国24都道府県に分散（フルリモート）

### 主要サービス
- Backlog: プロジェクト・タスク管理（145万人利用）
- Cacoo: オンライン作図ツール（300万人利用）
- Nulab Pass: セキュリティ対策ツール

### 技術スタック
- Backlog: Scala (Play Framework)
- Cacoo: Go
- Typetalk: Scala (Play Framework)  
- Nulab Apps: Java (Spring), Kotlin
- バックオフィス: Python, Google Apps Script, Scala

### 企業文化・価値観
- ブランドメッセージ: 「このチームで一緒に仕事できてよかった」を世界中に生み出す
- 行動指針「Nuice Ways Ver 2.0」:
  - TRY FIRST: 常に学び、実践。常識にとらわれず挑戦
  - LOVE DIFFERENCES: 多様性を受け入れ、力に変える
  - GOAL ORIENTED: 目標を共有し、共に目的地を目指す

## 最新AI取り組み

### Backlog AI アシスタント（2025年7月β版開始）
- チャット型UIでBacklogデータを読み取り
- プロジェクトの状況整理・可視化
- 進捗レポート・ドキュメント自動生成
- 会話ベースでの実務支援
- リスク・ボトルネック事前察知

#### 機能特長
- プロジェクト情報蓄積・状況整理
- 進捗レポート・ドキュメント自動生成  
- 業務に応じた会話ベース支援
- リスク・ボトルネック予測

## 面接結果

### 一次面接企業フィードバック
- 明るい人柄、笑顔を交えた面接
- OSS活動への積極性、エンジニアカルチャーとのフィット
- AI関連経験と募集ポジションとの親和性
- 技術課題のアウトプットに期待

### 面接で聞かれた質問
- 転職動機
- チームワークに対する考え方
- 他メンバーから学んだ経験
- 普段の業務でのAI活用状況
- オープンソースプロジェクト参加のきっかけ
- 今後のキャリアプラン

## 提案アイディア

### 選択した方向性
「Intelligent Presenter for HTML Slides」- Backlog AI アシスタントの機能拡張

### 核心機能設計

#### 1. インテリジェントスライド生成システム
**10種類のスライドテーマ自動生成:**
- プロジェクト概況と基本情報
- プロジェクト進度と状態
- 課題詳細と管理
- プロジェクトリスクと瓶颈予警
- チームメンバーと協作状態
- ドキュメントと知識庫管理
- コードベースと開発活動  
- 通知とコミュニケーション管理
- 智能化輔助と予測分析
- 総括と次段階計画

#### 2. マルチモーダルコンテンツ生成
- **Markdown + Mermaid**: フローチャート、図表自動生成
- **Chart.js統合**: プロジェクトデータ可視化
- **日本語TTS**: RealtimeTTSによる音声解説
- **Slidev統合**: HTML5モダンプレゼンテーション

#### 3. リアルタイム流式処理
- **WebSocket**: 逐次スライド生成・配信
- **MCP Protocol**: 非同期データ取得
- **漸進的レンダリング**: ユーザー体験最適化

### 技術アーキテクチャ

#### システム構成
```
TypeScript Frontend (Vue 3 + Slidev) ↔ Go Backend (MCP Gateway + REST API)
     (REST API + WebSocket)              ↕ MCP Protocol
                                    Backlog MCP Server + Speech MCP Server
                                         ↕ HTTP API
                                    Backlog API + OpenAI API + TTS
```

#### 技術スタック
- **フロントエンド**: TypeScript + Vue 3 + Slidev + Chart.js + Mermaid
- **バックエンド**: Go + Gin + WebSocket + MCP Client Libraries
- **認証**: OAuth 2.0 + JWT (Backlog API要求)
- **AI統合**: OpenAI API (LLM) + RealtimeTTS (日本語音声合成)
- **プロトコル**: REST API + WebSocket + MCP JSON-RPC 2.0
- **デプロイ**: Docker + Docker Compose

#### 主要コンポーネント
- **Markdown Slide Generator** (Go): データ聚合 + LLM調用
- **Markdown Slide Narrator** (Go): 文本生成 + 日本語処理  
- **HTML Slide Compiler** (TypeScript): Slidev API調用
- **HTML Slide Renderer** (TypeScript): 前端渲染 + 用户交互

## 技術課題進行状況

### Backlogプロジェクト情報
- プロジェクト名: nulab-exam
- 招待メール受信済み
- 開発部長・Backlog開発チームメンバー参加
- メンション先: @Masashi Kotani @Kana Miyoshi

### 開発進捗 (8月12日時点)

#### 設計フェーズ完了 ✅
- 技術アーキテクチャ設計完成
- API設計ドキュメント作成
- MCP統合設計確定
- Docker Compose構成完成

#### MVP実装計画

**第一階段 (天数 1-3): 基础架构搭建 [優先度 P0]**
- [ ] Go後端基础框架 (Gin + WebSocket)
- [ ] Backlog MCP Client集成
- [ ] TypeScript前端基础框架 (Vue 3 + Vite)
- [ ] OAuth 2.0認証実装

**第二階段 (天数 4-6): 核心功能实现 [優先度 P0]**
- [ ] Markdown Slide Generator (Go)
- [ ] HTML Slide Compiler (TypeScript + Slidev)
- [ ] HTML Slide Renderer (TypeScript)
- [ ] 基本的なスライドテーマ実装

**第三階段 (天数 7-8): 语音和高级功能 [優先度 P1]**
- [ ] Speech MCP Server (Go TTS)
- [ ] Markdown Slide Narrator (Go)
- [ ] 日本語音声合成統合

**第四階段 (天数 9-10): 测试和优化 [優先度 P1]**
- [ ] 単体テスト実装
- [ ] 統合テスト
- [ ] ドキュメント完善

### 技術的リスク評価
- **Slidev API統合**: 中リスク → 技術検証済み
- **MCP Protocol実装**: 低リスク → 公式サーバー活用
- **日本語TTS**: 中リスク → RealtimeTTS採用
- **OAuth 2.0統合**: 低リスク → 成熟ライブラリ使用

## 参考資料・API情報

### Backlog API & MCP統合
- **公式ドキュメント**: https://developer.nulab.com/ja/docs/backlog/
- **OAuth 2.0認証**: https://developer.nulab.com/ja/docs/backlog/auth/#oauth-2-0
- **Backlog MCP Server**: https://github.com/nulab/backlog-mcp-server
- **MCP Protocol仕様**: https://modelcontextprotocol.io/

### 技術参考資料
- **Slidev**: https://sli.dev/guide/
- **Vue 3 + TypeScript**: https://vuejs.org/guide/typescript/
- **Go TTS Libraries**: RealtimeTTS, go-ego/gse (日本語分詞)
- **Chart.js**: https://www.chartjs.org/
- **Mermaid**: https://mermaid-js.github.io/

### 取得可能データ
- プロジェクト情報、課題、コメント
- ユーザーアクティビティ、Git履歴
- Wiki内容、ファイル、マイルストーン
- 時間記録、カスタムフィールド

### Nulab関連資料
- 決算説明資料: https://nulab.com/ja/ir/presentation/
- 技術課題ブログ: https://note.com/tatsuru_nulab/n/nbaed7f026683
- AI助手発表: https://nulab.com/ja/press/pr-2506-nulab-ai-by-backlog/

## 注意事項

### 避けるべきこと
- 既存AI助手機能の単純複製
- APIキー認証の使用
- セキュリティ軽視
- テスト不備

### 重視すべきこと
- 実用的な価値提供
- コード品質（可読性・保守性）
- 適切なアーキテクチャ設計
- セキュリティ・認証の実装
- 包括的なテストカバレッジ
- 明確なドキュメント

### 期限管理
- 開発期間: ~8月17日（約10日間）
- 夏季休業: 8月9日-17日（Nulab）
- 評価期間: 提出後3営業日
- 二次面接: 評価通過後設定
