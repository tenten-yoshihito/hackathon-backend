package model

import "testing"

// TestUserCreateRequest_IsValid : ユーザー作成リクエストのバリデーションtest
func TestUserCreateRequest_IsValid(t *testing.T) {
	testCases := []struct {
		name string
		req  UserCreateRequest
		want bool // IsValid() の戻り値は bool
	}{
		{
			name: "成功: 正常な値",
			req:  UserCreateRequest{Name: "Taro", Age: 25, Email: "taro@example.com"},
			want: true,
		},
		{
			name: "失敗: 名前が空",
			req:  UserCreateRequest{Name: "", Age: 25, Email: "test@example.com"},
			want: false,
		},
		{
			name: "失敗: 名前が長すぎる (MaxNameLen=50)",
			req:  UserCreateRequest{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Age: 25, Email: "test@example.com"}, // 51文字
			want: false,
		},
		{
			name: "成功: 年齢制限撤廃 (20歳)",
			req:  UserCreateRequest{Name: "Jiro", Age: 20, Email: "jiro@example.com"},
			want: true,
		},
		{
			name: "成功: 年齢制限撤廃 (80歳)",
			req:  UserCreateRequest{Name: "Saburo", Age: 80, Email: "saburo@example.com"},
			want: true,
		},
		{
			name: "成功: 正常な値 (EmailなしでもOK)",
			req:  UserCreateRequest{Name: "yuu", Age: 25, Email: ""},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.req.IsValid()
			if got != tc.want {
				t.Errorf("IsValid() = %v, want %v", got, tc.want)
			}
		})
	}
}
