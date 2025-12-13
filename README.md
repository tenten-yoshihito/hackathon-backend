# Backend Architecture Documentation

## 📁 プロジェクト構成

```
hackathon-backend/
├── main.go                 # エントリーポイント、DI、ルーティング設定
├── controller/             # HTTPリクエストハンドラー
├── usecase/               # ビジネスロジック層
├── dao/                   # データアクセス層
├── model/                 # データモデル定義
├── middleware/            # 認証・ログなどのミドルウェア
└── db/                    # データベース接続設定
```

## 🏗️ アーキテクチャパターン

### レイヤードアーキテクチャ（3層）

```
Controller → Usecase → DAO → Database
    ↓          ↓        ↓
  HTTP      ビジネス   データ
 ハンドラ    ロジック   アクセス
```

### 依存関係の方向
- **Controller** は **Usecase** に依存
- **Usecase** は **DAO** に依存
- **DAO** は **Model** と **Database** に依存
- 逆方向の依存は禁止（依存性逆転の原則）

---

## 📂 ディレクトリ詳細

### 1. `controller/`
HTTPリクエストを受け取り、レスポンスを返す層

#### ファイル構成
- `helper.go` - 共通ヘルパー関数（`respondJSON`, `respondError`）
- `item_query_controller.go` - 商品の読み取り操作
- `item_command_controller.go` - 商品の書き込み操作
- `like_controller.go` - いいね機能
- `user_controller.go` - ユーザー管理
- `chat_controller.go` - チャット機能

#### 責務
- リクエストパラメータの取得
- バリデーション（基本的なもの）
- Usecaseの呼び出し
- HTTPレスポンスの生成

#### 依存関係
```
Controller
  ├─→ Usecase (ビジネスロジック呼び出し)
  ├─→ Middleware (認証情報取得)
  └─→ Model (レスポンス型定義)
```

---

### 2. `usecase/`
ビジネスロジックを実装する層

#### ファイル構成
- `item_list_usecase.go` - 商品一覧・検索
- `item_get_usecase.go` - 商品詳細取得
- `item_create_usecase.go` - 商品作成
- `item_update_usecase.go` - 商品更新
- `item_purchase_usecase.go` - 商品購入
- `like_usecase.go` - いいね機能
- `user_usecase.go` - ユーザー管理
- `chat_usecase.go` - チャット機能

#### 責務
- ビジネスルールの実装
- 入力値の詳細バリデーション
- トランザクション管理（必要に応じて）
- DAOの呼び出し

#### 依存関係
```
Usecase
  ├─→ DAO (データアクセス)
  └─→ Model (データ型定義)
```

---

### 3. `dao/`
データベースアクセスを担当する層

#### ファイル構成
- `item_dao.go` - 商品データアクセス
- `like_dao.go` - いいねデータアクセス
- `user_dao.go` - ユーザーデータアクセス
- `chat_dao.go` - チャットデータアクセス

#### 責務
- SQLクエリの実行
- データベーストランザクション管理
- 結果のマッピング（DB → Model）

#### 依存関係
```
DAO
  ├─→ database/sql (DB接続)
  └─→ Model (データ型定義)
```

#### 重要な実装パターン

**検索機能のセキュリティ対策**
```go
// LIKE特殊文字のエスケープ
func escapeLikeString(s string) string {
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, "%", "\\%")
    s = strings.ReplaceAll(s, "_", "\\_")
    return s
}

// 結果数の制限
query := `... LIMIT 100`
```

---

### 4. `model/`
データ構造を定義する層

#### ファイル構成
- `item.go` - 商品関連の型
- `like.go` - いいね関連の型
- `user.go` - ユーザー関連の型
- `chat.go` - チャット関連の型

#### 主要な型

**Item（商品）**
```go
type Item struct {
    ItemId        string    `json:"id"`
    UserId        string    `json:"user_id"`
    Name          string    `json:"name"`
    Price         int       `json:"price"`
    Description   string    `json:"description,omitempty"`
    ImageURLs     []string  `json:"image_urls"`
    Status        string    `json:"status"`
    SellerName    string    `json:"seller_name"`
    SellerIconURL string    `json:"seller_icon_url"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

**ステータス定数**
```go
const (
    StatusOnSale = "ON_SALE"
    StatusSold   = "SOLD"
)
```

---

### 5. `middleware/`
横断的関心事を処理

#### ファイル構成
- `auth.go` - Firebase認証ミドルウェア
- `cors.go` - CORS設定
- `logger.go` - ログ出力

#### 認証フロー
```
Request → AuthMiddleware → Controller
            ↓
      Firebase検証
            ↓
      Context.Set("userID")
```

---

### 6. `main.go`
アプリケーションのエントリーポイント

