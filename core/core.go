package core

import (
	goprocess "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
)

type IpfsNode struct {
	proc goprocess.Process
}

func (n *IpfsNode) Close() error {
	return n.proc.Close()
}
