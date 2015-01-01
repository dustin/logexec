package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var stdoutLog, stderrLog *syslog.Writer

var facility = logFacility(syslog.LOG_LOCAL0)
var stdoutLevel = logLevel(syslog.LOG_INFO)
var stderrLevel = logLevel(syslog.LOG_WARNING)
var tag string

var maxLogLine = flag.Int("maxline", 8*1024,
	"maximum amount of text to log in a line")

func init() {
	flag.Var(&facility, "facility", "logging facility")
	flag.Var(&stdoutLevel, "stdoutLevel", "log level for stdout")
	flag.Var(&stderrLevel, "stderrLevel", "log level for stderr")
	flag.StringVar(&tag, "tag", "logexec", "Tag for all log messages")

}

var logErr = make(chan error)

var sigs = make(chan os.Signal, 1)
var passSigs = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGTERM,
}

var wg sync.WaitGroup

func logPipe(w io.Writer, r io.Reader) {
	defer wg.Done()
	s := bufio.NewScanner(r)
	for s.Scan() {
		l := s.Bytes()

		l = bytes.TrimSpace(l)
		if len(l) > *maxLogLine {
			l = l[:*maxLogLine]
		}

		_, werr := w.Write(l)
		if werr != nil {
			logErr <- werr
			break
		}

	}
	if s.Err() != nil {
		logErr <- s.Err()
	}

}

func startCmd(cmdName string, args ...string) (*exec.Cmd, error) {
	var err error
	lvl := syslog.Priority(stdoutLevel) | syslog.Priority(facility)
	stdoutLog, err = syslog.New(lvl, tag)
	if err != nil {
		log.Fatalf("Error initializing stdout syslog: %v", err)
	}
	lvl = syslog.Priority(stderrLevel) | syslog.Priority(facility)
	stderrLog, err = syslog.New(lvl, tag)
	if err != nil {
		log.Fatalf("Error initializing stderr syslog: %v", err)
	}

	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = os.Stdin
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error initializing stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Error initializing stderr pipe: %v", err)
	}

	wg.Add(2)
	go logPipe(stdoutLog, stdoutPipe)
	go logPipe(stderrLog, stderrPipe)

	return cmd, cmd.Start()
}

func getExitStatus(err error) int {
	if err == nil {
		return 0
	}

	ose, ok := err.(*exec.ExitError)
	if !ok {
		// Unknown error type, default to 1
		return 1
	}

	estatus, ok := ose.Sys().(syscall.WaitStatus)
	if !ok {
		// Unknown sys type, default to 1
	}

	return estatus.ExitStatus()
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("No command provided")
	}

	signal.Notify(sigs, passSigs...)

	cmd, err := startCmd(flag.Arg(0), flag.Args()[1:]...)
	if err != nil {
		log.Fatalf("Error starting command: %v", err)
	}

	cmdChan := make(chan error)
	go func() {
		cmdChan <- cmd.Wait()
	}()

	// Signal with a channel when the loggers have completed
	doneChan := make(chan bool)
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	for !(cmdChan == nil && doneChan == nil) {
		select {
		case sig := <-sigs:
			log.Printf("logexec caught signal %v, passing through", sig)
			cmd.Process.Signal(sig)
		case <-doneChan:
			doneChan = nil
		case err = <-cmdChan:
			cmdChan = nil
			if estatus := getExitStatus(err); estatus != 0 {
				fmt.Fprintf(stderrLog, "Command return non-zero exit status: %v", estatus)
				os.Exit(estatus)
			}
		case err = <-logErr:
			if err != nil && err != io.EOF &&
				!strings.Contains(err.Error(), "bad file descriptor") {

				cmd.Process.Kill()
				fmt.Fprintf(stderrLog, "Error logging command output: %v", err)
				log.Fatalf("Error logging command output: %v", err)
			}
		}
	}
}
