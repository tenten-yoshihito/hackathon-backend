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
	GetUser(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
}

type userDao struct {
	DB *sql.DB
}

// NewUserDao : UserDAOの生成
func NewUserDao(db *sql.DB) UserDAO {
	return &userDao{DB: db}
}

// List : ユーザー一覧を取得する
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

	now := time.Now()
	query := `INSERT INTO users 
              (id, name, age, email, bio, icon_url, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := dao.DB.ExecContext(ctx, query, user.Id, user.Name, user.Age, user.Email, user.Bio, user.IconURL, now, now)
	if err != nil {
		return fmt.Errorf("fail:db.Exec: %w", err)
	}

	return nil
}

// GetUser : 指定されたIDのユーザーを取得
func (dao *userDao) GetUser(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, name, age, email, bio, icon_url, created_at, updated_at 
              FROM users WHERE id = ?`

	var user model.User
	err := dao.DB.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.Name,
		&user.Age,
		&user.Email,
		&user.Bio,
		&user.IconURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("fail:dao.DB.QueryRow:%w", err)
	}

	return &user, nil
}

// UpdateUser : ユーザー情報を更新
func (dao *userDao) UpdateUser(ctx context.Context, user *model.User) error {
	now := time.Now()
	query := `UPDATE users 
              SET name = ?, age = ?, bio = ?, icon_url = ?, updated_at = ? 
              WHERE id = ?`

	result, err := dao.DB.ExecContext(ctx, query, user.Name, user.Age, user.Bio, user.IconURL, now, user.Id)
	if err != nil {
		return fmt.Errorf("fail:db.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fail:result.RowsAffected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
