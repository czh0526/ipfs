package commands

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type EncodingType string
type ErrorType uint

// Response 代表了一个request的结果，由Handlers构建
type Response interface {
	Request() Request
	// error
	SetError(err error, code ErrorType)
	Error() *Error
	//
	Output() interface{}
	Reader() (io.Reader, error)
}

const (
	JSON     = "json"
	XML      = "xml"
	Protobuf = "protobuf"
	Text     = "text"
)

const (
	ErrNormal ErrorType = iota
	ErrClient
	ErrNotFound
	ErrImplementation
)

type Error struct {
	Message string
	Code    ErrorType
}

func (e Error) Error() string {
	return e.Message
}

var marshallers = map[EncodingType]Marshaler{
	XML: func(res Response) (io.Reader, error) {
		var value interface{}
		if res.Error() != nil {
			value = res.Error()
		} else {
			value = res.Output()
		}

		b, err := xml.Marshal(value)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	},
}

type response struct {
	req    Request
	err    *Error
	value  interface{}
	out    io.Reader
	stdout io.Writer
	stderr io.Writer
	closer io.Closer
}

func (r *response) Request() Request {
	return r.req
}

func (r *response) Error() *Error {
	return r.err
}

func (r *response) Output() interface{} {
	return r.value
}

func (r *response) SetError(err error, code ErrorType) {
	r.err = &Error{Message: err.Error(), Code: code}
}

func (r *response) Marshal() (io.Reader, error) {
	if r.err == nil && r.value == nil {
		return bytes.NewReader([]byte{}), nil
	}

	enc, found, err := r.req.Option(EncShort).String()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("No encoding type was specified")
	}
	encType := EncodingType(strings.ToLower(enc))

	// Special case: if text encoding and an error, just print it out.
	if encType == Text && r.Error() != nil {
		return strings.NewReader(r.Error().Error()), nil
	}

	var marshaller Marshaler
	if r.req.Command() != nil && r.req.Command().Marshalers != nil {
		marshaller = r.req.Command().Marshalers[encType]
	}
	if marshaller == nil {
		var ok bool
		marshaller, ok = marshallers[encType]
		if !ok {
			return nil, fmt.Errorf("No marshaller found for encoding type '%s'", enc)
		}
	}

	output, err := marshaller(r)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return bytes.NewReader([]byte{}), nil
	}
	return output, nil
}

func (r *response) Reader() (io.Reader, error) {
	if r.out == nil {
		if out, ok := r.value.(io.Reader); ok {
			r.out = out
		} else {
			marshalled, err := r.Marshal()
			if err != nil {
				return nil, err
			}
			r.out = marshalled
		}
	}
	return r.out, nil
}

func NewResponse(req Request) Response {
	return &response{
		req:    req,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}
