package main

import (
	"fmt"

	"github.com/charter-se/barrelman/cmd"
)

var (
	version = "No version declared at build"
	commit  = "No git commit declared at build"
)

func main() {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)

	cmd.Execute()
}
