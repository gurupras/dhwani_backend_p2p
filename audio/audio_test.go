package audio

import (
	"sync"
	"testing"
	"time"

	soundio "github.com/crow-misia/go-libsoundio"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStreamAudio(t *testing.T) {
	// log.SetLevel(log.DebugLevel)
	require := require.New(t)
	audio, err := NewAudio(soundio.BackendPulseAudio)
	require.Nil(err)

	stream, err := audio.StreamAudio("alsa_output.pci-0000_0d_00.4.analog-stereo.monitor", 20*time.Millisecond)
	require.Nil(err)
	require.NotNil(stream)
	gotBytes := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for b := range stream.DataChan {
			require.Greater(len(b), 0)
			gotBytes = true
		}
	}()
	stream.Start()
	time.Sleep(300 * time.Millisecond)
	stream.Stop()
	wg.Wait()
	require.True(gotBytes)
}

func TestEncodeOpus(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	require := require.New(t)

	audio, err := NewAudio(soundio.BackendPulseAudio)
	require.Nil(err)

	stream, err := audio.StreamAudio("alsa_output.pci-0000_0d_00.4.analog-stereo.monitor", 20*time.Millisecond)
	require.Nil(err)
	require.NotNil(stream)

	opusStream, err := EncodeOpus(stream)
	require.Nil(err)

	gotBytes := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range opusStream.DataChan {
			gotBytes = true
		}
	}()
	stream.Start()
	time.Sleep(100 * time.Millisecond)
	stream.Stop()
	wg.Wait()
	require.True(gotBytes)
}
