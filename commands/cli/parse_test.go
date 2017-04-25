package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/czh0526/ipfs/commands"
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

func TestOptionParsing(t *testing.T) {
	fmt.Println("			-------------------------------------")
	fmt.Println("			|		Test Option Parse           |")
	fmt.Println("			-------------------------------------")

	subCmd := &commands.Command{}
	cmd := &commands.Command{
		Options: []commands.Option{
			commands.StringOption("string", "s", "a string"),
			commands.BoolOption("bool", "b", "a bool"),
		},
		Subcommands: map[string]*commands.Command{
			"test": subCmd,
		},
	}

	testHelper := func(args string, expectedOpts kvs, expectedWords words, expectErr bool) {
		var opts map[string]interface{}
		var input []string

		path, opts, input, _, err := parseOpts(strings.Split(args, " "), cmd)
		if expectErr {
			if err == nil {
				t.Errorf("Command line '%v' parsing should have failed", args)
			}
			fmt.Printf("args = %v,\t\t err = %v \n", args, err)

		} else if err != nil {
			t.Errorf("Command line '%v' failed to parse: %v", args, err)

		} else if !sameWords(input, expectedWords) || !sameKVs(opts, expectedOpts) {
			t.Errorf("Command line '%v':\n parsed as %v %v\n instead of %v %v",
				args, opts, input, expectedOpts, expectedWords)
		} else {
			fmt.Printf("args = %v,\t\t path = %v,\t\t stringVals = %v,\t\t opts = %v \n", args, path, input, opts)
		}
	}

	testFail := func(args string) {
		testHelper(args, kvs{}, words{}, true)
	}

	test := func(args string, expectedOpts kvs, expectedWords words) {
		testHelper(args, expectedOpts, expectedWords, false)
	}

	test("test -", kvs{}, words{"-"})
	testFail("-b -b")
	testFail("test -m 123")
	test("test beep boop", kvs{}, words{"beep", "boop"})
	test("-s foo", kvs{"s": "foo"}, words{})
	test("-sfoo", kvs{"s": "foo"}, words{})
	test("-s=foo", kvs{"s": "foo"}, words{})
	test("-b", kvs{"b": true}, words{})
	test("-bs foo", kvs{"b": true, "s": "foo"}, words{})
	test("-sb", kvs{"s": "b"}, words{})
	test("-b test foo", kvs{"b": true}, words{"foo"})
	test("--bool test foo", kvs{"bool": true}, words{"foo"})
	testFail("--bool=foo")
	testFail("--string")
	test("--string foo", kvs{"string": "foo"}, words{})
	test("--string=foo", kvs{"string": "foo"}, words{})
	test("-- -b", kvs{}, words{"-b"})
	test("test foo -b", kvs{"b": true}, words{"foo"})
	test("-b=false", kvs{"b": false}, words{})
	test("-b=true", kvs{"b": true}, words{})
	test("-b=false test foo", kvs{"b": false}, words{"foo"})
	test("-b=true test foo", kvs{"b": true}, words{"foo"})
	test("--bool=true test foo", kvs{"bool": true}, words{"foo"})
	test("--bool=false test foo", kvs{"bool": false}, words{"foo"})
	test("-b test true", kvs{"b": true}, words{"true"})
	test("-b test false", kvs{"b": true}, words{"false"})
	test("-b=FaLsE test foo", kvs{"b": false}, words{"foo"})
	test("-b=TrUe test foo", kvs{"b": true}, words{"foo"})
	test("-b test true", kvs{"b": true}, words{"true"})
	test("-b test false", kvs{"b": true}, words{"false"})
	test("-b --string foo test bar", kvs{"b": true, "string": "foo"}, words{"bar"})
	test("-b=false --string bar", kvs{"b": false, "string": "bar"}, words{})

	testFail("foo test")
}

