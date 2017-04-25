package fsrepo

import (
	"io"
	"os"
	"path/filepath"

	"strings"

	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	lockfile "github.com/czh0526/ipfs/repo/fsrepo/lock"

	"github.com/czh0526/ipfs/repo"
	"github.com/czh0526/ipfs/repo/config"
)

var log = logging.Logger("fsrepo")

const apiFile = "api"
const swarmKeyFile = "swarm.key"

func ConfigAt(repoPath string) (*config.Config, error) {
	return &config.Config{}, nil
}

func APIAddr(repoPath string) (ma.Multiaddr, error) {
	repoPath = filepath.Clean(repoPath)
	apiFilePath := filepath.Join(repoPath, apiFile)

	f, err := os.Open(apiFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, repo.ErrApiNotRunning
		}
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, 2048)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	s := string(buf[:n])
	s = strings.TrimSpace(s)
	return ma.NewMultiaddr(s)
}

func LockedByOtherProcess(repoPath string) (bool, error) {
	repoPath = filepath.Clean(repoPath)
	locked, err := lockfile.Locked(repoPath)
	if locked {
		log.Debugf("(%t)<->Lock is held at %s", locked, repoPath)
	}
	return locked, err
}
