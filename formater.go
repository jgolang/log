package log

import (
	"os"
	"strings"
	"time"
)

// formater doc ...
type formater interface {
	Development(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte)
	Production(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte)
}

// These flags define which text to prefix to each log entry generated by the Logger.
// Bits are or'ed together to control what's printed.
// There is no control over the order they appear (the order listed
// here) or the format they present (as described in the comments).
// The prefix is followed by a colon only when Llongfile or Lshortfile
// is specified.
// For example, flags Ldate | Ltime (or LstdFlags) produce,
//	2009/01/23 01:23:23 message
// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
const (
	Ldate          = 1 << iota                                                                                // the date in the local time zone: 2009-01-23
	Ltime                                                                                                     // the time in the local time zone: 01:23:23
	Lmicroseconds                                                                                             // microsecond resolution: 01:23:23.123123.  assumes Ltime. * available in mode prod
	Llevel                                                                                                    // the level message information
	Lcaller                                                                                                   // full file name and line number: /a/b/c/d.go:23 * available in mode prod
	Lfile                                                                                                     // file name: package/d.go * available in mode prod
	Lline                                                                                                     // the line file: package/d.go:23. assumes Lfile. * available in mode prod
	Lfunc                                                                                                     // func name: (function) * available in mode prod
	Lstack                                                                                                    // stack trace
	LUTC                                                                                                      // if Ldate or Ltime is set, use UTC rather than the local time zone
	Linfo                                                                                                     // add flags info to mode production JSON format * available in mode prod
	JSONFormat                                                                                                // activate JSON Format output mode
	TerminalFormat                                                                                            // activate Terminal Format output mode
	LstdDevFlags   = Ldate | Ltime | Lmicroseconds | Llevel | Lfile | Lline | Lfunc | Lstack | TerminalFormat // initial values for the standard development logger
	LstdProdFlags  = Ldate | Ltime | Lmicroseconds | Lcaller | JSONFormat                                     // initial values for the standard production logger
)

// Formater doc ...
type Formater struct {
	devFlag        int
	prodFlag       int
	additionalInfo string
}

// Production doc
func (f Formater) Production(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte) {
	if p < 2 {
		return
	}
	if f.prodFlag&(TerminalFormat) != 0 {
		f.terminalFormat(buf, t, file, line, function, p, s, stack)
		return
	}
	f.jsonFormat(buf, t, file, line, function, p, s, stack)
}

// Development writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (f Formater) Development(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte) {
	if f.devFlag&(JSONFormat) != 0 {
		f.jsonFormat(buf, t, file, line, function, p, s, stack)
		return
	}
	f.terminalFormat(buf, t, file, line, function, p, s, stack)
}

