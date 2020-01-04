package pdb

import (
	"io/ioutil"
	"testing"
)

const nPDBsites = 3

func testSite(acq string, siteNum int, t *testing.T) {
	rdr, err := getHTTP(acq, siteNum)
	if err != nil {
		t.Fatal(err)
	}

	c, err := ioutil.ReadAll(rdr)
	defer rdr.Close()
	if len(c) < 100 || err != nil {
		t.Errorf("Reading from http got %v bytes, err = %v", len(c), err)
	}
}

func Test_get_http(t *testing.T) {
	for i := 0; i < nPDBsites; i++ {
		testSite ("5zck", i, t)
	}
}
