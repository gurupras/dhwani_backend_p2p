package record

import (
	"bufio"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type recorder struct {
	cmdline string
	proc    *exec.Cmd
}

func (r *recorder) Start() error {
	stdout, _ := r.proc.StdoutPipe()
	stderr, _ := r.proc.StderrPipe()
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
	return r.proc.Start()
}

func (r *recorder) Stop() error {
	return r.proc.Process.Kill()
}

type Recorder interface {
	Start() error
	Stop() error
}
