package main

import (
  "fmt"
  "os"
)

func main() {
  os.Exit(mainRet())
}

func mainRet() int {
  fmt.Println("Hello IPFS .")
  return 1
}
