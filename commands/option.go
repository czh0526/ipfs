package commands

import (
	"fmt"
	"reflect"
	"strings"

	util "gx/ipfs/QmZuY8aV7zbNXVy6DyN9SmnuH3o9nG852F4aTiSBpts8d1/go-ipfs-util"
)

// Types of Command options
const (
	Invalid = reflect.Invalid
	Bool    = reflect.Bool
	Int     = reflect.Int
	Uint    = reflect.Uint
	Float   = reflect.Float64
	String  = reflect.String
)

// Option is used to specify a field that will be provided by a consumer
type Option interface {
	Names() []string
	Type() reflect.Kind
	Description() string
	Default(interface{}) Option
	DefaultVal() interface{}
}

type option struct {
	names       []string
	kind        reflect.Kind
	description string
	defaultVal  interface{}
}

func (o *option) Names() []string {
	return o.names
}

func (o *option) Type() reflect.Kind {
	return o.kind
}

func (o *option) Description() string {
	if len(o.description) == 0 {
		return ""
	}
	if !strings.HasSuffix(o.description, ".") {
		o.description += "."
	}
	if o.defaultVal != nil {
		if strings.Contains(o.description, "<<default>>") {
			return strings.Replace(o.description, "<<default>>", fmt.Sprintf("Default: %v.", o.defaultVal), -1)
		} else {
			return fmt.Sprintf("%s Default: %v.", o.description, o.defaultVal)
		}
	}
	return o.description
}

func (o *option) Default(v interface{}) Option {
	o.defaultVal = v
	return o
}

func (o *option) DefaultVal() interface{} {
	return o.defaultVal
}

func NewOption(kind reflect.Kind, names ...string) Option {
	if len(names) < 2 {
		panic("Options require at least two string values (name and description)")
	}
	desc := names[len(names)-1]
	names = names[:len(names)-1]

	return &option{
		names:       names,
		kind:        kind,
		description: desc,
	}
}

func BoolOption(names ...string) Option {
	return NewOption(Bool, names...)
}

func IntOption(names ...string) Option {
	return NewOption(Int, names...)
}

func StringOption(names ...string) Option {
	return NewOption(String, names...)
}

type OptionValue struct {
	value interface{}
	found bool
	def   Option
}

func (ov OptionValue) Found() bool {
	return ov.found
}

func (ov OptionValue) Definition() Option {
	return ov.def
}

func (ov OptionValue) Bool() (value bool, found bool, err error) {
	if !ov.found && ov.value == nil {
		return false, false, nil
	}
	val, ok := ov.value.(bool)
	if !ok {
		err = util.ErrCast()
	}
	return val, ov.found, err
}

func (ov OptionValue) String() (value string, found bool, err error) {
	if !ov.found && ov.value == nil {
		return "", false, nil
	}
	val, ok := ov.value.(string)
	if !ok {
		err = util.ErrCast()
	}
	return val, ov.found, err
}

// Flag names
const (
	EncShort   = "enc"
	EncLong    = "encoding"
	RecShort   = "r"
	RecLong    = "recursive"
	ChanOpt    = "stream-channels"
	TimeoutOpt = "timeout"
)

// options that are used by this package
var OptionEncodingType = StringOption(EncLong, EncShort, "The encoding type the output should be encoded with (json, xml, or text)")
var OptionRecursivePath = BoolOption(RecLong, RecShort, "Add directory paths recursively").Default(false)
var OptionStreamChannels = BoolOption(ChanOpt, "Stream channel output")
var OptionTimeout = StringOption(TimeoutOpt, "set a global timeout on the command")

var globalOptions = []Option{
	OptionEncodingType,
	OptionStreamChannels,
	OptionTimeout,
}

var globalCommand = &Command{
	Options: globalOptions,
}
