package main

import (
	"context"
	"fmt"
	"os"
	"time"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	loggables "gx/ipfs/QmXs1igHHEaUmMxKtbP8Z9wTjitQ75sqxaKQP4QgnLN4nn/go-libp2p-loggables"

	cmds "github.com/czh0526/ipfs/commands"
	cmdsCli "github.com/czh0526/ipfs/commands/cli"
	core "github.com/czh0526/ipfs/core"
	fsrepo "github.com/czh0526/ipfs/repo/fsrepo"
)

func init() {
	logging.SetDebugLogging()
}

var log = logging.Logger("cmd/ipfs")

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

	invoc.Parse(ctx, os.Args[1:])
	if invoc.req != nil {

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

	return nil
}

func (i *cmdInvocation) close() {
	if i.node != nil {
		i.node.Close()
	}
}
