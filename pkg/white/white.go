// I will put this somewhere else, when I am happy

package white

import (
	"fmt"
)

func isWhite(c byte) bool {
	var asciiSpace = [256]bool{
		'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}
	return asciiSpace[c]
}

// WhiteRemove acts on a byte slice, in place and removes all the white
// space. Check if we return it with the length adjusted, but the capacity
// unchanged.
type ByteSlice []byte

func (sIn ByteSlice) WhiteRemove() {
	fmt.Println("hello from whiteremove got", string(sIn))
	var s, t *byte
	s = &(sIn[0])
	t = s
	for i, s := range sIn {

	}
}

/* C version
no_blnk_num (char *s)
{
    size_t n = 0;
    char *t = s;
    while (*t) {
        while( isspace ((int)*t) || isdigit ((int)*t))
            t++;
        *s++ = *t;
        if (! *t++)
            break;
        n++;
    }
    return n;
}
*/
