package main

import (
	"fmt"
	"os"

	"bb-lite/cmd"
	"bb-lite/internal/config"
)

func main() {
	config.Init()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
