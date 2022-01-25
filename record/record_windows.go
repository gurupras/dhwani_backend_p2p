package record

func NewRecorder(identifier string, port int) Recorder {
	prog := "gst-launch-1.0"
	args := fmt.Sprintf("wasapisrc device=%v low-latency=true ! queue ! rawaudioparse ! audioresample ! opusenc frame-size=20 ! rtpopuspay ! udpsink host=127.0.0.1 port=%v", identifier, port)
	cmd := exec.Command(prog, strings.Split(args, " ")...)
	return &recorder{
		cmdline: fmt.Sprintf("%v %v", cmd, args),
		proc:    cmd,
	}
}
