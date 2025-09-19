package main

import (
	"os"

	"github.com/svanhalla/prompt-lab/greetd/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
