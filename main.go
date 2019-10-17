package main

import (
	"bolm/commands"
	"fmt"
	"os"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