func TestArgumentParsing(t *testing.T) {
	fmt.Println("			--------------------------------------")
	fmt.Println("			|		Test Argument Parse          |")
	fmt.Println("			--------------------------------------")
	// if runtime.GOOS == "windows" {
	// 	t.Skip("stdin handling doesn't yet work on windows")
	// }

	rootCmd := &commands.Command{
		Subcommands: map[string]*commands.Command{
			"noarg": {},
			"onearg": {
				Arguments: []commands.Argument{
					commands.StringArg("a", true, false, "some arg"),
				},
			},
			"twoargs": {
				Arguments: []commands.Argument{
					commands.StringArg("a", true, false, "some arg"),
					commands.StringArg("b", true, false, "another arg"),
				},
			},
			"variadic": {
				Arguments: []commands.Argument{
					commands.StringArg("a", true, true, "some arg"),
				},
			},
			"optional": {
				Arguments: []commands.Argument{
					commands.StringArg("b", false, true, "another arg"),
				},
			},
			"optionalsecond": {
				Arguments: []commands.Argument{
					commands.StringArg("a", true, false, "some arg"),
					commands.StringArg("b", false, false, "another arg"),
				},
			},
			"reversedoptional": {
				Arguments: []commands.Argument{
					commands.StringArg("a", false, false, "some arg"),
					commands.StringArg("b", true, false, "another arg"),
				},
			},
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
				t.Fatal(err)
			}
		}
		req, _, path, err := Parse(cmd, f, rootCmd)
		if err != nil {
			fmt.Printf("err = %v", err)
			t.Errorf("Command '%v' should have passed parsing: %v", cmd, err)
		}
		fmt.Printf("cmd = %v,\t\t path = %v, \t\t arguments = %v.\n", cmd, path, req.Arguments())

		if !sameWords(req.Arguments(), res) {
			t.Errorf("Arguments parsed from '%v' are '%v' instead of '%v'", cmd, req.Arguments(), res)
		}
	}

	testFail := func(cmd words, fi *os.File, msg string) {
		req, _, path, err := Parse(cmd, nil, rootCmd)
		if err == nil {
			t.Errorf("Should have failed: %v", msg)
		}
		fmt.Printf("cmd = %v,\t\t path = %v,\t\t arguments = %v \t\t err=%v.\n", cmd, path, req.Arguments(), err)
	}

	fmt.Println("==> noarg")
	test([]string{"noarg"}, nil, []string{})
	testFail([]string{"noarg", "value!"}, nil, "provided an arg, but command didn't define any")

	fmt.Println("==> onearg")
	test([]string{"onearg", "value!"}, nil, []string{"value!"})
	testFail([]string{"onearg"}, nil, "didn't provide any args, arg is required")

	fmt.Println("==> twoargs")
	test([]string{"twoargs", "value1", "value2"}, nil, []string{"value1", "value2"})
	testFail([]string{"twoargs", "value!"}, nil, "only provided 1 arg, needs 2")
	testFail([]string{"twoargs"}, nil, "didn't provide any args, 2 required")

	fmt.Println("==> variadic")
	test([]string{"variadic", "value!"}, nil, []string{"value!"})
	test([]string{"variadic", "value1", "value2", "value3"}, nil, []string{"value1", "value2", "value3"})
	testFail([]string{"variadic"}, nil, "didn't provide any args, 1 required")

	fmt.Println("==> optional")
	test([]string{"optional", "value!"}, nil, []string{"value!"})
	test([]string{"optional"}, nil, []string{})
	test([]string{"optional", "value1", "value2"}, nil, []string{"value1", "value2"})

	fmt.Println("==> optionalsecond")
	test([]string{"optionalsecond", "value!"}, nil, []string{"value!"})
	test([]string{"optionalsecond", "value1", "value2"}, nil, []string{"value1", "value2"})
	testFail([]string{"optionalsecond"}, nil, "didn't provide any args, 1 required")
	testFail([]string{"optionalsecond", "value1", "value2", "value3"}, nil, "provided too many args, takes 2 maximum")

	fmt.Println("==> recersedoptional")
	test([]string{"reversedoptional", "value1", "value2"}, nil, []string{"value1", "value2"})
	test([]string{"reversedoptional", "value1"}, nil, []string{"value1"})
	testFail([]string{"reversedoptional"}, nil, "didn't provide any args, 1 required")
	testFail([]string{"reversedoptional", "value1", "value2", "value3"}, nil, "provided too many args, only takes 1")

	fmt.Println("==> Stdin Enabled")
	fileToSimulateStdin := func(t *testing.T, content string) *os.File {
		fstdin, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(fstdin.Name())

		if _, err := io.WriteString(fstdin, content); err != nil {
			t.Fatal(err)
		}
		return fstdin
	}

	test([]string{"stdinenabled", "value1", "value2"}, nil, []string{"value1", "value2"})
	fstdin := fileToSimulateStdin(t, "stdin1\nstdin2")
	test([]string{"stdinenabled"}, fstdin, []string{"stdin1"})
}
