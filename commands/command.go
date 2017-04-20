package commands

import (
	"fmt"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/czh0526/ipfs/path"
)

var log = logging.Logger("command")

// Function comment to be added
type Function func(Request, Response)

// Command 是一个可运行的command，带着arguments和options(flags)
// 它也可以有子命令
type Command struct {
	Options   []Option
	Arguments []Argument
	Helptext  HelpText

	PreRun  func(req Request) error
	Run     Function
	PostRun Function

	External    bool
	Type        interface{}
	Subcommands map[string]*Command
}

func (c *Command) Subcommand(id string) *Command {
	return c.Subcommands[id]
}

// Resolve path ==> Commands
func (c *Command) Resolve(pth []string) ([]*Command, error) {
	cmds := make([]*Command, len(pth)+1)
	cmds[0] = c

	cmd := c
	for i, name := range pth {
		cmd = cmd.Subcommand(name)
		if cmd == nil {
			pathS := path.Join(pth[:i])
			return nil, fmt.Errorf("Undefined command: '%s'", pathS)
		}

		cmds[i+1] = cmd
	}

	return cmds, nil
}

// GetOptions  path ==> Options
func (c *Command) GetOptions(path []string) (map[string]Option, error) {
	options := make([]Option, 0, len(c.Options))

	// path ==> command
	cmds, err := c.Resolve(path)
	if err != nil {
		return nil, err
	}
	log.Debugf("cmds = %v", cmds)
	cmds = append(cmds, globalCommand)

	for _, cmd := range cmds {
		options = append(options, cmd.Options...)
	}

	optionsMap := make(map[string]Option)
	for _, opt := range options {
		for _, name := range opt.Names() {
			if _, found := optionsMap[name]; found {
				return nil, fmt.Errorf("Option name '%s' used multiple times", name)
			}
			optionsMap[name] = opt
		}
	}

	return optionsMap, nil
}

// HelpText comment to be added
type HelpText struct {
	Tagline               string
	ShortDescription      string
	SynopsisOptionsValues map[string]string
	Usage                 string
	LongDescription       string
	Options               string
	Arguments             string
	Subcommands           string
	Synopsis              string
}
