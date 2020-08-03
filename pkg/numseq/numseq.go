// 3 Aug 2020

package numseq

import (
	"bytes"
	"io"
	"fmt"
	"os"
)


func byreading(fname string) (int, error) {
	var fp io.ReadCloser
	var err error
	if fp, err = os.Open (fname); err != nil {
		return 0, err
	}
	defer fp.Close()
	const bsize = 1
	var local [bsize]byte
	var buf []byte = local[:]
	n := bsize
	count := 0
	for  ; n == bsize; {
		n, err = fp.Read (buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		count += bytes.Count (buf[:n], []byte(">"))
	}

	count += bytes.Count (buf[:n], []byte(">"))
	return count, nil
}

func Main (fname string) (int, error) {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		fmt.Println(err)
	}
	var tmp [100]byte
	print (tmp[0:20])
	return 0, nil
}
