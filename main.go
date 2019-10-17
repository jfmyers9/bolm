package main

import (
	"fmt"
	"os"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(exitstatus.FromError(err))
	}

	os.Exit(0)
}
