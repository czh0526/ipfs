package cli

import (
	"fmt"
	"os"
	"strings"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	cmds "github.com/czh0526/ipfs/commands"
)

var log = logging.Logger("commands/cli")

func Parse(input []string, stdin *os.File, root *cmds.Command) (cmds.Request, *cmds.Command, []string, error) {
	path, opts, _, cmd, err := parseOpts(input, root)
	if err != nil {
		return nil, nil, path, err
	}

	optDefs, err := root.GetOptions(path)
	if err != nil {
		return nil, cmd, path, err
	}

	req, err := cmds.NewRequest(path, opts, nil, nil, cmd, optDefs)
	if err != nil {
		return nil, cmd, path, err
	}

	return req, cmd, path, nil
}

func parseOpts(args []string, root *cmds.Command) (
	path []string,
	opts map[string]interface{},
	stringVals []string,
	cmd *cmds.Command,
	err error,
) {
	path = make([]string, 0, len(args))
	stringVals = make([]string, 0, len(args))
	optDefs := map[string]cmds.Option{}
	opts = map[string]interface{}{}
	cmd = root

	parseFlag := func(name string, arg *string, mustUse bool) (bool, error) {
		if _, ok := opts[name]; ok {
			return false, fmt.Errorf("Duplicate values for option '%s'", name)
		}

		optDef, found := optDefs[name]
		if !found {
			err = fmt.Errorf("Unrecognized option '%s'", name)
			return false, err
		}

		if optDef.Type() == cmds.Bool {
			if arg == nil || !mustUse {
				opts[name] = true
				return false, nil
			}
			argVal := strings.ToLower(*arg)
			switch argVal {
			case "true":
				opts[name] = true
				return true, nil
			case "false":
				opts[name] = false
				return true, nil
			default:
				return true, fmt.Errorf("Option '%s' takes true/false arguments, but was passed '%s'", name, argVal)
			}
		} else {
			if arg == nil {
				return true, fmt.Errorf("Missing argument for option '%s'", name)
			}
			opts[name] = *arg
			return true, nil
		}
	}

	optDefs, err = root.GetOptions(path)
	if err != nil {
		return
	}
	log.Debugf("optDefs = %v.", optDefs)

	consumed := false
	for i, arg := range args {
		switch {
		case consumed:
			consumed = false
			continue

		case arg == "--":
			stringVals = append(stringVals, args[i+1:]...)
			return

		case strings.HasPrefix(arg, "--"):
			var slurped bool
			var next *string
			split := strings.SplitN(arg, "=", 2)
			if len(split) == 2 {
				slurped = false
				arg = split[0]
				next = &split[1]
			} else {
				slurped = true
				if i+1 < len(args) {
					next = &args[i+1]
				} else {
					next = nil
				}
			}
			consumed, err = parseFlag(arg[2:], next, len(split) == 2)
			if err != nil {
				return
			}
			if !slurped {
				consumed = false
			}

		case strings.HasPrefix(arg, "-") && arg != "-":
			for arg = arg[1:]; len(arg) != 0; arg = arg[1:] {
				var rest *string
				var slurped bool
				mustUse := false
				if len(arg) > 1 {
					slurped = false
					str := arg[1:]
					if len(str) > 0 && str[0] == '=' {
						str = str[1:]
						mustUse = true
					}
					rest = &str
				} else {
					slurped = true
					if i+1 < len(args) {
						rest = &args[i+1]
					} else {
						rest = nil
					}
				}

				var end bool
				end, err = parseFlag(arg[:1], rest, mustUse)
				if err != nil {
					return
				}
				if end {
					consumed = slurped
					break
				}
			}
		default:
			sub := cmd.Subcommand(arg)
			if sub != nil {
				cmd = sub
				path = append(path, arg)
				optDefs, err = root.GetOptions(path)
				if err != nil {
					return
				}

				if cmd.External {
					stringVals = append(stringVals, args[i+1:]...)
					return
				}

			} else {
				stringVals = append(stringVals, arg)
				if len(path) == 0 {
					err = printSuggestions(stringVals, root)
					return
				}
			}
		}
	}
	return
}
