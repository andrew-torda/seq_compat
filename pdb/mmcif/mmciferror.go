// An error implementation that saves the line number and the
// line we were trying to read.
// The key is to call xxxx.fill() where xxxx is the name of the comment
// scanner/mmcif reader.
package mmcif

import (
	"strconv"
)

const maxMsgLen = 70

type readError struct {
	n      int    // line number
	inline string // The line that provoked the error
	desc   string // Description of error
}

// fill stores the problem we have seen for printing
// out when it is convenient. It is in the scanner, but
// can be seen (by inclusion) in the mmcifreader.
// If the error string is not nil, it means we neglected an error.
// Add this to the message.
func (m *cmmtScanner) fill(desc string, saveLine bool) {
	const multErrStr string = "\nNew error, but there was already an error from line "
	if m.Ok == false {
		ln := strconv.FormatInt(int64(m.n), 10) // line num
		desc = m.l_err.desc + multErrStr + ln + ":\n" + desc + "\n"
	}
	m.Ok = false
	if saveLine {
		m.l_err.n = m.n
	}
	m.l_err.inline = string(m.cbytes()) // Saves current line in scanner m
	m.l_err.desc = desc
}

func firstPart(s string) string {
	l := len(s)
	if l > maxMsgLen {
		l = maxMsgLen
	}
	return s[:l]
}

// Error takes what is known about the state and causes and returns a
// single string. This should include the number of the last line read
// and any description of the error we have.
// We assume you have used fill() to fill out the information when an
// error was seen.
func (e readError) Error() string {
	var errmsg string
	if e.n != 0 {
		errmsg = "Line: " + strconv.FormatInt(int64(e.n), 10) + " "
	}
	errmsg += e.desc
	if e.n != 0 {
		errmsg += "\nLine starting with\n" + firstPart(e.inline)
	}
	return errmsg
}
