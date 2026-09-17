//go:build !system

package system

import "testing"

func TestSystemTestTopicCount_usesOneByDefault(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "未設定", raw: "", want: 1},
		{name: "空白", raw: "  ", want: 1},
		{name: "不正な文字列", raw: "invalid", want: 1},
		{name: "ゼロ", raw: "0", want: 1},
		{name: "負数", raw: "-1", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			t.Setenv(systemTestTopicCountEnv, tt.raw)

			// When
			got := systemTestTopicCount()

			// Then
			if got != tt.want {
				t.Fatalf("systemTestTopicCount() = %d、期待値 = %d", got, tt.want)
			}
		})
	}
}
