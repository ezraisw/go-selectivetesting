package main

import (
	"fmt"
	"os"

	"github.com/ezraisw/go-selectivetesting/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
		return
	}
}
