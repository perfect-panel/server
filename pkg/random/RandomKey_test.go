package random

import (
	"math/rand"
	"testing"
	"time"

	"github.com/perfect-panel/server/pkg/snowflake"

	"github.com/stretchr/testify/assert"
)

func TestEncodeBase62(t *testing.T) {
	start := 1112275807
	length := 1558080
	n := length + start
	// n := 328564998144
	m := make(map[string]struct{})
	// m := make(map[string]struct{}, length)
	var inviteCode string
	for i := start; i < n; i++ {
		// inviteCode = EncodeBase36(int64(i))
		inviteCode = EncodeBase36(snowflake.GetID())
		if v, ok := m[inviteCode]; ok {
			t.Fatal(v, inviteCode)
		}
		m[inviteCode] = struct{}{}
	}
	t.Log(inviteCode)

	assert.Equal(t, length, len(m))
}

func TestInt64ToDashedString(t *testing.T) {
	type args struct {
		strNum string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				strNum: "123",
			},
			want: "123",
		},
		{
			name: "",
			args: args{
				strNum: "1234",
			},
			want: "1234",
		},
		{
			name: "",
			args: args{
				strNum: "12345",
			},
			want: "1234-5",
		},
		{
			name: "",
			args: args{
				strNum: "12345678",
			},
			want: "1234-5678",
		},
		{
			name: "",
			args: args{
				strNum: "123456789",
			},
			want: "1234-5678-9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, StrToDashedString(tt.args.strNum), "StrToDashedString(%v)", tt.args.strNum)
		})
	}
}

// ShuffleString shuffles the characters in a string.
func ShuffleString(s string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	runes := []rune(s)
	for i := range runes {
		j := r.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
