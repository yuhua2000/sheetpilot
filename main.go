package main

import (
	"os"

	"github.com/yuhua2000/sheetpilot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
