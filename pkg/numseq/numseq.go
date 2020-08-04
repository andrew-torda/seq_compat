// 3 Aug 2020

package numseq

import (
	"bytes"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

// byMmap reads from a file via mmap() and counts ">"
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

//byReadingFixed reads from a file, piecewise into a buffer and counts ">".
func byReadingFixed(fname string) (int, error) {
	const bsize = 64 * 1024
	var buf [bsize]byte
 	var fp io.ReadCloser
	var err error
	if fp, err = os.Open(fname); err != nil {
		return 0, err
	}
	defer fp.Close()
	count := 0
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
	return byMmap (fname)
}
