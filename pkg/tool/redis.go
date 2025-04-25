package tool

import (
	"context"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
)

func ParseRedisURI(uri string) (addr, password string, database int, err error) {
	parsedURI, err := url.Parse(uri)

	if err != nil {
		return "", "", 0, err
	}
	host := parsedURI.Hostname()
	port := parsedURI.Port()
	if port == "" {
		port = "6379"
	}
	addr = fmt.Sprintf("%s:%s", host, port)

	// password
	if parsedURI.User != nil {
		password, _ = parsedURI.User.Password()
	}
	if len(parsedURI.Path) > 1 { // Path: "/0"
		var dbIndex int
		_, err = fmt.Sscanf(parsedURI.Path, "/%d", &dbIndex)
		if err == nil {
			database = dbIndex
		}
	}
	return
}

func RedisPing(addr, password string, database int) error {
	rds := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       database,
	})
	return rds.Ping(context.Background()).Err()
}
