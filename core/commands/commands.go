package commands

import (
	"errors"
	"fmt"
	"io"

	cmds "github.com/czh0526/ipfs/commands"
)

type Option struct {
	Names []string
}

type Command struct {
	Name        string
	Subcommands []Command
	Options     []Option
}

const (
	flagsOptionName = "flags"
)

func CommandsCmd(root *cmds.Command) *cmds.Command {
	return &cmds.Command{
		Helptext: cmds.HelpText{
			Tagline:          "List all available commands.",
			ShortDescription: "Lists all available commands (and subcommands) and exists.",
		},
		Options: []cmds.Option{
			cmds.BoolOption(flagsOptionName, "f", "Show command flags").Default(false),
		},
		Run: func(req cmds.Request, res cmds.Response) {
			fmt.Println("core/commands/Run() has not implemented.")
		},
		Marshalers: cmds.MarshalerMap{
			cmds.Text: func(res cmds.Response) (io.Reader, error) {
				fmt.Println("core/commands/Text()")
				return nil, errors.New("core/commands/Text() has not implements !")
			},
		},
		Type: Command{},
	}
}
