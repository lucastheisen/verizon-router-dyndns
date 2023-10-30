package main

import (
	"os"

	"github.com/lucastheisen/verizon-router-dyndns/cmd/vrd/root"
)

func main() {
	if err := root.New().Execute(); err != nil {
		os.Exit(1)
	}
}
