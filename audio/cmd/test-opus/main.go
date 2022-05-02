package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/alecthomas/kingpin"
	soundio "github.com/crow-misia/go-libsoundio"
	dhwaniAudio "github.com/gurupras/dhwani_backend_p2p/audio"
	log "github.com/sirupsen/logrus"
)

var (
	verbose          = kingpin.Flag("verbose", "Debug logs").Short('v').Bool()
	backend          = kingpin.Flag("backend", "soundio backend").Short('b').Required().String()
	deviceIdentifier = kingpin.Flag("device", "device identifier").Short('d').Required().String()
	frameSize        = kingpin.Flag("frame-size", "Opus frame size").Short('f').Default("20").Int()
	outfile          = kingpin.Flag("outfile", "output file").Short('o').Required().String()
)

func main() {
	kingpin.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	var soundioBackend soundio.Backend
	switch *backend {
	case "pulseaudio":
		soundioBackend = soundio.BackendPulseAudio
	default:
		soundioBackend = soundio.BackendNone
	}
	audio, err := dhwaniAudio.NewAudio(soundioBackend)
	if err != nil {
		log.Fatalf("Failed to create new Audio: %v\n", err)
	}

	stream, err := audio.StreamAudio(*deviceIdentifier, time.Duration(*frameSize)*time.Millisecond)
	if err != nil {
		log.Fatalf("Failed to stream audio: %v\n", err)
	}
	// opusStream, err := dhwaniAudio.EncodeOpus(stream)
	// if err != nil {
	// 	log.Fatalf("Failed to encode stream to opus: %v\n", err)
	// }

	file, err := os.Create(fmt.Sprintf("%v", *outfile))
	if err != nil {
		log.Fatalf("Failed to create output opus file: %v\n", err)
	}
	defer file.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for b := range stream.DataChan {
			_, err = file.Write(b)
			if err != nil {
				log.Fatalf("Failed to write to file: %v\n", err)
			}
		}
	}()
	stream.Start()
	wg.Wait()
}
