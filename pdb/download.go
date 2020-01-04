// Package pdb covers reading PDB coordinates.
// Go to a pdb website and download coordinates.
// pdb europe files are at http://www.ebi.ac.uk/pdbe/entry-files/download/5pti.cif
// or maybe http://www.ebi.ac.uk/pdbe/entry-files/download/5pti_updated.cif
// The main point is to visit the web page and return a reader that
// can be used like the file readers.
// The interface should be changed so we can specify the server we like.
package pdb

import (
	"errors"
	"io"
	"net/http"
	"github.com/andrew-torda/goutil/pdb/zwrap"
)

// getHttp is given a four letter pdb code. It goes to the protein data
// bank and should return a reader.
// There are three sites for structures. You can pick which one you want with
// siteNum. If you give a value that it too big, we use a modulo to wrap
// it around, rather than generate an error. This makes it easier to cycle
// through them or pick one at random.
// Sites return normal or gzipped data, but if it is a gzipping site, we
// call zwrap to decompress and return that as the reader.
//func getHttp(acqCode string, siteNum int) (rslt pdbDownload, err error) {
func getHTTP(acqCode string, siteNum int) (io.ReadCloser, error) {
	var url string
	var urls = []struct {
		urlBase   string
		urlSuffix string
		gzipped   bool
	}{
		{"https://files.rcsb.org/download/",
			".cif.gz",
			true},
		{"http://www.ebi.ac.uk/pdbe/entry-files/download/",
			".cif",
			false},
		{"http://ftp.pdbj.org/mmcif/",
			".cif.gz",
			true},
	}

	if siteNum > len(urls) {
		siteNum = siteNum % len(urls)
	}

	if len(acqCode) != 4 {
		return nil, errors.New("acq code should be four char, not " + acqCode)
	}

	url = urls[siteNum].urlBase + acqCode + urls[siteNum].urlSuffix

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		s := "Wanted " + acqCode + " using " + url
		t := ", got " + resp.Status
		err = errors.New(s + t)
		return nil, err
	}

	if urls[siteNum].gzipped {
		if resp.Body, err = zwrap.Wrap(resp.Body) ; err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}
