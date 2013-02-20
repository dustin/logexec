package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"strings"
)

var stdoutLog, stderrLog *syslog.Writer

var facility = logFacility(syslog.LOG_LOCAL0)
var stdoutLevel = logLevel(syslog.LOG_INFO)
var stderrLevel = logLevel(syslog.LOG_WARNING)

var maxLogLine = flag.Int("maxline", 8*1024,
	"maximum amount of text to log in a line")

func init() {
	flag.Var(&facility, "facility", "logging facility")
	flag.Var(&stdoutLevel, "stdoutLevel", "log level for stdout")
	flag.Var(&stderrLevel, "stderrLevel", "log level for stderr")
}

var logErr = make(chan error)

func logPipe(w io.Writer, r io.Reader) {
	br := bufio.NewReader(r)
	for {
		l, err := br.ReadString('\n')
		if err != nil {
			logErr <- err
			return
		}

		l = strings.TrimSpace(l)
		if len(l) > *maxLogLine {
			l = l[:*maxLogLine]
		}

		_, err = w.Write([]byte(l))
		if err != nil {
			logErr <- err
			return
		}
	}
}

func main() {
	tag := flag.String("tag", "logexec", "Tag for all log messages")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("No command provided")
	}

	var err error
	lvl := syslog.Priority(stdoutLevel) | syslog.Priority(facility)
	stdoutLog, err = syslog.New(lvl, *tag)
	if err != nil {
		log.Fatalf("Error initializing stdout syslog: %v", err)
	}
	lvl = syslog.Priority(stderrLevel) | syslog.Priority(facility)
	stderrLog, err = syslog.New(lvl, *tag)
	if err != nil {
		log.Fatalf("Error initializing stderr syslog: %v", err)
	}

	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdin = os.Stdin
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error initializing stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Error initializing stderr pipe: %v", err)
	}

	go logPipe(stdoutLog, stdoutPipe)
	go logPipe(stderrLog, stderrPipe)

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error starting command: %v", err)
	}

	cmdChan := make(chan error)
	go func() {
		defer close(cmdChan)
		cmdChan <- cmd.Wait()
	}()

	for msgs := 0; msgs < 3; msgs++ {
		select {
		case err = <-cmdChan:
			cmdChan = nil
			if err != nil {
				fmt.Fprintf(stderrLog, "Command failed: %v", err)
				log.Fatalf("Command failed: %v", err)
			}
		case err = <-logErr:
			if err != nil && err != io.EOF {
				cmd.Process.Kill()
				fmt.Fprintf(stderrLog, "Error logging command output: %v", err)
				log.Fatalf("Error logging command output: %v", err)
			}
		}
	}
}
