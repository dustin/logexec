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

var logger *syslog.Writer

var maxLogLine = flag.Int("maxline", 8*1024,
	"maximum amount of text to log in a line")

var logErr = make(chan error)

func logPipe(logFun func(string) error, r io.Reader) {
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

		err = logFun(l)
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
	logger, err = syslog.New(syslog.LOG_ERR, *tag)
	if err != nil {
		log.Fatalf("Error initializing syslog: %v", err)
	}

	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdin = os.Stdin
	infoPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error initializing stdout pipe: %v", err)
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Error initializing stderr pipe: %v", err)
	}

	go logPipe(func(s string) error { return logger.Info(s) }, infoPipe)
	go logPipe(func(s string) error { return logger.Err(s) }, errPipe)

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
				logger.Err(fmt.Sprintf("Command failed: %v", err))
				log.Fatalf("Command failed: %v", err)
			}
		case err = <-logErr:
			if err != nil && err != io.EOF {
				cmd.Process.Kill()
				logger.Err(fmt.Sprintf("Error logging command output: %v", err))
				log.Fatalf("Error logging command output: %v", err)
			}
		}
	}
}
