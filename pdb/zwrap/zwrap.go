// Package zwrap takes a file pointer and optionally wraps it so upon
// calling Close, the decompressor will be closed, followed by the
// underlying file.
// I benchmarked with and without buffering in Wrap(). I could not measure
// any difference.

package zwrap

import (
	"compress/gzip"
	"errors"
	"io"
)

type FpGzip struct { // This is what we return.
	fp   io.ReadCloser
	zrdr *gzip.Reader
}

// Close closes the compressor, then the underlying backing readCloser.
// It should work if the source is a file or an http stream.
func (fc *FpGzip) Close() error {
	if fc.zrdr == nil {
		return fc.fp.Close()
	}
	var s string
	if e := fc.zrdr.Close(); e != nil { // Close decompressor
		s = e.Error()
	}
	if e := fc.fp.Close(); e != nil { // and backing file
		s = s + " " + e.Error()
	}
	if s == "" {
		return nil
	}
	return errors.New(s)
}

// Read makes sure we read from the compressed stream and
// not the underlying file stream.
func (fc *FpGzip) Read(p []byte) (int, error) {
	if fc.zrdr != nil {
		return fc.zrdr.Read(p)
	}
	return fc.fp.Read(p)
}

// Wrap takes a source like a file pointer or http stream and wraps it
// so the correct Close and Read will be called. Although we use the
// name fp, it should be happy if it is fed an http stream.
func Wrap(fp io.ReadCloser) (*FpGzip, error) {
	var fpz FpGzip
	var err error
	fpz.fp = fp
	fpz.zrdr, err = gzip.NewReader(fpz.fp) // No need to check error.
	return &fpz, err                       // Just pass it back
}

// ReadSeekCloser does not seem to be in the standard library
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// WrapMaybe will decide if the underlying stream is compressed
// and wrap the file pointer if necessary.
// You do lose something. If you pass in something which can seek,
// you get back a ReadCloser which cannot seek. This is the price
// one pays for reading from a compressed reader.
func WrapMaybe(fpIn ReadSeekCloser) (*FpGzip, error) {
	if out, err := Wrap(fpIn); err == nil {
		return out, nil // It was compressed. Return compressed reader.
	}
	_, err := fpIn.Seek(0, io.SeekStart)
	r := &FpGzip{
		fp: fpIn, // Leave the zrdr implicitly nil
	}

	return r, err
}
