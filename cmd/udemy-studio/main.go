package main

import (
	"os"

	"github.com/vukiman1/udemy-studio/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
