package main

import (
  "context"
  "os"
  cmds "github.com/czh0526/ipfs/commands"
  logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
  loggables "gx/ipfs/QmXs1igHHEaUmMxKtbP8Z9wTjitQ75sqxaKQP4QgnLN4nn/go-libp2p-loggables"
)

func main() {
  os.Exit(mainRet())
}

func mainRet() int {
  var err error
  var invoc cmdInvocation
  ctx := logging.ContextWithLoggable(context.Background(), loggables.Uuid("session"))
  defer invoc.close()

  parseErr := invoc.Parse(ctx, os.Args[1:])
  if invoc.req != nil {

  }
}

type cmdInvocation struct {
  path  []string
  cmd   *cmds.Command
  req   cmds.Request
}

func (i *cmdInvocation) Parse(ctx context.Context, args []string) error {
  var err   error

}
