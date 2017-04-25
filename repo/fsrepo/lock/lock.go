package lock

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	util "github.com/ipfs/go-ipfs-util"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	lock "gx/ipfs/QmWi28zbQG6B1xfaaWx5cYoLn3kBFU6pQ6GWQNRV5P6dNe/lock"
)

var log = logging.Logger("lock")

const LockFile = "repo.lock"

func errPerm(path string) error {
	return fmt.Errorf("failed to take lock at %s: permission denied", path)
}

func Lock(confdir string) (io.Closer, error) {
	return lock.Lock(path.Join(confdir, LockFile))
}

func Locked(confdir string) (bool, error) {
	log.Debugf("Checking lock")
	if !util.FileExists(path.Join(confdir, LockFile)) {
		log.Debugf("File doesn't exist: %s", path.Join(confdir, LockFile))
		return false, nil
	}

	if lk, err := Lock(confdir); err != nil {
		if err == syscall.EAGAIN {
			log.Debugf("Someone else has the lock: %s", path.Join(confdir, LockFile))
			return true, nil
		}
		if strings.Contains(err.Error(), "resource temporarily unavailable") {
			log.Debugf("Can't lock file: %s.\n reason: %s", path.Join(confdir, LockFile), err.Error())
			return true, nil
		}

		if os.IsPermission(err) {
			return false, errPerm(confdir)
		}
		if isLockCreatePermFail(err) {
			return false, errPerm(confdir)
		}

		return false, err

	} else {
		log.Debugf("No one has a lock")
		lk.Close()
		return false, nil
	}
}

func isLockCreatePermFail(err error) bool {
	s := err.Error()
	return strings.Contains(s, "Lock Create of") && strings.Contains(s, "permission denied")
}
