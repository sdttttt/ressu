//go:build windows

package logging

import "os"

func expectedFilePerm() os.FileMode {
	return 0o666
}
