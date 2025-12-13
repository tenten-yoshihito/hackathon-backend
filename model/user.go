package model

import "time"

// MaxNameLen definition of user rule
const MaxNameLen = 50
const MaxBioLen = 500

// User struct and method used for operation related to user
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

// UserCreateRequest struct and method used to create user
type UserCreateRequest struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	IconURL string `json:"icon_url"`
}

// IsValid バリデーション
func (req *UserCreateRequest) IsValid() bool {
	return req.Name != "" && len(req.Name) <= MaxNameLen
}

// UserUpdateRequest struct for updating user profile
type UserUpdateRequest struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Bio     string `json:"bio"`
	IconURL string `json:"icon_url"`
}

// IsValid バリデーション
func (req *UserUpdateRequest) IsValid() bool {
	if req.Name == "" || len(req.Name) > MaxNameLen {
		return false
	}
	// 年齢は-1（未設定）、0（非公開）、または1-150の範囲
	if req.Age < -1 || req.Age > 150 {
		return false
	}
	if len(req.Bio) > MaxBioLen {
		return false
	}
	return true
}
