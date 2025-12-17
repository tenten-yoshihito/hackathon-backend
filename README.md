# Backend Architecture Documentation

##  プロジェクト構成

```
hackathon-backend/
├── main.go                # エントリーポイント、DI、ルーティング設定
├── controller/            # HTTPリクエストハンドラー
├── usecase/               # ビジネスロジック層
├── dao/                   # データアクセス層
├── model/                 # データモデル定義
├── middleware/            # 認証・ログなどのミドルウェア
└── db/                    # データベース接続設定
```

## アーキテクチャパターン

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
- 逆方向の依存は禁止

---

## ディレクトリ詳細

### 1. `controller/`
HTTPリクエストを受け取り、レスポンスを返す層(url情報もここで取る)

#### ファイル構成
- `helper.go` - 共通ヘルパー関数（`respondJSON`, `respondError`）
- `item_query_controller.go` - 商品の読み取り操作
- `item_command_controller.go` - 商品の書き込み操作
- `like_controller.go` - いいね機能
- `user_controller.go` - ユーザー管理
- `chat_controller.go` - チャット機能

#### 責務
- リクエストパラメータの取得
- バリデーション（useridなど基本的なもの）
- Usecaseの呼び出し
- HTTPレスポンスの生成(respondJson/respondErrorでusecaseからの返り値orエラーを返す)

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
- `chat_usecase.go` - チャット機能(チャットルームの作成・取得、メッセージの送信・取得)->責任分離の観点から微妙かも

- `description_generate_usecase.go` - imageURLのバリデーションと商品説明文の生成

- `item_detail_usecase.go` - 商品詳細取得
- `item_get_usecase.go` - 特定の商品の取得
- `item_list_usecase.go` - 商品一覧取得(home画面用)
- `item_purchase_usecase.go` - 商品購入処理(soldにするだけ)
- `item_register_usecase.go` - 商品登録
- `item_update_usecase.go` - 商品更新(削除は未実装)

- `like_usecase.go` - いいね機能

- `my_items_list_usecase.go` - 特定のユーザーの出品商品一覧取得(名前はかなり怪しくて別にログインしているユーザー以外のものも取得できる)

- `user_get_usecase.go` - 特定のユーザーの取得
- `user_items_list_usecase.go` - 特定のユーザーの出品商品一覧取得
- `user_register_usecase.go` - ユーザー登録
- `user_search_usecase.go` - ユーザー一覧を取得,カリキュラムの名残なので使用していないが一応残している
- `user_update_usecase.go` - ユーザー更新(削除は未実装)


#### 責務
- ビジネスルールの実装
- 入力値の詳細バリデーション
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
- dbの接続は依存性の注入の観点からmainで行った。
- SQLクエリの実行
- データベーストランザクション管理(商品購入処理など)
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

```go
type ChatRoom struct {
	Id        string    `json:"id"`
	ItemId    string    `json:"item_id"`
	BuyerId   string    `json:"buyer_id"`
	SellerId  string    `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	Id         string    `json:"id"`
	ChatRoomId string    `json:"chat_room_id"`
	SenderId   string    `json:"sender_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}


type Item struct {
	ItemId        string    `json:"id"`
	UserId        string    `json:"user_id"`
	Name          string    `json:"name"`
	Price         int       `json:"price"`
	Description   string    `json:"description,omitempty"`
	ImageURLs     []string  `json:"image_urls"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	SellerName    string    `json:"seller_name"`
	SellerIconURL string    `json:"seller_icon_url"`
}

type User struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	IconURL   string    `json:"icon_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

```



---

### 5. `middleware/`
横断的関心事を処理

#### ファイル構成
- `auth.go` - Firebase認証ミドルウェア
- `cors.go` - CORS設定(デプロイ前に要チェック)

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
1. **DB接続**
db, err := DBInit()
2. **依存性注入（DI）**
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

3. **ルーティング設定**
```go
mux := http.NewServeMux()
// 公開エンドポイント(関数をそのまま渡す)
mux.HandleFunc(...)
// 認証必須エンドポイント(middleware FirebaseAuthMiddlewareを適用して関数を渡す)
mux.Handle(...)

// CORS設定
wrappedHandler := middleware.CORSMiddleware(mux)

---

##  データフロー例

### 商品検索のフロー

```
1. Client → GET /items?name=keyword

2. Controller (item_query_controller.go)
   ├─ クエリパラメータ取得: keyword
   └─ Usecase呼び出し: Search(keyword)

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
   ├─ バリデーション(userIdがあるか)
   └─ DAO呼び出し: ToggleLike(userID, itemID)

5. DAO (like_dao.go)
   ├─ 既存レコード確認
   ├─ 存在する → DELETE
   └─ 存在しない → INSERT
```

---

## セキュリティ対策

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

## データベーススキーマ

### テーブル一覧

**items（商品）**
```sql
       Table: chat_rooms
Create Table: CREATE TABLE `chat_rooms` (
  `id` varchar(26) NOT NULL,
  `item_id` varchar(26) NOT NULL,
  `buyer_id` varchar(255) NOT NULL,
  `seller_id` varchar(255) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_chat` (`item_id`,`buyer_id`,`seller_id`),
  CONSTRAINT `chat_rooms_ibfk_1` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

       Table: item_images
Create Table: CREATE TABLE `item_images` (
  `id` int NOT NULL AUTO_INCREMENT,
  `item_id` varchar(255) NOT NULL,
  `image_url` text NOT NULL,
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `item_id` (`item_id`),
  CONSTRAINT `item_images_ibfk_1` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

       Table: items
Create Table: CREATE TABLE `items` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `name` varchar(100) NOT NULL,
  `description` text,
  `price` int NOT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'ON_SALE',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `buyer_id` varchar(255) DEFAULT NULL,
  `purchased_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  CONSTRAINT `items_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

       Table: likes
Create Table: CREATE TABLE `likes` (
  `user_id` varchar(255) NOT NULL,
  `item_id` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`user_id`,`item_id`),
  KEY `item_id` (`item_id`),
  CONSTRAINT `likes_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `likes_ibfk_2` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

       Table: messages
Create Table: CREATE TABLE `messages` (
  `id` varchar(26) NOT NULL,
  `chat_room_id` varchar(26) NOT NULL,
  `sender_id` varchar(255) NOT NULL,
  `content` text NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `chat_room_id` (`chat_room_id`),
  CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`chat_room_id`) REFERENCES `chat_rooms` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

       Table: users
Create Table: CREATE TABLE `users` (
  `id` varchar(255) NOT NULL,
  `name` varchar(50) NOT NULL,
  `age` int DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `bio` text,
  `icon_url` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
```

## コーディング規約

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
