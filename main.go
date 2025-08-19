package main

import (
	"context"

	"github.com/thelonelyghost/vault-aws-credential-protocol/cmd"
)

func main() {
	ctx := context.Background()
	cmd.Execute(ctx)
}
