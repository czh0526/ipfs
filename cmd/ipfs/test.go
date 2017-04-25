package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/czh0526/ipfs/commands"
	"github.com/czh0526/ipfs/commands/cli"
)

type kvs map[string]interface{}
type words []string

func sameWords(a words, b words) bool {
	if len(a) != len(b) {
		return false
	}
	for i, w := range a {
		if w != b[i] {
			return false
		}
	}
	return true
}

func sameKVs(a kvs, b kvs) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func tttmain() {
	rootCmd := &commands.Command{
		Subcommands: map[string]*commands.Command{
			"noarg": {},
			"stdinenabled": {
				Arguments: []commands.Argument{
					commands.StringArg("a", true, true, "some arg").EnableStdin(),
				},
			},
		},
	}

	test := func(cmd words, f *os.File, res words) {
		if f != nil {
			if _, err := f.Seek(0, os.SEEK_SET); err != nil {
				fmt.Printf("%v\n", err)
			}
		}
		req, _, path, err := cli.Parse(cmd, f, rootCmd)
		if err != nil {
			fmt.Printf("err = %v", err)
			fmt.Printf("Command '%v' should have passed parsing: %v", cmd, err)
		}
		fmt.Printf("cmd = %v,\t\t path = %v, \t\t arguments = %v.\n", cmd, path, req.Arguments())

		if !sameWords(req.Arguments(), res) {
			fmt.Printf("Arguments parsed from '%v' are '%v' instead of '%v'", cmd, req.Arguments(), res)
		}
	}

	fmt.Println("==> Stdin Enabled")
	fileToSimulateStdin := func(content string) *os.File {
		fstdin, err := ioutil.TempFile("", "")
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		defer os.Remove(fstdin.Name())

		if _, err := io.WriteString(fstdin, content); err != nil {
			fmt.Printf("%v\n", err)
		}
		return fstdin
	}

	fstdin := fileToSimulateStdin("stdin1\nstdin2")
	test([]string{"stdinenabled"}, fstdin, []string{"stdin1"})
}
