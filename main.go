package main

import (
	"os"

	"github.com/ygncode/gdk/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
