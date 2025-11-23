package model

import "time"

// MaxNameLen definition of user rule
const MaxNameLen = 50

// User struct and method used for operation related to user
type User struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCreateRequest struct and method used to create user
type UserCreateRequest struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

// IsValid バリデーション
func (req *UserCreateRequest) IsValid() bool {
	return req.Name != "" && len(req.Name) <= MaxNameLen
}
