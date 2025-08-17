package cli

import (
	"log"
	"os"

	"github.com/tdewolff/argp"
)

var Input string

func Parser() {
	Input = "."
	argsCmd := argp.NewCmd(&Commands{}, "droplink")
	argsCmd.AddArg(&Input, "input", "Input file or dirctory to upload")
	if len(os.Args) == 1 {
		argsCmd.PrintHelp()
		os.Exit(0)
	}

	argsCmd.Parse()
	if argsCmd.Error != nil {
		log.Fatal("Error: ", argsCmd.Error)
	}
}
