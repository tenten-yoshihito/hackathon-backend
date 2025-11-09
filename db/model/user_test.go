package model

import "testing"

func TestUserCreateRequest_IsValid(t *testing.T) {
	testCases := []struct {
		name string
		req  UserCreateRequest
		want bool // IsValid() の戻り値は bool
	}{
		{
			name: "成功: 正常な値",
			req:  UserCreateRequest{Name: "Taro", Age: 25},
			want: true,
		},
		{
			name: "失敗: 名前が空",
			req:  UserCreateRequest{Name: "", Age: 25},
			want: false,
		},
		{
			name: "失敗: 名前が長すぎる (MaxNameLen=50)",
			req:  UserCreateRequest{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Age: 25}, // 51文字
			want: false,
		},
		{
			name: "失敗: 年齢が若すぎる (MinAge=20)",
			req:  UserCreateRequest{Name: "Jiro", Age: 20}, // IsValid は Age > MinAge (20)
			want: false,
		},
		{
			name: "失敗: 年齢が高すぎる (MaxAge=80)",
			req:  UserCreateRequest{Name: "Saburo", Age: 80}, // IsValid は Age < MaxAge (80)
			want: false,
		},
		{
			name: "成功: 正常な値",
			req:  UserCreateRequest{Name: "yuu", Age: 25},
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
