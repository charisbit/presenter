# Backlog API 完全参考文档

## 基本信息

### 基础URL
- .backlog.jp: `https://{space}.backlog.jp/api/v2`
- .backlog.com: `https://{space}.backlog.com/api/v2`

## 认证方式

### 1. API Key认证 (不推荐用于技术课题)
```
GET /api/v2/users/myself?apiKey=your_api_key
```

### 2. OAuth 2.0认证 (推荐)

#### 认可请求 (Authorization Request)
```
GET /OAuth2AccessRequest.action
```
参数:
- response_type: "code" (固定)
- client_id: アプリケーション登録で取得
- redirect_uri: リダイレクト先URL

#### 访问令牌请求 (Access Token Request)
```
POST /api/v2/oauth2/token
Content-Type: application/x-www-form-urlencoded
```
参数:
- grant_type: "authorization_code"
- code: 认可代码
- redirect_uri: 重定向URI

响应:
```json
{
  "access_token": "YOUR_ACCESS_TOKEN",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "YOUR_REFRESH_TOKEN"
}
```

#### API调用时的认证
```
GET /api/v2/space
Authorization: Bearer YOUR_ACCESS_TOKEN
```

#### 令牌更新
```
POST /api/v2/oauth2/token
Content-Type: application/x-www-form-urlencoded
```
参数:
- grant_type: "refresh_token"
- client_id: クライアントID
- client_secret: クライアントシークレット
- refresh_token: リフレッシュトークン

## 主要API端点

### 空间 (Space)
```
GET /api/v2/space                    // 空间信息
GET /api/v2/space/activities         // 空间活动
GET /api/v2/space/notification       // 空间通知
```

### 用户 (Users)
```
GET /api/v2/users                    // 用户列表
GET /api/v2/users/myself             // 当前用户信息
GET /api/v2/users/{userId}           // 指定用户信息
GET /api/v2/users/{userId}/activities // 用户活动
GET /api/v2/users/{userId}/recentlyViewedIssues // 最近查看的问题
GET /api/v2/users/{userId}/recentlyViewedProjects // 最近查看的项目
GET /api/v2/users/{userId}/recentlyViewedWikis // 最近查看的Wiki
```

### 项目 (Projects)
```
GET /api/v2/projects                 // 项目列表
GET /api/v2/projects/{projectIdOrKey} // 项目详情
GET /api/v2/projects/{projectIdOrKey}/activities // 项目活动
GET /api/v2/projects/{projectIdOrKey}/users // 项目用户
GET /api/v2/projects/{projectIdOrKey}/administrators // 项目管理员
POST /api/v2/projects                // 创建项目
PATCH /api/v2/projects/{projectIdOrKey} // 更新项目
DELETE /api/v2/projects/{projectIdOrKey} // 删除项目
```

### 问题/课题 (Issues)
```
GET /api/v2/issues                   // 问题列表
GET /api/v2/issues/{issueIdOrKey}    // 问题详情
POST /api/v2/issues                  // 创建问题
PATCH /api/v2/issues/{issueIdOrKey}  // 更新问题
DELETE /api/v2/issues/{issueIdOrKey} // 删除问题

GET /api/v2/issues/{issueIdOrKey}/comments // 问题评论
POST /api/v2/issues/{issueIdOrKey}/comments // 添加评论
PATCH /api/v2/issues/{issueIdOrKey}/comments/{commentId} // 更新评论

GET /api/v2/issues/{issueIdOrKey}/attachments // 问题附件
POST /api/v2/issues/{issueIdOrKey}/attachments // 添加附件
DELETE /api/v2/issues/{issueIdOrKey}/attachments/{attachmentId} // 删除附件
```

### 问题类型 (Issue Types)
```
GET /api/v2/projects/{projectIdOrKey}/issueTypes // 问题类型列表
POST /api/v2/projects/{projectIdOrKey}/issueTypes // 创建问题类型
PATCH /api/v2/projects/{projectIdOrKey}/issueTypes/{id} // 更新问题类型
DELETE /api/v2/projects/{projectIdOrKey}/issueTypes/{id} // 删除问题类型
```

