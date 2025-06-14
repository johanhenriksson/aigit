package main

import (
	"github.com/johanhenriksson/aigit"
)

func main() {
	if err := aigit.VerifyGit(); err != nil {
		panic(err)
	}

	print("diff:\n")
	aigit.GetStatus()

	model := aigit.GetDefaultModel()

	cli := aigit.NewCli(model)
	cli.Run()
}
