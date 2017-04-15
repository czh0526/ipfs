package main

import (
  "fmt"
  "os"
  cmds "github.com/czh0526/ipfs/commands"
)

func main() {
  os.Exit(mainRet())
}

func mainRet() int {
  var err error
  var invoc cmdInvocation
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
