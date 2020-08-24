// 3 Aug 2020

package numseq

import (
	"bytes"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

// byMmap reads from a file pointer via mmap() and counts ">"
func ByMmap(fp *os.File) (int, error) {

	mm, err := mmap.Map(fp, mmap.RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer mm.Unmap()
	return bytes.Count(mm, []byte(">")), nil
}

// ByReading reads from a source, piecewise into a buffer and counts ">".
// It puts the file offset back to wherever it was at the start.
func ByReading(rdr io.ReadSeeker) (int, error) {
	const bsize = 64 * 1024
	var buf [bsize]byte
	iniOffset, err := rdr.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if _, err = rdr.Seek (0, io.SeekStart); err != nil {
		return 0, err
	}
	count := 0
	for n := bsize; n == bsize; {
		n, err = rdr.Read(buf[:])
		if err != nil && err != io.EOF {
			return 0, err
		}
		count += bytes.Count(buf[:n], []byte(">"))

	}
	_, err = rdr.Seek(iniOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// FromFile opens a file and returns the number of sequences by counting ">".
func FromFile(fname string) (int, error) {
	fp, err := os.Open(fname)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	return ByMmap(fp)
}
