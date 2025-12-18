package dao

import (
	"context"
	"db/model"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestItemDao_ItemInsert(t *testing.T) {
	// mock DB の作成
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dao := NewItemDao(db)

	ctx := context.Background()
	item := &model.Item{
		ItemId:      "item1",
		UserId:      "user1",
		Name:        "Test Item",
		Description: "Description",
		Price:       1000,
		Embedding:   []float32{0.1, 0.2},
		ImageURLs:   []string{"http://example.com/1.jpg"},
	}

	// 成功ケース
	t.Run("成功: 商品登録", func(t *testing.T) {
		mock.ExpectBegin()

		// items テーブルへの INSERT
		// Embedding は JSON 文字列になるため、実データに合わせて修正が必要
		// ここでは簡略化して引数のマッチングは甘く
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO items")).
			WithArgs(
				item.ItemId,
				item.UserId,
				item.Name,
				item.Description,
				item.Price,
				sqlmock.AnyArg(), // embedding JSON
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// item_images テーブルへの INSERT
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO item_images")).
			WithArgs(
				item.ItemId,
				item.ImageURLs[0],
				sqlmock.AnyArg(), // created_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := dao.ItemInsert(ctx, item)
		if err != nil {
			t.Errorf("ItemInsert() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// 失敗ケース: items INSERT 失敗
	t.Run("失敗: items INSERT エラー", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO items")).
			WillReturnError(sqlmock.ErrCancelled) // 適当なエラー

		mock.ExpectRollback()

		err := dao.ItemInsert(ctx, item)
		if err == nil {
			t.Errorf("ItemInsert() expected error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestItemDao_PurchaseItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dao := NewItemDao(db)
	ctx := context.Background()

	itemID := "item1"
	buyerID := "buyer1"

	// 成功ケース
	t.Run("成功: 購入処理", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta("UPDATE items SET status = ?, buyer_id = ?, purchased_at = ? WHERE id = ? AND status = ?")).
			WithArgs(
				model.StatusSold,
				buyerID,
				sqlmock.AnyArg(), // purchased_at
				itemID,
				model.StatusOnSale,
			).
			WillReturnResult(sqlmock.NewResult(1, 1)) // 1行更新

		mock.ExpectCommit()

		err := dao.PurchaseItem(ctx, itemID, buyerID)
		if err != nil {
			t.Errorf("PurchaseItem() error = %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	// 失敗ケース: 更新対象なし (既に売れている場合など)
	t.Run("失敗: 更新対象なし", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta("UPDATE items SET status = ?, buyer_id = ?, purchased_at = ? WHERE id = ? AND status = ?")).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0行更新

		mock.ExpectRollback()

		err := dao.PurchaseItem(ctx, itemID, buyerID)
		if err == nil {
			t.Errorf("PurchaseItem() expected error, got nil")
		} else if err.Error() != "item not found or already sold" {
			t.Errorf("PurchaseItem() unexpected error: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
