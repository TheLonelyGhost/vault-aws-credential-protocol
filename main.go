package main

import (
	"context"
	"log"

	"github.com/thelonelyghost/vault-aws-credential-protocol/cmd"
)

func main() {
	ctx := context.Background()
	if err := cmd.Execute(ctx); err != nil {
		log.Fatal(err)
	}
}
