package record

import (
	"bufio"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type subprocessRecorder struct {
	cmdline string
	proc    *exec.Cmd
}

func (s *subprocessRecorder) Start() error {
	stdout, _ := s.proc.StdoutPipe()
	stderr, _ := s.proc.StderrPipe()
	go func() {
		log.Debugf("Start stdout reader\n")
		defer func() {
			log.Debugf("Stopped stdout reader\n")
		}()
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()
	go func() {
		log.Debugf("Start stderr reader\n")
		defer func() {
			log.Debugf("Stopped stderr reader\n")
		}()
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()
	return s.proc.Start()
}

func (s *subprocessRecorder) Stop() error {
	return s.proc.Process.Kill()
}

type Recorder interface {
	Start() error
	Stop() error
}
