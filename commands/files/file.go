package files

import (
	"errors"
	"io"
)

var (
	ErrNotDirectory = errors.New("Couldn't call NextFile(), this isn't a directory")
	ErrNotReader    = errors.New("This file is a directory, can't use Reader functions")
)

type File interface {
	io.ReadCloser
	FileName() string
	FullPath() string
	IsDirectory() bool
	NextFile() (File, error)
}
