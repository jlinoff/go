/*
Package msg creates a simple messaging package with flexible formatting that
can be used for stdout and/or logging.

Here is an example use.

    import (
      "jlinoff/utils/msg"
      "io"
      "os"
    )

    // My package logger.
    var log *msg.Object

    // Initialize it at startup.
    func init() {
      // Only write to stdout.
      w := []io.Writer{os.Stdout}

      // The name of my package. It is only used %pkg is specified in the
      // format string.
      n := "MyPackage"

      // Format string. Note that i could use %utc instead of %time to get
      // UTC time.
      f := `%pkg %(-27)time %(-7)type %file %line - %msg`

      // Time format string, only used if %time or %utc are specified in the
      // the format string.
      t := `2006-01-02 15:05:05.000 MST`

      // Create the message object.
      // Note that this is the same as this because I used the defaults.
      //     msg.NewMsg("MyPackage", "", "", []io.Writer{})
      l, e := msg.NewMsg(n, f, t, w)
      if e != nil { panic(e) }
      log = l
    }

    func test() {
      log.Debug("message of type %v", "debug")
      log.Info("info message")
      log.Warn("warning")

      // Now print messages to stdout and to a log while in this scope.
      fp, _ := os.Create("log.txt")
      log.Writers = append(log.Writers, fp)

      // This stuff will go to stdout and the log file.
      log.Info("both")
      log.ErrNoExit("bad stuff happened but i can recover!")

      // Clean up by removing the file from the writers and then
      // closing it.
      log.Writers = log.Writers[:len(log.Writers)-1]
      fp.Close()
  }
*/
package msg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

// Interface defines logger functions.
type Interface interface {
	Debug(f string, a ...interface{})
	Info(f string, a ...interface{})
	Warn(f string, a ...interface{})
	Err(f string, a ...interface{})
	ErrNoExit(f string, a ...interface{})

	DebugWithLevel(l int, f string, a ...interface{})
	InfoWithLevel(l int, f string, a ...interface{})
	WarnWithLevel(l int, f string, a ...interface{})
	ErrWithLevel(l int, f string, a ...interface{})
	ErrNoExitWithLevel(l int, f string, a ...interface{})
}

// Object defines the logger.
type Object struct {
	// Name is the package name. It is accessed in the format string by %pkg.
	Name string

	// DebugEnabled enables debug messages if true.
	// It is true by default.
	DebugEnabled bool

	// InfoEnabled enables info messages if true.
	// It is true by default.
	InfoEnabled bool

	// WarningEnabled enables warning messages if true.
	// It is true by default.
	WarningEnabled bool

	// ErrorExitCode is the exit code to use for the Error function.
	// The default is 1.
	ErrorExitCode int

	// Writers for the message output.
	// If no writers are specified, messages go to os.Stdout.
	Writers []io.Writer

	// TimeFormat is the format of the prefix timestamp.
	// See time.Format for details.
	// The default format is: "2006-01-02 15:05:05.000 MST"
	TimeFormat string

	// Format is the template for the output. It has the following specifiers.
	//
	//   %file is the caller file name
	//   %func is the function name
	//   %line is the line number
	//   %msg  is the actual message
	//   %pkg  is the package name
	//   %time is the time format in the current locale
	//   %utc is the time format in the UTC locale
	//   %type is the msg type: DEBUG, INFO, WARNING, ERROR
	//   %% is a single % character
	//
	// You can explicitly format each field by specifying the formatting
	// options in parentheses.
	//
	//   %(-28)time
	//
	// Any other text is left verbatim.
	//
	// The default format is.
	//   `%(-27)time %(-7)type %file %line - %msg`
	Format string

	// outputFormat created by NewMsgObject and used to generate a message.
	outputFormat string

	// outputFlds created by NewMsgObject and used to specify the fields.
	outputFlds []string
}

