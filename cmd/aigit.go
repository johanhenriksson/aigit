package main

import (
	"github.com/johanhenriksson/aigit"
)

func main() {
	model := aigit.GetDefaultModel()

	cli := aigit.NewCli(model)
	cli.Run()
}
