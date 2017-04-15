package commands

import (
)

type Function func(Request, Response)

type Command struct {
  Options     []Option
  Arguments   []Argument

  PreRun      func(req Request) error
  Run         Function
  PostRun     Function
}

type HelpText struct {
  Tagline                 string
  ShortDescription        string
  SynopsisOptionsValues   map[string]string
  Usage                   string
  LongDescription         string
  Options                 string
  Arguments               string
  Subcommands             string
  Synopsis                string 
}