func itoa(buf *[]byte, i int64, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (f Formater) jsonFormat(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte) {
	*buf = append(*buf, '{')
	*buf = append(*buf, "\"level\":\""...)
	*buf = append(*buf, getTypeMsg(true, p)...)
	*buf = append(*buf, "\",\"ts\":\""...)
	ts := t.Unix()
	itoa(buf, ts, 10)
	if f.prodFlag&(Lmicroseconds) != 0 {
		*buf = append(*buf, '.')
		itoa(buf, int64(t.Nanosecond()/1e3), 6)
	}
	if f.prodFlag&(Linfo) != 0 {
		*buf = append(*buf, "\",\"flags\":\""...)
		*buf = append(*buf, f.additionalInfo...)
	}
	if f.prodFlag&(Lcaller) != 0 {
		pk := getLastToFirstStrSlice(function, '/', 0)
		fl := getLastStrSlice(file, '/', 1)
		*buf = append(*buf, "\",\"caller\":\""...)
		*buf = append(*buf, pk...)
		*buf = append(*buf, fl...)

		// *buf = append(*buf, pk...)
		// *buf = append(*buf, '/')
		// *buf = append(*buf, fileName...)
		// *buf = append(*buf, ".go"...)
		*buf = append(*buf, ':')
		itoa(buf, int64(line), 2)
	}
	if f.prodFlag&(Lfile) != 0 {
		fileName := getLastStrSlice(file, os.PathSeparator, 0)
		*buf = append(*buf, "\",\"file\":\""...)
		*buf = append(*buf, fileName...)
	}
	if f.prodFlag&(Lline) != 0 {
		*buf = append(*buf, "\",\"line\":\""...)
		itoa(buf, int64(line), 2)
	}
	if f.prodFlag&(Lfunc) != 0 {
		funcName := getLastStrSlice(function, '.', 0)
		*buf = append(*buf, "\",\"func\":\""...)
		*buf = append(*buf, funcName...)
	}
	*buf = append(*buf, "\",\"msg\":\""...)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	*buf = append(*buf, s...)
	if stack != nil && f.prodFlag&(Lstack) != 0 {
		*buf = append(*buf, "\",\"stack\":["...)
		len := len(stack)
		for i, trace := range stack {
			*buf = append(*buf, "{\"trace\":\""...)
			*buf = append(*buf, trace...)
			*buf = append(*buf, "\"}"...)
			if len > i+1 {
				*buf = append(*buf, ',')
			}
		}
		*buf = append(*buf, ']')
		*buf = append(*buf, "}"...)
	} else {
		*buf = append(*buf, "\"}"...)
	}
}

func (f Formater) terminalFormat(buf *[]byte, t time.Time, file string, line int, function string, p priority, s string, stack [][]byte) {
	if f.devFlag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if f.devFlag&LUTC != 0 {
			t = t.UTC()
		}
		if f.devFlag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, int64(year), 4)
			*buf = append(*buf, '-')
			itoa(buf, int64(month), 2)
			*buf = append(*buf, '-')
			itoa(buf, int64(day), 2)
		}
		if f.devFlag&Ltime != 0 {
			*buf = append(*buf, 'T')
			hour, min, sec := t.Clock()
			itoa(buf, int64(hour), 2)
			*buf = append(*buf, ':')
			itoa(buf, int64(min), 2)
			*buf = append(*buf, ':')
			itoa(buf, int64(sec), 2)
		}
		if f.devFlag&Lmicroseconds != 0 {
			*buf = append(*buf, '.')
			itoa(buf, int64(t.Nanosecond()/1e3), 6)
		}
		*buf = append(*buf, '\t')
	}
	if f.devFlag&(Llevel) != 0 {
		*buf = append(*buf, getTypeMsg(false, p)...)
		*buf = append(*buf, '\t')
	}
	if f.devFlag&(Lcaller) != 0 {
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, int64(line), 2)
	} else if f.devFlag&(Lfile) != 0 {
		pk := getLastToFirstStrSlice(function, '/', 0)
		fl := getLastStrSlice(file, '/', 1)
		*buf = append(*buf, pk...)
		*buf = append(*buf, fl...)
		if f.devFlag&(Lline) != 0 {
			*buf = append(*buf, ':')
			itoa(buf, int64(line), 2)
		}
	}
	if f.devFlag&(Lfunc) != 0 {
		function := getLastStrSlice(function, '.', 0)
		*buf = append(*buf, ' ')
		*buf = append(*buf, '(')
		*buf = append(*buf, function...)
		*buf = append(*buf, ')')
	}
	*buf = append(*buf, '\t')
	*buf = append(*buf, s...)
	if stack != nil && f.devFlag&(Lstack) != 0 {
		*buf = append(*buf, "\n--- TRACE:"...)
		for _, trace := range stack {
			*buf = append(*buf, '\n')
			*buf = append(*buf, '\t')
			*buf = append(*buf, trace...)
		}
		*buf = append(*buf, "\n---"...)
	}
}

// NewFormaterConfig doc ...
func NewFormaterConfig(devFlag, prodFlag int, additionalInfo string) Formater {
	return Formater{devFlag: devFlag, prodFlag: prodFlag, additionalInfo: additionalInfo}
}
