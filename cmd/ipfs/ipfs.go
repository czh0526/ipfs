package main

import (
	cmds "github.com/czh0526/ipfs/commands"
	commands "github.com/czh0526/ipfs/core/commands"
)

var Root = &cmds.Command{
	Options:  commands.Root.Options,
	Helptext: commands.Root.Helptext,
}

var commandsClientCmd = commands.CommandsCmd(Root)

var localCommands = map[string]*cmds.Command{
	"daemon":   daemonCmd,
	"init":     initCmd,
	"commands": commandsClientCmd,
}

var localMap = make(map[*cmds.Command]bool)

func init() {
	Root.Subcommands = localCommands

	for k, v := range commands.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}

	for _, v := range localCommands {
		localMap[v] = true
	}
}

type cmdDetails struct {
	cannotRunOnClient bool
	cannotRunOnDaemon bool
	doesNotUseRepo    bool

	doesNotUseConfigAsInput bool
	preemptsAutoUpdate      bool
}

func (d *cmdDetails) usesConfigAsInput() bool        { return !d.doesNotUseConfigAsInput }
func (d *cmdDetails) doesNotPreemptAutoUpdate() bool { return !d.preemptsAutoUpdate }
func (d *cmdDetails) canRunOnClient() bool           { return !d.cannotRunOnClient }
func (d *cmdDetails) canRunOnDaemon() bool           { return !d.cannotRunOnDaemon }
func (d *cmdDetails) usesRepo() bool                 { return !d.doesNotUseRepo }

var cmdDetailsMap = map[*cmds.Command]cmdDetails{
	initCmd: {
		doesNotUseConfigAsInput: true,
		cannotRunOnDaemon:       true,
		doesNotUseRepo:          true,
	},
	daemonCmd: {
		doesNotUseConfigAsInput: true,
		cannotRunOnDaemon:       true,
	},
	commandsClientCmd: {
		doesNotUseRepo: true,
	},
}
