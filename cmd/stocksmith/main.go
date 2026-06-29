package main

import (
	"github.com/craftybase/stocksmith-cli/commands"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	commands.SetVersion(version, commit, buildDate)
	commands.Execute(version)
}