#### 責務
1. **依存性注入（DI）**
```go
// DAO初期化
itemDAO := dao.NewItemDAO(db)
likeDAO := dao.NewLikeDAO(db)

// Usecase初期化
itemList := usecase.NewItemList(itemDAO)
likeUsecase := usecase.NewLikeUsecase(likeDAO)

// Controller初期化
itemQueryController := controller.NewItemQueryController(itemList, ...)
likeController := controller.NewLikeController(likeUsecase)
```

2. **ルーティング設定**
```go
// 公開エンドポイント
e.GET("/items", itemQueryController.HandleItemList)
e.GET("/items/:id", itemQueryController.HandleItemDetail)

// 認証必須エンドポイント
auth := e.Group("")
auth.Use(middleware.AuthMiddleware(firebaseAuth))
auth.POST("/items", itemCommandController.HandleItemCreate)
auth.POST("/items/:id/like", likeController.ToggleLike)
```

---

## 🔄 データフロー例

### 商品検索のフロー

```
1. Client → GET /items?name=keyword

2. Controller (item_query_controller.go)
   ├─ クエリパラメータ取得: keyword
   └─ Usecase呼び出し: SearchItems(keyword)

3. Usecase (item_list_usecase.go)
   ├─ バリデーション: keyword != ""
   ├─ キーワードのトリミング
   └─ DAO呼び出し: SearchItems(keyword)

4. DAO (item_dao.go)
   ├─ LIKE特殊文字エスケープ
   ├─ SQLクエリ実行
   │  SELECT ... WHERE name LIKE ? LIMIT 100
   └─ 結果をModel.ItemSimpleにマッピング

5. Controller
   └─ JSON形式でレスポンス
```

### いいね機能のフロー

```
1. Client → POST /items/:id/like (認証必須)

2. Middleware (auth.go)
   ├─ Firebaseトークン検証
   └─ Context.Set("userID", uid)

3. Controller (like_controller.go)
   ├─ userID取得: GetUserIDFromContext()
   ├─ itemID取得: PathValue("id")
   └─ Usecase呼び出し: ToggleLike(userID, itemID)

4. Usecase (like_usecase.go)
   ├─ バリデーション
   └─ DAO呼び出し: ToggleLike(userID, itemID)

5. DAO (like_dao.go)
   ├─ 既存レコード確認
   ├─ 存在する → DELETE
   └─ 存在しない → INSERT
```

---

## 🔐 セキュリティ対策

### 1. 認証・認可
- Firebase Authentication使用
- JWTトークン検証
- ユーザーIDをContextに保存

### 2. SQLインジェクション対策
- パラメータ化クエリ使用
- LIKE特殊文字のエスケープ

### 3. DoS攻撃対策
- 検索結果を100件に制限
- ページネーション（将来実装推奨）

---

## 📊 データベーススキーマ

### 主要テーブル

**items（商品）**
```sql
CREATE TABLE items (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'ON_SALE',
    buyer_id VARCHAR(255),
    purchased_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_status_created (status, created_at DESC),
    INDEX idx_user_id (user_id)
);
```

**likes（いいね）**
```sql
CREATE TABLE likes (
    user_id VARCHAR(255) NOT NULL,
    item_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    PRIMARY KEY (user_id, item_id),
    INDEX idx_user_id (user_id),
    INDEX idx_item_id (item_id)
);
```

**users（ユーザー）**
```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    icon_url VARCHAR(500),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

---

## 🚀 起動方法

```bash
# 依存関係インストール
go mod download

# 環境変数設定
export FIREBASE_CREDENTIALS_PATH=/path/to/serviceAccountKey.json
export DB_USER=root
export DB_PASSWORD=password
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_NAME=uttc

# 起動
go run main.go
```

---

## 🧪 テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ確認
go test -cover ./...
```

---

## 📝 コーディング規約

### 命名規則
- **パッケージ**: 小文字、単数形（`controller`, `usecase`, `dao`）
- **インターフェース**: 大文字開始（`ItemDAO`, `ItemList`）
- **構造体**: 大文字開始（`Item`, `User`）
- **関数**: 大文字開始（公開）、小文字開始（非公開）

### エラーハンドリング
```go
// エラーをラップして返す
if err != nil {
    return fmt.Errorf("fail:operation: %w", err)
}
```

### ログ出力
```go
log.Printf("INFO: message")
log.Printf("ERROR: %v", err)
```

---

## 🔧 今後の改善提案

1. **パフォーマンス**
   - データベースインデックスの最適化
   - クエリキャッシング
   - コネクションプーリング調整

2. **セキュリティ**
   - レート制限の実装
   - 入力値サニタイゼーション強化

3. **機能**
   - ページネーション実装
   - 全文検索エンジン導入（Elasticsearch）
   - 画像最適化・CDN導入

4. **テスト**
   - ユニットテストカバレッジ向上
   - 統合テスト追加
   - E2Eテスト導入
