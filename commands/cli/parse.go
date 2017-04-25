package cli

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	osh "gx/ipfs/QmXuBJ7DR6k3rmUEKtvVMhwjmXDuJgXXPUt4LQXKBMsU93/go-os-helper"
	u "gx/ipfs/QmZuY8aV7zbNXVy6DyN9SmnuH3o9nG852F4aTiSBpts8d1/go-ipfs-util"

	cmds "github.com/czh0526/ipfs/commands"
	"github.com/czh0526/ipfs/commands/files"
)

var log = logging.Logger("commands/cli")

func Parse(input []string, stdin *os.File, root *cmds.Command) (cmds.Request, *cmds.Command, []string, error) {
	// 分析命令行中的 Options
	path, opts, stringVals, cmd, err := parseOpts(input, root)
	if err != nil {
		return nil, nil, path, err
	}

	// 得到相关 Options 的标准定义
	optDefs, err := root.GetOptions(path)
	if err != nil {
		return nil, cmd, path, err
	}

	// 构造 Request
	req, err := cmds.NewRequest(path, opts, nil, nil, cmd, optDefs)
	if err != nil {
		return nil, cmd, path, err
	}

	if len(path) == 2 && path[0] == "bootstrap" {
		if (path[1] == "add" && opts["default"] == true) ||
			(path[1] == "rm" && opts["all"] == true) {
			stdin = nil
		}
	}

	// 分析命令行中的 Arguments
	stringArgs, _, err := ParseArgs(req, stringVals, stdin, cmd.Arguments, root)
	if err != nil {
		return req, cmd, path, err
	}
	req.SetArguments(stringArgs)

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

	// 获得全局的属性值
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

const msgStdinInfo = "ipfs: Reading from %s; send Ctrl-d to stop."

func parseArgs(inputs []string, stdin *os.File, argDefs []cmds.Argument, recursive, hidden bool, root *cmds.Command) ([]string, []files.File, error) {
	// if osh.IsWindows() {
	// 	stdin = nil
	// }

	// 计数 Required arg
	numRequired := 0
	for _, argDef := range argDefs {
		if argDef.Required {
			numRequired++
		}
	}

	//
	numInputs := len(inputs)
	if len(argDefs) > 0 && argDefs[len(argDefs)-1].SupportsStdin && stdin != nil {
		numInputs++
	}

	notVariadic := len(argDefs) == 0 || !argDefs[len(argDefs)-1].Variadic
	if notVariadic && len(inputs) > len(argDefs) {
		err := printSuggestions(inputs, root)
		return nil, nil, err
	}

	stringArgs := make([]string, 0, numInputs)
	fileArgs := make(map[string]files.File)
	argDefIndex := 0

	for i := 0; i < numInputs; i++ {
		argDef := getArgDef(argDefIndex, argDefs)
		for numInputs-i <= numRequired && !argDef.Required {
			argDefIndex++
			argDef = getArgDef(argDefIndex, argDefs)
		}
		if argDef.Required {
			numRequired--
		}

		fillingVariadic := argDefIndex+1 > len(argDefs)
		switch argDef.Type {
		case cmds.ArgString:
			if len(inputs) > 0 {
				stringArgs, inputs = append(stringArgs, inputs[0]), inputs[1:]
			} else if stdin != nil && argDef.SupportsStdin && !fillingVariadic {
				if r, err := maybeWrapStdin(stdin, msgStdinInfo); err == nil {
					fileArgs[stdin.Name()] = files.NewReaderFile("stdin", "", r, nil)
					stdin = nil
				}
			}
		case cmds.ArgFile:
			if len(inputs) > 0 {
				fpath := inputs[0]
				inputs = inputs[1:]
				var file files.File
				if fpath == "-" {
					r, err := maybeWrapStdin(stdin, msgStdinInfo)
					if err != nil {
						return nil, nil, err
					}
					fpath = stdin.Name()
					file = files.NewReaderFile("", fpath, r, nil)

				} else {
					nf, err := appendFile(fpath, argDef, recursive, hidden)
					if err != nil {
						return nil, nil, err
					}
					file = nf
				}
				fileArgs[fpath] = file

			} else if stdin != nil && argDef.SupportsStdin &&
				argDef.Required && !fillingVariadic {
				r, err := maybeWrapStdin(stdin, msgStdinInfo)
				if err != nil {
					return nil, nil, err
				}
				fpath := stdin.Name()
				fmt.Printf("fpath = %v ...\n", fpath)
				fileArgs[fpath] = files.NewReaderFile("", fpath, r, nil)
			}
		}
		argDefIndex++
	}

	if len(argDefs) > argDefIndex {
		for _, argDef := range argDefs[argDefIndex:] {
			if argDef.Required {
				return nil, nil, fmt.Errorf("Argument '%s' is required", argDef.Name)
			}
		}
	}

	return stringArgs, filesMapToSortedArr(fileArgs), nil
}

const notRecursiveFmtStr = "'%s' is a directory, use the '-%s' flag to specify directories"
const dirNotSupportedFmtStr = "Invalid path '%s', argument '%s' does not support directories"
const winDriveLetterFmtStr = "%q is a drive letter, not a drive path"

func appendFile(fpath string, argDef *cmds.Argument, recursive, hidden bool) (files.File, error) {
	// resolve Windows relative dot paths like `X:.\somepath`
	if osh.IsWindows() {
		if len(fpath) >= 3 && fpath[1:3] == ":." {
			var err error
			fpath, err = filepath.Abs(fpath)
			if err != nil {
				return nil, err
			}
		}
	}

	if fpath == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cwd, err = filepath.EvalSymlinks(cwd)
		if err != nil {
			return nil, err
		}
		fpath = cwd
	}

	fpath = filepath.Clean(fpath)

	stat, err := os.Lstat(fpath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		if !argDef.Recursive {
			return nil, fmt.Errorf(dirNotSupportedFmtStr, fpath, argDef.Name)
		}
		if !recursive {
			return nil, fmt.Errorf(notRecursiveFmtStr, fpath, cmds.RecShort)
		}
	}

	if osh.IsWindows() {
		return windowsParseFile(fpath, hidden, stat)
	}

	return files.NewSerialFile(path.Base(fpath), fpath, hidden, stat)
}

