package urlvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Validate(t *testing.T) {
	validator := New()
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
			got := validator.Validate(tt.str)

			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	validator := New()
	inputs := []string{
		"https://www.ya.ru",
		"://",
		"",
		"testing",
	}

	b.Run("validateByURL", func(b *testing.B) {
		for b.Loop() {
			for in := range inputs {
				validator.validateByURL(inputs[in])
			}
		}
	})

	b.Run("validateByRegexp", func(b *testing.B) {
		for b.Loop() {
			for in := range inputs {
				validator.validateByRegexp(inputs[in])
			}
		}
	})

	b.Run("validateByStringsAndURL", func(b *testing.B) {
		for b.Loop() {
			for in := range inputs {
				validator.validateByStringsAndURL(inputs[in])
			}
		}
	})
}
