package dao

import (
	"context"
	"database/sql"
	"db/model"
	"errors"
	"fmt"
	"log"
	"time"
)

type UserDAO interface {
	List(ctx context.Context) ([]model.User, error)
	DBInsert(ctx context.Context, user *model.User) error
}

type userDao struct {
	DB *sql.DB
}

func NewUserDao(db *sql.DB) UserDAO {
	return &userDao{DB: db}
}

func (dao *userDao) List(ctx context.Context) ([]model.User, error) {

	query := "SELECT id, name, icon_url FROM users"
	rows, err := dao.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fail:dao.DB.Query:%w", err)
	}

	defer func() {
		if CloseErr := rows.Close(); CloseErr != nil {
			log.Printf("fail:rows.Close,%v\n", CloseErr)
		}
	}()

	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.Id, &u.Name, &u.IconURL); err != nil {
			return nil, fmt.Errorf("fail:ows.Scan:%w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fial:rows.Err:%w", err)
	}

	return users, nil
}

// DBInsert 指定されたuerをinsertする
func (dao *userDao) DBInsert(ctx context.Context, user *model.User) error {

	tx, err := dao.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail:txBegin(): %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("fail:tx.Rollback,%v\n", err)
		}
	}()

	now := time.Now()
	query := `INSERT INTO users 
              (id, name, age, email, bio, icon_url, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query, user.Id, user.Name, user.Age, user.Email, user.Bio, user.IconURL, now, now)
	if err != nil {
		return fmt.Errorf("fail:db.Exec: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fail:tx.Commit: %w", err)
	}

	return nil
}
