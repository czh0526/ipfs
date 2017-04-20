package fsrepo

import (
	"os"

	"github.com/czh0526/ipfs/repo/config"
	homedir "github.com/mitchellh/go-homedir"
)

func BestKnownPath() (string, error) {
	ipfsPath := config.DefaultPathRoot
	if os.Getenv(config.EnvDir) != "" {
		ipfsPath = os.Getenv(config.EnvDir)
	}
	ipfsPath, err := homedir.Expand(ipfsPath)
	if err != nil {
		return "", err
	}

	return ipfsPath, nil
}
