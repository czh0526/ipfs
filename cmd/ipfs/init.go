package main

import (
	"errors"
	"fmt"

	cmds "github.com/czh0526/ipfs/commands"
)

const (
	nBitsForKeypairDefault = 2048
)

var initCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Initializes ipfs config file.",
		ShortDescription: `
Initializes ipfs configuration files and generates a new keypair.

ipfs uses a repository in the local file system. By default, the repo is
located at ~/.ipfs. To change the repo location, set the $IPFS_PATH
environment variable:

    export IPFS_PATH=/path/to/ipfsrepo
`,
	},
	Arguments: []cmds.Argument{
		cmds.FileArg("default-config", false, false, "Initialize with the given configuration.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.IntOption("bits", "b", "Number of bits to use in the generated RSA private key.").Default(nBitsForKeypairDefault),
		cmds.BoolOption("empty-repo", "e", "Don't add and pin help files to the local storage.").Default(false),

		// TODO need to decide whether to expose the override as a file or a
		// directory. That is: should we allow the user to also specify the
		// name of the file?
		// TODO cmds.StringOption("event-logs", "l", "Location for machine-readable event logs."),
	},
	PreRun: func(req cmds.Request) error {
		fmt.Println("initCmd.PreRun()")
		return errors.New("PreRun() has not be implemented.")
	},
	Run: func(req cmds.Request, res cmds.Response) {
		fmt.Println("initCmd.Run()")
	},
}