func filesMapToSortedArr(fs map[string]files.File) []files.File {
	var names []string
	for name, _ := range fs {
		names = append(names, name)
	}

	sort.Strings(names)

	var out []files.File
	for _, f := range names {
		out = append(out, fs[f])
	}
	return out
}

func ParseArgs(req cmds.Request, inputs []string, stdin *os.File, argDefs []cmds.Argument, root *cmds.Command) ([]string, []files.File, error) {
	var err error
	recursiveOpt := req.Option(cmds.RecShort)
	recursive := false
	if recursiveOpt != nil && recursiveOpt.Definition() == cmds.OptionRecursivePath {
		recursive, _, err = recursiveOpt.Bool()
		if err != nil {
			return nil, nil, u.ErrCast()
		}
	}

	hiddenOpt := req.Option("hidden")
	hidden := false
	if hiddenOpt != nil {
		hidden, _, err = hiddenOpt.Bool()
		if err != nil {
			return nil, nil, u.ErrCast()
		}
	}

	return parseArgs(inputs, stdin, argDefs, recursive, hidden, root)
}

func getArgDef(i int, argDefs []cmds.Argument) *cmds.Argument {
	if i < len(argDefs) {
		return &argDefs[i]
	} else if len(argDefs) > 0 {
		return &argDefs[len(argDefs)-1]
	}
	return nil
}

func maybeWrapStdin(f *os.File, msg string) (io.ReadCloser, error) {
	isTty, err := isTty(f)
	if err != nil {
		return nil, err
	}

	if isTty {
		return newMessageReader(f, fmt.Sprintf(msg, f.Name())), nil
	}
	return f, nil
}

func isTty(f *os.File) (bool, error) {
	fInfo, err := f.Stat()
	if err != nil {
		return false, err
	}
	return (fInfo.Mode() & os.ModeCharDevice) != 0, nil
}

type messageReader struct {
	r       io.ReadCloser
	done    bool
	message string
}

func newMessageReader(r io.ReadCloser, msg string) io.ReadCloser {
	return &messageReader{
		r:       r,
		message: msg,
	}
}

func (r *messageReader) Read(b []byte) (int, error) {
	if !r.done {
		r.done = true
	}
	return r.r.Read(b)
}

func (r *messageReader) Close() error {
	return r.r.Close()
}

func windowsParseFile(fpath string, hidden bool, stat os.FileInfo) (files.File, error) {
	// special cases for Windows drive roots i.e. `X:\` and their long form `\\?\X:\`
	// drive path must be preserved as `X:\` (or it's longform) and not converted to `X:`, `X:.`, `\`, or `/` here
	switch len(fpath) {
	case 3:
		// `X:` is cleaned to `X:.` which may not be the expected behaviour by the user, they'll need to provide more specific input
		if fpath[1:3] == ":." {
			return nil, fmt.Errorf(winDriveLetterFmtStr, fpath[:2])
		}
		// `X:\` needs to preserve the `\`, path.Base(filepath.ToSlash(fpath)) results in `X:` which is not valid
		if fpath[1:3] == ":\\" {
			return files.NewSerialFile(fpath, fpath, hidden, stat)
		}
	case 6:
		// `\\?\X:` long prefix form of `X:`, still ambiguous
		if fpath[:4] == "\\\\?\\" && fpath[5] == ':' {
			return nil, fmt.Errorf(winDriveLetterFmtStr, fpath)
		}
	case 7:
		// `\\?\X:\` long prefix form is translated into short form `X:\`
		if fpath[:4] == "\\\\?\\" && fpath[5] == ':' && fpath[6] == '\\' {
			fpath = string(fpath[4]) + ":\\"
			return files.NewSerialFile(fpath, fpath, hidden, stat)
		}
	}

	return files.NewSerialFile(path.Base(filepath.ToSlash(fpath)), fpath, hidden, stat)
}
