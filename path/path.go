package path

import "strings"

func Join(pths []string) string {
	return strings.Join(pths, "/")
}
