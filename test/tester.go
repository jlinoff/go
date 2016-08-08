/*
The tester tool verifies that the go utils package are working as expected.
*/
package main

import (
	"io"
	"os"
	"strings"
	"github.com/jlinoff/go/msg"
	"github.com/jlinoff/go/run"
)

var log *msg.Object

func main() {
	testMsg()
	testRun()
	log.Info("success")
}

func testMsg() {
	log.Debug("debug message")
	log.Info("Info message pid = %v", os.Getpid())
	log.Warn("warning")
	log.ErrNoExit("this error is recoverable!")

	log.DebugEnabled = false
	log.Debug("this debug message will not display")

	log.DebugEnabled = true
	log.Debug("this debug message will display")

	log.InfoEnabled = false
	log.Info("this info message will not display")

	log.InfoEnabled = true
	log.Info("this info message will display")

	log.WarningEnabled = false
	log.Warn("this warning message will not display")

	log.WarningEnabled = true
	log.Warn("this warning message will display")

	log.Printf("any old random stuff\n")
}

func testRun() {
	log.Info("testing the run.Cmd() function")
	cmd := "./genout.sh 10 72"
	log.Info("cmd = %v", cmd)
	o, e := run.Cmd(strings.Fields(cmd))
	log.Info("size = %v", len(o))
	log.Info("err = %v", e)
	if e != nil {
		panic(e)
	}

	log.Info("testing the run.CmdSilent() function")
	cmd = "./genout.sh 10 72"
	log.Info("cmd = %v", cmd)
	o, e = run.CmdSilent(strings.Fields(cmd))
	log.Info("size = %v", len(o))
	log.Info("err = %v", e)
	if e != nil {
		panic(e)
	}
}

func init() {
	n := "Tester"
	f := `%pkg %(-27)time %(-7)type %file %line - %msg`
	t := `2006-01-02 15:05:05.000 MST`
	w := []io.Writer{os.Stdout}
	m, e := msg.NewMsg(n, f, t, w)
	if e != nil {
		panic(e)
	}
	log = m
}
