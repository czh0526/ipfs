package commands

import (
	"context"
	"io"
	"os"

	"github.com/czh0526/ipfs/commands/files"
)

type OptMap map[string]interface{}

type Context struct {
	Online bool
}

// Request 代表了一个 consumer 对 command 的调用
type Request interface {
	Path() []string
	Option(name string) *OptionValue
	Options() OptMap
	SetOption(name string, val interface{})
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

	return nil
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
