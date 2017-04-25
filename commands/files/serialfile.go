package files

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type serialFile struct {
	name              string
	path              string
	files             []os.FileInfo
	stat              os.FileInfo
	current           *File
	handleHiddenFiles bool
}

func NewSerialFile(name, path string, hidden bool, stat os.FileInfo) (File, error) {

	switch mode := stat.Mode(); {
	case mode.IsRegular():
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return NewReaderPathFile(name, path, file, stat)
	case mode.IsDir():
		contents, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		return &serialFile{name, path, contents, stat, nil, hidden}, nil
	case mode&os.ModeSymlink != 0:
		target, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}
		return NewLinkFile(name, path, target, stat), nil
	default:
		return nil, fmt.Errorf("Unrecognized file type for %s: %s", name, mode.String())
	}
}

func (f *serialFile) IsDirectory() bool {
	return true
}

func (f *serialFile) NextFile() (File, error) {
	err := f.Close()
	if err != nil {
		return nil, err
	}

	if len(f.files) == 0 {
		return nil, io.EOF
	}

	stat := f.files[0]
	f.files = f.files[1:]
	for !f.handleHiddenFiles && strings.HasPrefix(stat.Name(), ".") {
		if len(f.files) == 0 {
			return nil, io.EOF
		}
		stat = f.files[0]
		f.files = f.files[1:]
	}

	fileName := filepath.ToSlash(filepath.Join(f.name, stat.Name()))
	filePath := filepath.ToSlash(filepath.Join(f.path, stat.Name()))

	sf, err := NewSerialFile(fileName, filePath, f.handleHiddenFiles, stat)
	if err != nil {
		return nil, err
	}

	f.current = &sf
	return sf, nil
}

func (f *serialFile) FileName() string {
	return f.name
}

func (f *serialFile) FullPath() string {
	return f.path
}

func (f *serialFile) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (f *serialFile) Close() error {
	if f.current != nil {
		err := (*f.current).Close()
		if err != nil && err != syscall.EINVAL {
			return err
		}
	}
	return nil
}

func (f *serialFile) Stat() os.FileInfo {
	return f.stat
}

func (f *serialFile) Size() (int64, error) {
	if !f.stat.IsDir() {
		return f.stat.Size(), nil
	}

	var du int64
	err := filepath.Walk(f.FullPath(), func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi != nil && fi.Mode()&(os.ModeSymlink|os.ModeNamedPipe) == 0 {
			du += fi.Size()
		}
		return nil
	})

	return du, err
}