### 优先级 (Priorities)
```
GET /api/v2/priorities               // 优先级列表
```

### 分类 (Categories)
```
GET /api/v2/projects/{projectIdOrKey}/categories // 分类列表
POST /api/v2/projects/{projectIdOrKey}/categories // 创建分类
PATCH /api/v2/projects/{projectIdOrKey}/categories/{id} // 更新分类
DELETE /api/v2/projects/{projectIdOrKey}/categories/{id} // 删除分类
```

### Wiki
```
GET /api/v2/wikis                    // Wiki列表
GET /api/v2/wikis/{wikiId}           // Wiki详情
POST /api/v2/wikis                   // 创建Wiki
PATCH /api/v2/wikis/{wikiId}         // 更新Wiki
DELETE /api/v2/wikis/{wikiId}        // 删除Wiki

GET /api/v2/wikis/{wikiId}/attachments // Wiki附件
POST /api/v2/wikis/{wikiId}/attachments // 添加Wiki附件
```

### 文件 (Files)
```
POST /api/v2/space/attachment        // 上传文件
GET /api/v2/space/attachment/{attachmentId} // 下载文件
```

### Git仓库
```
GET /api/v2/projects/{projectIdOrKey}/git/repositories // Git仓库列表
GET /api/v2/projects/{projectIdOrKey}/git/repositories/{repoIdOrName} // 仓库详情
GET /api/v2/projects/{projectIdOrKey}/git/repositories/{repoIdOrName}/pullRequests // PR列表
```

### 看板 (Watchings)
```
GET /api/v2/users/{userId}/watchings // 监视列表
POST /api/v2/watchings               // 添加监视
DELETE /api/v2/watchings/{watchingId} // 删除监视
```

## 响应数据结构

### 问题 (Issue) 对象
```json
{
  "id": 1,
  "projectId": 1,
  "issueKey": "BLG-1",
  "keyId": 1,
  "issueType": {
    "id": 2,
    "projectId": 1,
    "name": "Task",
    "color": "#7ea800",
    "displayOrder": 0
  },
  "summary": "问题摘要",
  "description": "问题描述",
  "resolution": null,
  "priority": {
    "id": 3,
    "name": "Normal"
  },
  "status": {
    "id": 1,
    "projectId": 1,
    "name": "Open",
    "color": "#ed8077",
    "displayOrder": 1000
  },
  "assignee": {
    "id": 2,
    "userId": "username",
    "name": "User Name",
    "roleType": 2,
    "lang": null,
    "mailAddress": "user@example.com",
    "lastLoginTime": "2022-09-01T06:35:39Z"
  },
  "category": [],
  "versions": [],
  "milestone": [],
  "startDate": null,
  "dueDate": null,
  "estimatedHours": null,
  "actualHours": null,
  "parentIssueId": null,
  "createdUser": { ... },
  "created": "2022-09-01T06:35:39Z",
  "updatedUser": { ... },
  "updated": "2022-09-01T06:35:39Z"
}
```

### 项目 (Project) 对象
```json
{
  "id": 1,
  "projectKey": "TEST",
  "name": "テストプロジェクト",
  "chartEnabled": true,
  "subtaskingEnabled": false,
  "projectLeaderCanEditProjectLeader": false,
  "textFormattingRule": "markdown",
  "archived": false
}
```

### 用户 (User) 对象
```json
{
  "id": 1,
  "userId": "admin",
  "name": "admin",
  "roleType": 1,
  "lang": "ja",
  "mailAddress": "eguchi@nulab.example",
  "lastLoginTime": "2013-05-30T09:11:36Z"
}
```

## 查询参数

### 常用查询参数
- `count`: 件数 (デフォルト: 20, 最大: 100)
- `offset`: オフセット (デフォルト: 0)
- `order`: ソート順 ("asc" または "desc")

