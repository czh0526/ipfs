package commands

import ()

type Request interface {
  Path()  []string
}
