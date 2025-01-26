package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func main() {
	// Initialize the logger
	// These are the defaults, final logging parameters are set inside of the cmd package, depending on command-line parameters set.
	log.SetHandler(text.Default)
	log.SetLevel(log.DebugLevel)

	// Jump over to the CLI part
	//_ = cmd.Execute()
}