### 问题列表查询参数
- `projectId[]`: プロジェクトID
- `issueTypeId[]`: 課題種別ID
- `categoryId[]`: カテゴリID
- `statusId[]`: 状態ID
- `priorityId[]`: 優先度ID
- `assigneeId[]`: 担当者ID
- `createdUserId[]`: 作成者ID
- `resolutionId[]`: 完了理由ID
- `parentChild`: 親子関係 (0, 1, 2, 3, 4)
- `attachment`: 添付ファイル有無
- `sharedFile`: 共有ファイル有無
- `sort`: ソートキー
- `keyword`: キーワード
- `createdSince`: 作成日時(開始)
- `createdUntil`: 作成日時(終了)
- `updatedSince`: 更新日時(開始)
- `updatedUntil`: 更新日時(終了)
- `startDateSince`: 開始日(開始)
- `startDateUntil`: 開始日(終了)
- `dueDateSince`: 期限日(開始)
- `dueDateUntil`: 期限日(終了)

## エラーレスポンス

### 認証エラー (401)
```json
{
  "errors": [
    {
      "message": "認証に失敗しました。",
      "code": 11,
      "moreInfo": ""
    }
  ]
}
```

### アクセス権限エラー (403)
```json
{
  "errors": [
    {
      "message": "アクセス権限がありません。",
      "code": 14,
      "moreInfo": ""
    }
  ]
}
```

### リソースが見つからない (404)
```json
{
  "errors": [
    {
      "message": "リソースが見つかりません。",
      "code": 6,
      "moreInfo": ""
    }
  ]
}
```

## 制限事項

### レート制限
- APIキー認証: 1時間あたり5000リクエスト
- OAuth認証: 1時間あたり5000リクエスト

### ペジネーション
- デフォルト件数: 20件
- 最大件数: 100件
- `count`と`offset`パラメーターでページ制御

### ファイルアップロード
- 最大ファイルサイズ: プランにより異なる
- 対応フォーマット: 画像、文書、アーカイブなど

## Webhook

### イベント種類
- 課題の追加・更新・削除
- コメントの追加
- Wikiの追加・更新・削除
- Gitのプッシュ・プルリクエスト
- ファイルの追加・更新・削除

### Webhook設定
プロジェクト設定画面で以下を設定:
- URL
- イベントタイプ
- 説明

### Webhook詳細情報
Webhookペイロードには以下の情報が含まれる:
- イベントタイプ
- プロジェクト情報
- 変更内容
- ユーザー情報
- タイムスタンプ

## 技術課題で活用できるAPI組み合わせ例

### プロジェクト健康度分析
```javascript
// 1. プロジェクト情報取得
GET /api/v2/projects/{projectId}

// 2. 課題一覧取得
GET /api/v2/issues?projectId[]={projectId}&count=100

// 3. プロジェクトアクティビティ取得
GET /api/v2/projects/{projectId}/activities?count=100

// 4. プロジェクトユーザー取得
GET /api/v2/projects/{projectId}/users
```

### チーム協働分析
```javascript
// 1. ユーザー一覧取得
GET /api/v2/users

// 2. 各ユーザーのアクティビティ取得
GET /api/v2/users/{userId}/activities

// 3. 課題のコメント取得
GET /api/v2/issues/{issueId}/comments

// 4. Wiki更新履歴取得
GET /api/v2/wikis?projectIdOrKey={projectId}
```

### 予測分析用データ収集
```javascript
// 1. 過去の課題データ取得
GET /api/v2/issues?projectId[]={projectId}&createdSince={date}&count=100

// 2. 課題の状態変更履歴取得
GET /api/v2/issues/{issueId}/comments

// 3. マイルストーン情報取得
GET /api/v2/projects/{projectId}/versions

// 4. Git活動取得
GET /api/v2/projects/{projectId}/git/repositories
```

## ベストプラクティス

### 認証
- OAuth 2.0を使用する
- アクセストークンは安全に保存
- リフレッシュトークンで自動更新

### パフォーマンス
- 必要なデータのみ取得
- ペジネーションを適切に使用
- キャッシュ機能を実装

### エラーハンドリング
- レート制限への対応
- ネットワークエラーのリトライ
- 適切なエラーメッセージ表示

### セキュリティ
- APIキーやトークンの適切な管理
- HTTPS通信の使用
- 入力値の検証とサニタイゼーション
