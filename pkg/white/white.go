// I will put this somewhere else, when I am happy

package white

func isWhite(c byte) bool {
	var asciiSpace = [256]bool{
		'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}
	return asciiSpace[c]
}

// WhiteRemove acts on a byte slice, in place and removes all the white
// space. Note, this change len().
type ByteSlice []byte

func (sIn *ByteSlice) WhiteRemove() {
	s := *sIn
	i, j := 0, 0
	for ; j < len(s); i, j = i+1, j+1{
		for ; j < len(s); {
			if isWhite(s[j]) {
				j++
			} else {
				break
			}
		}
		if j >= len(s) {
			break
		}
		s[i] = s[j]
	}
	const fill_in_with_nulls = false
	if fill_in_with_nulls {
		for n:= i; n < len(s); n++ {
			s[n] = 0
		}
	}
	*sIn = s[:i]
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
