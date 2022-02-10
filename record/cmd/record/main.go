package main

import (
	"log"
	"sync"

	"github.com/gurupras/dhwani_backend_p2p/record"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	device  = kingpin.Arg("device", "Name of alsa device. e.g. hw:1,0").Required().String()
	port    = kingpin.Arg("port", "UDP port. default 4444").Default("4444").Int()
)

func main() {
	kingpin.Parse()
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	r, err := record.NewRecorder(*device, *port)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	r.Start()
	wg.Wait()
}
