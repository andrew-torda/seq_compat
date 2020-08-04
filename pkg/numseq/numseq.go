// 3 Aug 2020

package numseq

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

// byMmap
func byMmap(fname string) (int, error) {
	var fp *os.File
	var err error
	var mm mmap.MMap
	if fp, err = os.Open(fname); err != nil {
		return 0, err
	}
	defer fp.Close()
	if mm, err = mmap.Map(fp, mmap.RDONLY, 0); err != nil {
		return 0, err
	}
	defer mm.Unmap()
	return bytes.Count(mm, []byte(">")), nil
}

func byreadingFixed(fname string, bufsizeignored int) (int, error) {
	const bsize = 64 * 1024
	var buf [bsize]byte
	return (innerbyreading(fname, buf[:]))
}

func byreadingVaries(fname string, bufsize int) (int, error) {
	buf := make([]byte, bufsize)
	return (innerbyreading(fname, buf))
}

func innerbyreading(fname string, buf []byte) (int, error) {
	var fp io.ReadCloser
	var err error
	if fp, err = os.Open(fname); err != nil {
		return 0, err
	}
	defer fp.Close()
	count := 0
	bsize := len(buf)
	for n := bsize; n == bsize; {
		n, err = fp.Read(buf[:])
		if err != nil && err != io.EOF {
			return 0, err
		}
		count += bytes.Count(buf[:n], []byte(">"))
	}

	return count, nil
}

func Main(fname string) (int, error) {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		fmt.Println(err)
	}
	var tmp [100]byte
	print(tmp[0:20])
	return 0, nil
}
