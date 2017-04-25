package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	loggables "gx/ipfs/QmXs1igHHEaUmMxKtbP8Z9wTjitQ75sqxaKQP4QgnLN4nn/go-libp2p-loggables"
	osh "gx/ipfs/QmXuBJ7DR6k3rmUEKtvVMhwjmXDuJgXXPUt4LQXKBMsU93/go-os-helper"

	u "gx/ipfs/QmZuY8aV7zbNXVy6DyN9SmnuH3o9nG852F4aTiSBpts8d1/go-ipfs-util"

	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	manet "gx/ipfs/QmVCNGTyD4EkvNYaAp253uMQ9Rjsjy2oGMvcdJJUoVRfja/go-multiaddr-net"

	cmds "github.com/czh0526/ipfs/commands"
	cmdsCli "github.com/czh0526/ipfs/commands/cli"
	cmdsHttp "github.com/czh0526/ipfs/commands/http"
	core "github.com/czh0526/ipfs/core"
	coreCmds "github.com/czh0526/ipfs/core/commands"
	repo "github.com/czh0526/ipfs/repo"
	config "github.com/czh0526/ipfs/repo/config"
	fsrepo "github.com/czh0526/ipfs/repo/fsrepo"
)

func init() {
	logging.SetDebugLogging()
}

var log = logging.Logger("cmd/ipfs")

var (
	errUnexpectedApiOutput = errors.New("api returned unexpected output")
	errRequestCanceled     = errors.New("request canceled")
)

func main() {
	fmt.Printf("当前目录：%s.\n", os.Args[0])
	now := time.Now()
	log.Debugf("System begin at %s.", now.Format("2006-01-02 15:04:05"))
	ret := mainRet()
	log.Debugf("System Exit with code %v.", ret)
	os.Exit(ret)
}

func mainRet() int {
	var invoc cmdInvocation
	ctx := logging.ContextWithLoggable(context.Background(), loggables.Uuid("session"))
	defer invoc.close()

	printErr := func(err error) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}

	printHelp := func(long bool, w io.Writer) {
		helpFunc := cmdsCli.ShortHelp
		if long {
			helpFunc = cmdsCli.LongHelp
		}
		helpFunc("ipfs", Root, invoc.path, w)
	}

	parseErr := invoc.Parse(ctx, os.Args[1:])
	if invoc.req != nil {
		longH, shortH, err := invoc.requestedHelp()
		if err != nil {
			printErr(err)
			return 1
		}
		if longH || shortH {
			printHelp(longH, os.Stdout)
			return 0
		}
	}

	if parseErr != nil {
		printErr(parseErr)
		if invoc.cmd != nil {
			fmt.Fprintf(os.Stderr, "\n")
			printHelp(false, os.Stderr)
		}
		return 1
	}

	if invoc.cmd == nil || invoc.cmd.Run == nil {
		printHelp(false, os.Stdout)
		return 0
	}

	_, err := invoc.Run(ctx)
	if err != nil {
		printErr(err)
	}

	return 0
}

func getRepoPath(req cmds.Request) (string, error) {
	repoOpt, found, err := req.Option("config").String()
	if err != nil {
		return "", err
	}
	if found && repoOpt != "" {
		return repoOpt, nil
	}

	repoPath, err := fsrepo.BestKnownPath()
	if err != nil {
		return "", err
	}
	return repoPath, nil
}

func loadConfig(path string) (*config.Config, error) {
	return fsrepo.ConfigAt(path)
}

type cmdInvocation struct {
	path []string
	cmd  *cmds.Command
	req  cmds.Request
	node *core.IpfsNode
}

func (i *cmdInvocation) Parse(ctx context.Context, args []string) error {
	var err error

	log.Debugf("args = %v.", args)
	i.req, i.cmd, i.path, err = cmdsCli.Parse(args, os.Stdin, Root)
	if err != nil {
		return err
	}

	repoPath, err := getRepoPath(i.req)
	if err != nil {
		return err
	}
	log.Debugf("config path is %s", repoPath)

	cmdctx := i.req.InvocContext()
	cmdctx.ConfigRoot = repoPath
	cmdctx.LoadConfig = loadConfig
	cmdctx.ConstructNode = i.constructNodeFunc(ctx)

	if !i.req.Option("encoding").Found() {
		if i.req.Command().Marshalers != nil && i.req.Command().Marshalers[cmds.Text] != nil {
			i.req.SetOption("encoding", cmds.Text)
		} else {
			i.req.SetOption("encoding", cmds.JSON)
		}
	}
	return nil
}

func (i *cmdInvocation) close() {
	if i.node != nil {
		i.node.Close()
	}
}

func (i *cmdInvocation) requestedHelp() (short bool, long bool, err error) {
	longHelp, _, err := i.req.Option("help").Bool()
	if err != nil {
		return false, false, err
	}
	shortHelp, _, err := i.req.Option("h").Bool()
	if err != nil {
		return false, false, err
	}
	return longHelp, shortHelp, nil
}

func (i *cmdInvocation) constructNodeFunc(ctx context.Context) func() (*core.IpfsNode, error) {
	return func() (*core.IpfsNode, error) {
		return &core.IpfsNode{}, nil
	}
}

func (i *cmdInvocation) Run(ctx context.Context) (output io.Reader, err error) {
	debug, _, err := i.req.Option("debug").Bool()
	if err != nil {
		return nil, err
	}
	if debug || os.Getenv("IPFS_LOGGING") == "debug" {
		u.Debug = true
		logging.SetDebugLogging()
	}
	if u.GetenvBool("DEBUG") {
		u.Debug = true
	}

	res, err := callCommand(ctx, i.req, Root, i.cmd)
	if err != nil {
		return nil, err
	}

	if err := res.Error(); err != nil {
		return nil, err
	}

	return res.Reader()
}

