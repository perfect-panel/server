package handler

import "testing"

func TestNormalizePluginDispatchPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "/"},
		{name: "wildcard", in: "*", want: "/"},
		{name: "with slash", in: "/webhook", want: "/webhook"},
		{name: "without slash", in: "webhook", want: "/webhook"},
		{name: "nested without slash", in: "api/webhook", want: "/api/webhook"},
		{name: "trim trailing slash", in: "/webhook/", want: "/webhook"},
		{name: "trim spaces", in: " webhook ", want: "/webhook"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePluginDispatchPath(tt.in); got != tt.want {
				t.Fatalf("normalizePluginDispatchPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
