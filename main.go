package main

import (
	"os"

	"github.com/hserkanyilmaz/costdiff/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
