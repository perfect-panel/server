package jsonx

import "testing"

type User struct {
	Id   int64
	Name string
	Age  int64
}

func TestJson(t *testing.T) {
	t.Log("TestJson")
	user := &User{
		Id:   1,
		Name: "test",
		Age:  18,
	}
	b, err := Marshal(user)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(b))
}
