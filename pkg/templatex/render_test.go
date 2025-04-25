package templatex

import "testing"

func TestRenderToString(t *testing.T) {
	tmpl := "hello {{.Name}}"
	data := map[string]interface{}{
		"Name": "world",
	}
	got, err := RenderToString(tmpl, data)
	if err != nil {
		t.Fatalf("RenderToString() error = %v", err)
		return
	}
	want := "hello world"
	t.Logf("got: %v, want: %v", got, want)
}
