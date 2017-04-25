package commands

import (
	"fmt"
	"io"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/czh0526/ipfs/path"
)

var log = logging.Logger("command")

// Function comment to be added
type Function func(Request, Response)

type Marshaler func(Response) (io.Reader, error)
type MarshalerMap map[EncodingType]Marshaler

// Command 是一个可运行的command，带着arguments和options(flags)
// 它也可以有子命令
type Command struct {
	Options   []Option
	Arguments []Argument
	Helptext  HelpText

	PreRun     func(req Request) error
	Run        Function
	PostRun    Function
	Marshalers map[EncodingType]Marshaler

	External    bool
	Type        interface{}
	Subcommands map[string]*Command
}

var ErrNotCallable = ClientError("This command can't be called directly. Try")

func (c *Command) Call(req Request) Response {
	res := NewResponse(req)

	cmds, err := c.Resolve(req.Path())
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}
	cmd := cmds[len(cmds)-1]

	if cmd.Run == nil {
		res.SetError(ErrNotCallable, ErrClient)
		return res
	}

	err = cmd.CheckArguments(req)
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}

	err = req.ConvertOptions()
	if err != nil {
		res.SetError(err, ErrClient)
		return res
	}

	cmd.Run(req, res)
	if res.Error() != nil {
		return res
	}

	return res
}

func (c *Command) Subcommand(id string) *Command {
	return c.Subcommands[id]
}

type CommandVisitor func(*Command)

func (c *Command) Walk(visitor CommandVisitor) {
	visitor(c)
	for _, cm := range c.Subcommands {
		cm.Walk(visitor)
	}
}

func (c *Command) ProcessHelp() {
	c.Walk(func(cm *Command) {
		ht := &cm.Helptext
		if len(ht.LongDescription) == 0 {
			ht.LongDescription = ht.ShortDescription
		}
	})
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

func (c *Command) CheckArguments(req Request) error {
	args := req.(*request).arguments
	numRequired := 0
	for _, argDef := range c.Arguments {
		if argDef.Required {
			numRequired++
		}
	}

	valueIndex := 0
	for i, argDef := range c.Arguments {
		if len(args)-valueIndex <= numRequired && !argDef.Required ||
			argDef.Type == ArgFile {
			continue
		}

		// the value for this argument definition. can be nil if it
		// wasn't provided by the caller
		v, found := "", false
		if valueIndex < len(args) {
			v = args[valueIndex]
			found = true
			valueIndex++
		}

		// in the case of a non-variadic required argument that supports stdin
		if !found && len(c.Arguments)-1 == i && argDef.SupportsStdin {
			found = true
		}

		err := checkArgValue(v, found, argDef)
		if err != nil {
			return err
		}

		// any additional values are for the variadic arg definition
		if argDef.Variadic && valueIndex < len(args)-1 {
			for _, val := range args[valueIndex:] {
				err := checkArgValue(val, true, argDef)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Command) Get(path []string) (*Command, error) {
	cmds, err := c.Resolve(path)
	if err != nil {
		return nil, err
	}
	return cmds[len(cmds)-1], nil
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

func ClientError(msg string) error {
	return &Error{Code: ErrClient, Message: msg}
}

func checkArgValue(v string, found bool, def Argument) error {
	if def.Variadic && def.SupportsStdin {
		return nil
	}

	if !found && def.Required {
		return fmt.Errorf("Argument '%s' is required", def.Name)
	}

	return nil
}