// NewMsg makes a message object.
//   n - package name
//   f - format string, set to "" to get the default.
//   t - time stamp format, set to "" to get the default
//   w - the list of writers, if empty all messages go to stdout
func NewMsg(n string, f string, t string, w []io.Writer) (obj *Object, err error) {
	obj = new(Object)
	obj.Name = n
	obj.DebugEnabled = true
	obj.InfoEnabled = true
	obj.WarningEnabled = true
	obj.ErrorExitCode = 1

	if len(w) == 0 {
		obj.Writers = append(obj.Writers, os.Stdout)
	} else {
		obj.Writers = w
	}

	// Set the time format. If it is empty, set the default.
	if t == "" {
		obj.TimeFormat = "2006-01-02 15:05:05.000 MST"
	} else {
		obj.TimeFormat = t
	}

	// Set the format. If it is empty use the default.
	if f == "" {
		obj.Format = `%(-27)time %(-7)type %file %line - %msg`
	} else {
		obj.Format = f
	}

	// Parse the format.
	ofmt, oflds, err := ParseFormatString(f)
	obj.outputFormat = ofmt
	obj.outputFlds = oflds

	return
}

/*
Debug prints a debug message obtaining the callers filename, function and
line number.

Example:
      msg.Debug("%v = %v", key, value)
*/
func (o Object) Debug(f string, a ...interface{}) {
	if o.DebugEnabled {
		o.Print("DEBUG", 2, f, a...)
	}
}

/*
DebugWithLevel prints a debug message obtaining the filename, function and
line number from the caller specified by level "l". l=2 is the same
as Debug().

Example:
      msg.DebugWithLevel(2, "%v = %v", key, value)
*/
func (o Object) DebugWithLevel(l int, f string, a ...interface{}) {
	if o.DebugEnabled {
		o.Print("DEBUG", l, f, a...)
	}
}

/*
Info prints an info message obtaining the callers filename, function and
line number.

Example:
      msg.Info("%v = %v", key, value)
*/
func (o Object) Info(f string, a ...interface{}) {
	if o.InfoEnabled {
		o.Print("INFO", 2, f, a...)
	}
}

/*
InfoWithLevel prints an info message obtaining the filename, function and
line number from the caller specified by level "l". l=2 is the same
as Debug().

Example:
      msg.InfoWithLevel(2, "%v = %v", key, value)
*/
func (o Object) InfoWithLevel(l int, f string, a ...interface{}) {
	if o.InfoEnabled {
		o.Print("INFO", l, f, a...)
	}
}

/*
Warn prints a warning message obtaining the callers filename, function and
line number.

Example:
      msg.Warn("%v = %v", key, value)
*/
func (o Object) Warn(f string, a ...interface{}) {
	if o.WarningEnabled {
		o.Print("WARNING", 2, f, a...)
	}
}

/*
WarnWithLevel prints a warning message obtaining the filename, function and
line number from the caller specified by level "l". l=2 is the same
as Debug().

Example:
      msg.WarnWithLevel(2, "%v = %v", key, value)
*/
func (o Object) WarnWithLevel(l int, f string, a ...interface{}) {
	if o.WarningEnabled {
		o.Print("WARNING", 2, f, a...)
	}
}

/*
Err prints an error message obtaining the callers filename, function and
line number and exits. It cannot be disabled.

Example:
      msg.Err("%v = %v", key, value)
*/
func (o Object) Err(f string, a ...interface{}) {
	o.Print("ERROR", 2, f, a...)
	os.Exit(o.ErrorExitCode)
}

/*
ErrWithLevel prints an error message obtaining the filename, function and
line number from the caller specified by level "l". l=2 is the same
as Debug() and exits. It cannot be disabled.

Example:
      msg.ErrWithLevel(2, "%v = %v", key, value)
*/
func (o Object) ErrWithLevel(l int, f string, a ...interface{}) {
	o.Print("ERROR", 2, f, a...)
	os.Exit(o.ErrorExitCode)
}

/*
ErrNoExit prints an error message obtaining the callers filename, function and
line number. It does not exit and cannot be disabled.

Example:
      msg.ErrNoExit("%v = %v", key, value)
*/
func (o Object) ErrNoExit(f string, a ...interface{}) {
	o.Print("ERROR", 2, f, a...)
}

/*
ErrNoExitWithLevel prints an error message obtaining the filename, function and
line number from the caller specified by level "l". l=2 is the same
as Debug(). It does not exit and cannot be disabled.

Example:
      msg.ErrNoExitWithLevel(2, "%v = %v", key, value)
*/
func (o Object) ErrNoExitWithLevel(l int, f string, a ...interface{}) {
	o.Print("ERROR", 2, f, a...)
}