func callPreCommandHooks(ctx context.Context, details cmdDetails, req cmds.Request, root *cmds.Command) error {
	log.Debug("calling pre-command hooks...")
	return nil
}

func callCommand(ctx context.Context, req cmds.Request, root *cmds.Command, cmd *cmds.Command) (cmds.Response, error) {
	var res cmds.Response
	fmt.Printf("cmd = %v.\n", reflect.TypeOf(cmd))
	err := req.SetRootContext(ctx)
	if err != nil {
		return nil, err
	}

	details, err := commandDetails(req.Path(), root)
	if err != nil {
		return nil, err
	}

	client, err := commandShouldRunOnDaemon(*details, req, root)
	if err != nil {
		return nil, err
	}

	err = callPreCommandHooks(ctx, *details, req, root)
	if err != nil {
		return nil, err
	}

	if cmd.PreRun != nil {
		err = cmd.PreRun(req)
		if err != nil {
			return nil, err
		}
	}

	if client != nil && !cmd.External {
		log.Debug("executing command via API")
		res, err = client.Send(req)
		if err != nil {
			if isConnRefused(err) {
				err = repo.ErrApiNotRunning
			}
			return nil, wrapContextCanceled(err)
		}
	} else {
		log.Debug("executing command locally")
		err := req.SetRootContext(ctx)
		if err != nil {
			return nil, err
		}
		res = root.Call(req)
	}

	if cmd.PostRun != nil {
		cmd.PostRun(req, res)
	}

	return res, nil
}

func commandDetails(path []string, root *cmds.Command) (*cmdDetails, error) {
	var details cmdDetails
	cmd := root
	for _, cmp := range path {
		var found bool
		cmd, found = cmd.Subcommands[cmp]
		if !found {
			return nil, fmt.Errorf("Subcommand %s should be in root", cmp)
		}

		if cmdDetails, found := cmdDetailsMap[cmd]; found {
			details = cmdDetails
		}
	}
	return &details, nil
}

func commandShouldRunOnDaemon(details cmdDetails, req cmds.Request, root *cmds.Command) (cmdsHttp.Client, error) {
	path := req.Path()
	if len(path) < 1 {
		return nil, nil
	}

	if details.cannotRunOnClient && details.cannotRunOnDaemon {
		return nil, fmt.Errorf("cammand disabled: %s", path[0])
	}

	if details.doesNotUseRepo && details.canRunOnClient() {
		return nil, nil
	}

	apiAddrStr, _, err := req.Option(coreCmds.ApiOption).String()
	if err != nil {
		return nil, err
	}

	client, err := getApiClient(req.InvocContext().ConfigRoot, apiAddrStr)
	if err == repo.ErrApiNotRunning {
		if apiAddrStr != "" && req.Command() != daemonCmd {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if client != nil {
		if details.cannotRunOnDaemon {
			log.Debugf("Command cannot run on daemon, Checking if daemon is locked")
			if daemonLocked, _ := fsrepo.LockedByOtherProcess(req.InvocContext().ConfigRoot); daemonLocked {
				return nil, cmds.ClientError("ipfs daemon is running. please stop it to run this command")
			}
			return nil, nil
		}
		return client, nil
	}

	if details.cannotRunOnClient {
		return nil, cmds.ClientError("must run on the ipfs daemon")
	}

	return nil, nil
}

var apiFileErrorFmt string = `Failed to parse '%[1]s/api' file. \nerror: %[2]s \nif you're sure go-ipfs isn't running, you can just delete it.`
var checkIPFSUnixFmt = "Otherwise check :\n\tps aux | grep ipfs"
var checkIPFSWinFmt = "Otherwise check:\n\ttasklist | findstr ipfs"

func getApiClient(repoPath, apiAddrStr string) (cmdsHttp.Client, error) {
	var apiErrorFmt string
	switch {
	case osh.IsUnix():
		apiErrorFmt = apiFileErrorFmt + checkIPFSUnixFmt
	case osh.IsWindows():
		apiErrorFmt = apiFileErrorFmt + checkIPFSWinFmt
	default:
		apiErrorFmt = apiFileErrorFmt
	}

	var addr ma.Multiaddr
	var err error
	if len(apiAddrStr) != 0 {
		addr, err = ma.NewMultiaddr(apiAddrStr)
		if err != nil {

		}
		if len(addr.Protocols()) == 0 {
			return nil, fmt.Errorf("multiaddr doesn't provide any protocols")
		}
	} else {
		addr, err = fsrepo.APIAddr(repoPath)
		if err == repo.ErrApiNotRunning {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf(apiErrorFmt, repoPath, err.Error())
		}
	}

	if len(addr.Protocols()) == 0 {
		return nil, fmt.Errorf(apiErrorFmt, repoPath, "multiaddr doesn't provide any protocols")
	}
	return apiClientForAddr(addr)
}

func apiClientForAddr(addr ma.Multiaddr) (cmdsHttp.Client, error) {
	_, host, err := manet.DialArgs(addr)
	if err != nil {
		return nil, err
	}
	return cmdsHttp.NewClient(host), nil
}

func isConnRefused(err error) bool {
	if urlerr, ok := err.(*url.Error); ok {
		err = urlerr.Err
	}
	netoperr, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	return netoperr.Op == "dial"
}

func wrapContextCanceled(err error) error {
	if strings.Contains(err.Error(), "request canceled") {
		err = errRequestCanceled
	}
	return err
}
