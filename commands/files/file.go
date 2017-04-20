package files

import "io"

type File interface {
	io.ReadCloser
	FileName() string
	FullPath() string
	IsDirectory() bool
	NextFile() (File, error)
}
