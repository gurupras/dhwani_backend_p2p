package record

import (
	"fmt"
	"os/exec"
	"strings"
)

func NewRecorder(identifier string, port int) Recorder {
	prog := "gst-launch-1.0"
	args := fmt.Sprintf("alsasrc device=%v latency-time=1500 buffer-time=10000 ! queue ! rawaudioparse ! audioresample ! opusenc frame-size=20 ! rtpopuspay ! udpsink host=127.0.0.1 port=%v", identifier, port)
	cmd := exec.Command(prog, strings.Split(args, " ")...)
	return &recorder{
		cmdline: fmt.Sprintf("%v %v", cmd, args),
		proc:    cmd,
	}
}
