/*
Packge run contains a few exec command wrappers and a function to get the exit code.

Here is an example usage.

    // run a command
    // usage: run [-s] command
    // examples:
    //     $ go run test.go ls -l
    //     $ go run test.go -s ls -l
    package main

    import (
        "fmt"
        "os"
        "github.com/jlinoff/go/run"
    )

    func main() {
        if len(os.Args) > 1 {
            var cmdargs []string
            var out string
            var err error
            if os.Args[1] == "-s" {
                cmdargs = os.Args[2:]
                out, err = run.CmdSilent(cmdargs)
            } else {
                cmdargs = os.Args[1:0]
                out, err = run.Cmd(cmdargs)
            }
            fmt.Printf("size = %v\n", len(out))
            fmt.Printf("err  = %v\n", err)
          }
        }
    }
*/
package run

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

/*
Cmd runs a command.

The output is displayed and the output is returned.

Example:
      // Run the command, display the output.
      if _, e := run.Cmd(strings.Fields("ls -l")); e != nil { panic(e) }
*/
func Cmd(a []string) (output string, err error) {
	if len(a) == 0 {
		err = fmt.Errorf("no command specified")
		return
	}

	// Create the command object.
	c := exec.Command(a[0], a[1:]...)

	// Write stdout and stderr to a buffer and to os.Stdout.
	var buf bytes.Buffer
	writers := []io.Writer{&buf, os.Stdout}
	w := io.MultiWriter(writers...)
	c.Stdout = w
	c.Stderr = w
	c.Stdin = os.Stdin

	// Run the command.
	err = c.Run()
	output = buf.String()
	return
}

/*
CmdWithWriters runs a command with customer sinks.

The caller decides what to display and/or receive.

Example:
      // Run the command, display the output to stdout, to a file and capture
      // it in a buffer.
      var buf bytes.Buffer

      fp, _ := os.Create("/tmp/log")
      defer fp.Close()

      w := io.Writer[]{os.Stdout, fp, &buf}
      if _, e := run.Cmd(strings.Fields("ls -l")); e != nil { panic(e) }
      fmt.Println(buf.String())  // print the buffer
*/
func CmdWithWriters(a []string, w []io.Writer) (err error) {
	if len(a) == 0 {
		err = fmt.Errorf("no command specified")
		return
	}

	// Create the command object.
	c := exec.Command(a[0], a[1:]...)

	// Write stdout and stderr to a buffer and to os.Stdout.
	m := io.MultiWriter(w...)
	c.Stdout = m
	c.Stderr = m
	c.Stdin = os.Stdin

	// Run the command.
	err = c.Run()
	return
}

/*
CmdSilent runs a command silently so that the output can be parsed.

The output is displayed and the output is returned.

Example:
      // Run the command, do not display the output.
      o, e := run.CmdSilent(strings.Fields("ls -l"))
      if e != nil { panic(e) }
      fmt.Printf("%v", o)
*/
func CmdSilent(a []string) (output string, err error) {
	if len(a) == 0 {
		err = fmt.Errorf("no command specified")
		return
	}

	// Create the command object.
	c := exec.Command(a[0], a[1:]...)

	// Run the command silently.
	out, e := c.CombinedOutput()
	err = e
	output = string(out)

	return
}

/*
GetExitCode gets the exit code of the last exec call, if possible.

CITATION: http://stackoverflow.com/questions/10385551/get-exit-code-go

Here is how you might use it.
      _, e := run.Cmd("ls -l")
      if e != nil {
        fmt.Println(e)
        code = GetExitCode(e)
        fmt.Printf("exit code %v", code)
      }
*/
func GetExitCode(err error) (code int) {
	code = 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				code = status.ExitStatus()
			} else {
				code = -1
			}
		}
	}
	return
}
