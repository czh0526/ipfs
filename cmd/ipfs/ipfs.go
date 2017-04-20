package main

import (
	cmds "github.com/czh0526/ipfs/commands"
	commands "github.com/czh0526/ipfs/core/commands"
)

var Root = &cmds.Command{
	Options:  commands.Root.Options,
	Helptext: commands.Root.Helptext,
}
