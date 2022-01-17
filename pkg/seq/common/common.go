// 29 Apr 2020

package common

import (
	"io"
	"fmt"
	"os"
)

const (
	ExitSuccess = iota
	ExitFailure
	ExitUsageError
)

const GapChar byte = '-'   // a minus sign is always used for gaps


// WrtTemp writes a string to a temporary file and returns
// the filename. It is used all over the place in testing.
func WrtTemp(s string) (string, error) {
	f_tmp, err := os.CreateTemp("", "_del_me_testing")
	if err != nil {
		return "", fmt.Errorf("tempfile fail")
	}

	if _, err := io.WriteString(f_tmp, s); err != nil {
		return "", fmt.Errorf("writing string to temp file %v", f_tmp.Name())
	}
	name := f_tmp.Name()
	f_tmp.Close()
	return name, nil
}
