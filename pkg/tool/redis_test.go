package tool

import "testing"

func TestParseRedisURI(t *testing.T) {
	uri := "redis://localhost:6379"
	addr, password, database, err := ParseRedisURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(addr, password, database)
}

func TestRedisPing(t *testing.T) {
	uri := "redis://localhost:6379"
	addr, password, database, err := ParseRedisURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	err = RedisPing(addr, password, database)
	if err != nil {
		t.Fatal(err)
	}
}
