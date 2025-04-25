package snowflake

import (
	"testing"
)

func TestGetLocalIp(t *testing.T) {
	tests := []struct {
		name string
		// want    byte
		wantErr  bool
		wantFunc func(byte) (bool, string)
	}{
		// TODO: Add test cases.
		{
			name:     "",
			wantErr:  false,
			wantFunc: func(got byte) (bool, string) { return got > 0, "got > 0" },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIp()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalIp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if r, s := tt.wantFunc(got); !r {
				t.Errorf("GetLocalIp() = %v, want %v", got, s)
			}
		})
	}
}

func TestGetID(t *testing.T) {
	tests := []struct {
		name string
		// want    int64
		wantErr  bool
		wantFunc func(int64) (bool, string)
	}{
		// TODO: Add test cases.
		{
			name:    "",
			wantErr: false,
			wantFunc: func(got int64) (bool, string) {
				return got > 0, "got > 0"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetID()
			if r, s := tt.wantFunc(got); !r {
				t.Errorf("GetID() = %v, want %v", got, s)
			}
		})
	}
}
