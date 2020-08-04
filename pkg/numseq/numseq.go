// 3 Aug 2020

package numseq

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

// byReadFile is half the speed of the buffered versions.
// Try a version which gets the size, allocates a buffer and reads
// in one slurp.
func byReadFile(fname string, bufsizeignored int) (int, error) {
	if buf, err := ioutil.ReadFile(fname); err != nil {
		return 0, err
	} else {
		return bytes.Count(buf, []byte(">")), nil
	}
}

// byOneSlurp will look at the size of the file, allocate a buffer and get the file
// in one slurp. It turns out to be as slow as using the library Readfile().
func byOneSlurp(fname string) (int, error) {
	var fp *os.File
	var sz int64
	var err error
	if fp, err = os.Open(fname); err != nil {
		return 0, err
	} else {
		defer fp.Close()
		if fi, err := fp.Stat(); err != nil {
			return 0, err
		} else {
			sz = fi.Size()
		}
	}

	buf := make([]byte, sz)
	if _, err := fp.Read(buf); err != nil {
		return 0, err
	}

	return bytes.Count(buf, []byte(">")), nil
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
