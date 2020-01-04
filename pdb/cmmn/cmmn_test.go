package cmmn_test
import (
	. "github.com/andrew-torda/goutil/pdb/cmmn"
	"testing"
)
func TestXyzOk (t *testing.T) {
	var xyz Xyz
	xyz = BrokenXyz
	if xyz.Ok() {
		t.Error ("cannot even check if a value is OK")
	}
	xyz = Xyz{1,1,1}
	if !xyz.Ok() {
		t.Error ("OK should be true")
	}
}
