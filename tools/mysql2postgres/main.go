package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/perfect-panel/server/internal/migration/mysql2postgres"
)

func main() {
	if err := mysql2postgres.Run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatalf("mysql2postgres failed: %v", err)
	}
}