/*
Print is the basis of all message printers. It normally would not be called
directly.

      t - is the type, normally one of DEBUG, INFO, WARNING or ERROR
      l - is the caller level: 0 is this function, 1 is the caller, 2 is the callers caller and so on
      f - format string
      a - argument list
*/
func (o Object) Print(t string, l int, f string, a ...interface{}) {
	pc, fname, lineno, _ := runtime.Caller(l)
	fct := runtime.FuncForPC(pc).Name()
	fname = path.Base(fname[0 : len(fname)-3]) // strip off ".go"

	// The variables map for the format string.
	m := map[string]string{
		"file": fname,
		"func": fct,
		"line": strconv.Itoa(lineno),
		"msg":  fmt.Sprintf(f, a...),
		"pkg":  o.Name,
		"time": time.Now().Truncate(time.Millisecond).Format(o.TimeFormat),
		"utc":  time.Now().UTC().Truncate(time.Millisecond).Format(o.TimeFormat),
		"type": t,
	}

	// Collect the field values.
	var flds []interface{}
	for _, k := range o.outputFlds {
		if v, ok := m[k]; ok {
			flds = append(flds, v)
		} else {
			// This is, essentially, an assert. It should never happen.
			fmt.Fprintf(os.Stderr, "ERROR: unexpected condition, invalid specification id '%v'\n", k)
			os.Exit(1)
		}
	}

	// Create the formatted output string.
	s := fmt.Sprintf(o.outputFormat, flds...) + "\n"

	// Output it for each writer.
	for _, w := range o.Writers {
		fmt.Fprintf(w, s)
	}
}

/*
ParseFormatString transforms a format template to a format string
and the list of fields to print in each message.

It is meant to be used internally by NewMsg().

Here is an example transformation:

      input = "MYSTUFF %(-27)time %(-7)type %file %line - %msg"

      // TRANSFORM
      ofmt  = "MYSTUFF %-27v %-7v %v %v - %v"
      oids  = ["time", "type", "type", "file", "line", "msg"]
*/
func ParseFormatString(input string) (ofmt string, oids []string, err error) {
	ofmtb := []byte{}
	valid := []string{"file", "func", "line", "msg", "pkg", "time", "type", "utc"}
	ics := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-$")

	// Define the parse states.
	//    normal - capture each byte
	//    spec   - capture a specification of the form %<id> or %(<fmt>)<id>.
	type state int
	const (
		normal state = iota
		spec
	)
	s := normal
	ib := []byte(input)
	for i := 0; i < len(ib); i++ {
		b := ib[i]
		switch s {
		case normal:
			// normal state, this is all of the stuff in the
			// template that is not part of a specification.
			if b == '%' {
				s = spec
			} else {
				ofmtb = append(ofmtb, b)
			}
		case spec:
			s = normal // after parsing the spec go back to the normal state

			// specification state, parse specifications of the form:
			//   %(<fmt>)<id>
			//   %<id>
			beg := i - 1
			if b == '(' {
				// This is a format specification. Capture it.
				// If ')' is not found, report an error.
				j := i // ib[j] == '('
				for ; i < len(ib) && ib[i] != ')'; i++ {
				}
				if i >= len(ib) {
					err = fmt.Errorf("missing ')' for '%v'", string(ib[beg:]))
					return
				}
				ofmtb = append(ofmtb, '%')
				ofmtb = append(ofmtb, ib[j+1:i]...)
				ofmtb = append(ofmtb, 'v')
				i++ // point past the ')'
			} else {
				ofmtb = append(ofmtb, []byte("%v")[:]...)
			}

			// Now parse out the id.
			id := ""
			for _, v := range valid {
				ba := []byte(v)
				if bytes.HasPrefix(ib[i:], ba) {
					// We MAY have a match.
					// for example '%line' matches but '%linex' does not.
					i += len(ba)
					if i < len(ib) {
						bs := []byte{ib[i]}
						if bytes.Contains([]byte(ics), bs) {
							ofmt = string(ofmtb)
							ba = append(ba, ib[i])
							err = fmt.Errorf("unrecognized specification id '%v'", string(ba))
							return
						}
					}
					id = string(ba)
					i--
					break
				}
			}
			if id == "" {
				ofmt = string(ofmtb)
				err = fmt.Errorf("specification syntax error '%v'", string(ib[beg:]))
				return
			}
			oids = append(oids, id)
		}
	}
	ofmt = string(ofmtb)
	return
}
