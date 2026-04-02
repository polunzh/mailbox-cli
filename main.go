package main

import (
	"os"

	"github.com/polunzh/mailbox-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
