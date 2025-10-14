package getshorturlhandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Validation(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "Valid url",
			str:  "https://www.ya.ru",
			want: true,
		},
		{
			name: "When got empty string",
			str:  "",
			want: false,
		},
		{
			name: "When got string without website address",
			str:  "https://",
			want: false,
		},
		{
			name: "When got string without scheme",
			str:  "://ya.ru",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation(tt.str)

			assert.Equal(t, tt.want, got)
		})
	}
}
