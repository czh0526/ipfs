package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"

	u "gx/ipfs/QmZuY8aV7zbNXVy6DyN9SmnuH3o9nG852F4aTiSBpts8d1/go-ipfs-util"

	"bufio"

	"github.com/czh0526/ipfs/commands/files"
	"github.com/czh0526/ipfs/core"
	"github.com/czh0526/ipfs/repo/config"
)

type OptMap map[string]interface{}

type Context struct {
	Online        bool
	ConfigRoot    string
	config        *config.Config
	LoadConfig    func(path string) (*config.Config, error)
	node          *core.IpfsNode
	ConstructNode func() (*core.IpfsNode, error)
}

// Request 代表了一个 consumer 对 command 的调用
type Request interface {
	Path() []string
	// options
	Option(name string) *OptionValue
	Options() OptMap
	SetOption(name string, val interface{})
	// args
	Arguments() []string
	StringArguments() []string
	SetArguments([]string)
	// context
	Context() context.Context
	SetRootContext(context.Context) error
	InvocContext() *Context
	SetInvocContext(Context)
	// others
	Command() *Command

	ConvertOptions() error
}

type request struct {
	path       []string
	options    OptMap
	arguments  []string
	files      files.File
	cmd        *Command
	ctx        Context
	rctx       context.Context
	optionDefs map[string]Option
	values     map[string]interface{}
	stdin      io.Reader
}

func (r *request) Path() []string {
	return r.path
}

func (r *request) Option(name string) *OptionValue {
	option, found := r.optionDefs[name]
	if !found {
		return nil
	}
	fmt.Printf("option = %v, found = %v \n", option, found)
	for _, n := range option.Names() {
		val, found := r.options[n]
		if found {
			return &OptionValue{val, found, option}
		}
	}
	return &OptionValue{option.DefaultVal(), false, option}
}

func (r *request) Options() OptMap {
	output := make(OptMap)
	for k, v := range r.options {
		output[k] = v
	}
	return output
}

func (r *request) SetOption(name string, val interface{}) {
	option, found := r.optionDefs[name]
	if !found {
		return
	}

	for _, n := range option.Names() {
		_, found := r.options[n]
		if found {
			r.options[n] = val
			return
		}
	}
	r.options[name] = val
}

func (r *request) ConvertOptions() error {
	// 遍历 options
	for k, v := range r.options {
		// 根据 optionDef 进行校准
		opt, ok := r.optionDefs[k]
		if !ok {
			continue
		}

		// 根据Option.Type()，进行取值的转换
		kind := reflect.TypeOf(v).Kind()
		if kind != opt.Type() {
			if kind == String {
				// v ==> string
				str, ok := v.(string)
				if !ok {
					return u.ErrCast()
				}
				// string ==> opt.Type()
				convert := converters[opt.Type()]
				val, err := convert(str)
				if err != nil {
					value := fmt.Sprintf("value '%v'", v)
					if len(str) == 0 {
						value = "empty value"
					}
					return fmt.Errorf("Could not convert %s to type '%s' (for option '-%s')",
						value, opt.Type().String(), k)
				}
				r.options[k] = val
			} else {
				return fmt.Errorf("Option '%s' should be type '%s', but got type '%s'",
					k, opt.Type().String(), kind.String())
			}
		} else {
			r.options[k] = v
		}
	}

	return nil
}

func (r *request) StringArguments() []string {
	return r.arguments
}

func (r *request) Arguments() []string {
	if r.haveVarArgsFromStdin() {
		err := r.VarArgs(func(s string) error {
			r.arguments = append(r.arguments, s)
			return nil
		})
		if err != nil && err != io.EOF {
			log.Error(err)
		}
	}
	return r.arguments
}

func (r *request) SetArguments(args []string) {
	r.arguments = args
}

func (r *request) haveVarArgsFromStdin() bool {
	if len(r.cmd.Arguments) == 0 {
		return false
	}

	last := r.cmd.Arguments[len(r.cmd.Arguments)-1]
	return last.SupportsStdin && last.Type == ArgString && (last.Required || last.Variadic) &&
		len(r.arguments) < len(r.cmd.Arguments)
}

func (r *request) VarArgs(f func(string) error) error {
	if len(r.arguments) >= len(r.cmd.Arguments) {
		for _, arg := range r.arguments[len(r.cmd.Arguments)-1:] {
			err := f(arg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if r.files == nil {
		return nil
	}

	fi, err := r.files.NextFile()
	if err != nil {
		return err
	}

	var any bool
	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		any = true
		err := f(scan.Text())
		if err != nil {
			return err
		}
	}
	if !any {
		return f("")
	}

	return nil
}

func (r *request) Context() context.Context {
	return r.rctx
}

func (r *request) InvocContext() *Context {
	return &r.ctx
}

func (r *request) SetInvocContext(ctx Context) {
	r.ctx = ctx
}

func (r *request) SetRootContext(ctx context.Context) error {
	ctx, err := getContext(ctx, r)
	if err != nil {
		return err
	}
	r.rctx = ctx
	return nil
}

func (r *request) Command() *Command {
	return r.cmd
}

func getContext(base context.Context, req Request) (context.Context, error) {
	tout, found, err := req.Option("timeout").String()
	if err != nil {
		return nil, fmt.Errorf("error parsing timeout option: %s", err)
	}

	var ctx context.Context
	if found {
		duration, err := time.ParseDuration(tout)
		if err != nil {
			return nil, fmt.Errorf("error parsing timeout option: %s", err)
		}
		tctx, _ := context.WithTimeout(base, duration)
		ctx = tctx
	} else {
		cctx, _ := context.WithCancel(base)
		ctx = cctx
	}
	return ctx, nil
}

func NewRequest(path []string, opts OptMap, args []string, file files.File, cmd *Command, optDefs map[string]Option) (Request, error) {
	if opts == nil {
		opts = make(OptMap)
	}
	if optDefs == nil {
		optDefs = make(map[string]Option)
	}

	ctx := Context{}
	values := make(map[string]interface{})
	req := &request{
		path:       path,
		options:    opts,
		arguments:  args,
		files:      file,
		cmd:        cmd,
		ctx:        ctx,
		optionDefs: optDefs,
		values:     values,
		stdin:      os.Stdin,
	}
	err := req.ConvertOptions()
	if err != nil {
		return nil, err
	}

	return req, nil
}

type converter func(string) (interface{}, error)

var converters = map[reflect.Kind]converter{
	Bool: func(v string) (interface{}, error) {
		if v == "" {
			return true, nil
		}
		return strconv.ParseBool(v)
	},
	Int: func(v string) (interface{}, error) {
		val, err := strconv.ParseInt(v, 0, 32)
		if err != nil {
			return nil, err
		}
		return int(val), err
	},
	Uint: func(v string) (interface{}, error) {
		val, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			return nil, err
		}
		return int(val), err
	},
	Float: func(v string) (interface{}, error) {
		return strconv.ParseFloat(v, 64)
	},
}
