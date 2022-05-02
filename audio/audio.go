package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	soundio "github.com/crow-misia/go-libsoundio"
	"github.com/glycerine/rbuf"
	log "github.com/sirupsen/logrus"
	"gopkg.in/hraban/opus.v2"
)

// App struct
type Audio struct {
	*soundio.SoundIo
	Stream *Stream
}

type AudioDevice struct {
	Name       string `json:"name"`
	Identifier string `json:"id"`
	CanPlay    bool   `json:"canPlay"`
	CanRecord  bool   `json:"canRecord"`
	Default    bool   `json:"default"`
}

type Stream struct {
	DataChan       chan []byte
	inStream       *soundio.InStream
	stopStreamChan chan struct{}
	wg             sync.WaitGroup
	Config         *StreamConfig
}

func (s *Stream) Start() error {
	return s.inStream.Start()
}

func (s *Stream) Stop() error {
	s.inStream.Pause(true)
	s.stopStreamChan <- struct{}{}
	s.wg.Wait()
	return nil
}

type StreamConfig struct {
	Format     soundio.Format
	SampleRate int
	Channels   int
}

func createSoundIoWithBackend(backend soundio.Backend) (*soundio.SoundIo, error) {
	opts := make([]soundio.Option, 0)
	opts = append(opts, soundio.WithBackend(backend))
	s := soundio.Create(opts...)
	if err := s.Connect(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewApp creates a new App application struct
func NewAudio(backend soundio.Backend) (*Audio, error) {
	audio := &Audio{}

	// // On Linux, we only try the PulseAudio backend
	// s, err := createSoundIoWithBackend(soundio.BackendPulseAudio)
	// if err != nil {
	// 	// Try None
	// 	s, err = createSoundIoWithBackend(soundio.BackendNone)
	// 	if err != nil {
	// 		log.Fatalf("Failed to create soundio backend: %v\n", err)
	// 	}
	// }
	s, err := createSoundIoWithBackend(backend)
	if err != nil {
		return nil, err
	}
	audio.SoundIo = s
	s.FlushEvents()
	return audio, nil
}

func (a *Audio) GetDevices() []*AudioDevice {
	outputCount := a.OutputDeviceCount()
	inputCount := a.InputDeviceCount()

	defaultOutput := a.DefaultOutputDeviceIndex()
	defaultInput := a.DefaultInputDeviceIndex()

	devices := make([]*AudioDevice, 0)

	for i := 0; i < inputCount; i++ {
		device := a.InputDevice(i)
		name := device.Name()
		id := device.ID()
		entry := &AudioDevice{
			Name:       name,
			Identifier: id,
			CanPlay:    false,
			CanRecord:  true,
			Default:    i == defaultInput,
		}
		devices = append(devices, entry)
		device.RemoveReference()
	}

	for i := 0; i < outputCount; i++ {
		device := a.OutputDevice(i)
		name := device.Name()
		id := device.ID()
		entry := &AudioDevice{
			Name:       name,
			Identifier: id,
			CanPlay:    true,
			CanRecord:  false,
			Default:    i == defaultOutput,
		}
		devices = append(devices, entry)
		device.RemoveReference()
	}
	return devices
}

var prioritizedFormats = []soundio.Format{
	soundio.FormatU16LE,
	soundio.FormatS16LE,
	soundio.FormatFloat32NE,
	soundio.FormatFloat32FE,
	soundio.FormatS32NE,
	soundio.FormatS32FE,
	soundio.FormatS24NE,
	soundio.FormatS24FE,
	soundio.FormatS16NE,
	soundio.FormatS16FE,
	soundio.FormatFloat64NE,
	soundio.FormatFloat64FE,
	soundio.FormatU32NE,
	soundio.FormatU32FE,
	soundio.FormatU24NE,
	soundio.FormatU24FE,
	soundio.FormatU16NE,
	soundio.FormatU16FE,
	soundio.FormatS8,
	soundio.FormatU8,
}

var prioritizedSampleRates = []int{
	48000,
	44100,
	96000,
	24000,
}

func (a *Audio) StreamAudio(deviceIdentifier string, bufferDuration time.Duration) (*Stream, error) {
	// First, we need to get the device
	count := a.InputDeviceCount()
	var selectedDevice *soundio.Device
	for i := 0; i < count; i++ {
		device := a.InputDevice(i)
		log.Debugf("[StreamAudio]: Checking %v == %v", device.ID(), deviceIdentifier)
		if device.ID() == deviceIdentifier {
			selectedDevice = device
			break
		}
		device.RemoveReference()
	}
	if selectedDevice == nil {
		return nil, fmt.Errorf("failed to find device: '%v'", deviceIdentifier)
	}

	selectedDevice.SortChannelLayouts()

	sampleRate := 0
	for _, rate := range prioritizedSampleRates {
		if selectedDevice.SupportsSampleRate(rate) {
			sampleRate = rate
			break
		}
	}
	if sampleRate == 0 {
		sampleRate = selectedDevice.SampleRates()[0].Max()
	}
	log.Printf("Sample rate: %d", sampleRate)

	format := soundio.FormatInvalid
	for _, f := range prioritizedFormats {
		if selectedDevice.SupportsFormat(f) {
			format = f
			break
		}
	}
	if format == soundio.FormatInvalid {
		format = selectedDevice.Formats()[0]
	}
	log.Printf("Format: %s", format)

	config := &soundio.InStreamConfig{
		Format:     format,
		SampleRate: sampleRate,
	}

	instream, err := selectedDevice.NewInStream(config)
	if err != nil {
		return nil, fmt.Errorf("unable to open input device: %s", err)
	}

	ret := &Stream{
		inStream:       instream,
		stopStreamChan: make(chan struct{}),
		DataChan:       make(chan []byte),
		wg:             sync.WaitGroup{},
		Config: &StreamConfig{
			Format:     format,
			SampleRate: sampleRate,
			Channels:   instream.Layout().ChannelCount(),
		},
	}
	go func() {
		defer selectedDevice.RemoveReference()
		defer close(ret.DataChan)
		defer instream.Destroy()

		var ringBuffer *rbuf.FixedSizeRingBuf

		channels := instream.Layout().ChannelCount()
		frameBytes := channels * instream.BytesPerFrame()

		overflowCount := 0
		stop := int32(0)

		instream.SetReadCallback(func(stream *soundio.InStream, frameCountMin int, frameCountMax int) {
			freeBytes := ringBuffer.N - ringBuffer.Readable
			freeCount := freeBytes / frameBytes
			writeFrames := freeCount
			if writeFrames > frameCountMax {
				writeFrames = frameCountMax
			}

			channelCount := stream.Layout().ChannelCount()
			frameLeft := writeFrames

			for {
				frameCount := frameLeft
				if frameCount <= 0 {
					break
				}

				areas, err := stream.BeginRead(&frameCount)
				if err != nil {
					log.Printf("begin read error: %s", err)
					ret.Stop()
					return
				}
				if frameCount <= 0 {
					break
				}
				if areas == nil {
					_, _ = ringBuffer.Write(make([]byte, frameCount*channelCount*stream.BytesPerFrame()))
				} else {
					for frame := 0; frame < frameCount; frame++ {
						for ch := 0; ch < channelCount; ch++ {
							buffer := areas.Buffer(ch, frame)
							n, err := ringBuffer.Write(buffer)
							_ = n
							if err != nil {
								log.Errorf("ringbuffer write error: %s, len %d", err, len(buffer))
							} else {
								// log.Debugf("Wrote %d bytes to ringbuffer", n)
							}
						}
					}
				}
				err = stream.EndRead()
				if err != nil {
					log.Printf("end read error: %s", err)
					ret.Stop()
					return
				}

				frameLeft -= frameCount
			}
		})
		instream.SetOverflowCallback(func(stream *soundio.InStream) {
			overflowCount++
			log.Printf("overflow %d", overflowCount)
		})

		bufSize := int(float64(instream.Layout().ChannelCount()) * 30 * float64(instream.SampleRate()) * float64(instream.BytesPerFrame()))
		capacity := bufSize * 5
		log.Debugf("bufSize=%v Capacity=%v\n", bufSize, capacity)
		ringBuffer = rbuf.NewFixedSizeRingBuf(capacity)

		for atomic.LoadInt32(&stop) == 0 {
			select {
			case <-ret.stopStreamChan:
				atomic.StoreInt32(&stop, 1)
				log.Debugf("Set stop variable")
			default:
				a.FlushEvents()
				time.Sleep(bufferDuration)
				buf := bytes.NewBuffer(nil)
				n, err := ringBuffer.WriteTo(buf)
				if err != nil {
					// log.Errorf("Failed to write ringbuffer -> tmp: %v\n", err)
				} else {
					log.Debugf("Wrote %d bytes from ringbuffer -> tmp\n", n)
					ret.DataChan <- buf.Bytes()
				}
			}
		}
		log.Debugf("Stopped loop")
	}()
	return ret, nil
}

func EncodeOpus(in *Stream) (*Stream, error) {
	ret := &Stream{
		DataChan:       make(chan []byte),
		inStream:       in.inStream,
		stopStreamChan: in.stopStreamChan,
		wg:             in.wg,
		Config:         in.Config,
	}

	byteArrayToIntArray := func(b []byte) []int16 {
		ret := make([]int16, len(b)/4)
		for idx := 0; idx < len(ret); idx++ {
			offset := idx * 4
			in := b[offset : offset+4]
			ret[idx] = int16(binary.LittleEndian.Uint16(in))
		}
		return ret
	}

	enc, err := opus.NewEncoder(ret.Config.SampleRate, ret.Config.Channels, opus.AppAudio)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ret.DataChan)
		for b := range in.DataChan {
			input := byteArrayToIntArray(b)
			frameSize := len(input)
			frameSizeMS := float32(frameSize / ret.Config.Channels * 1000 / ret.Config.SampleRate)
			switch frameSizeMS {
			case 2.5, 5, 10, 20, 40, 60:
			default:
				log.Errorf("Illegal frame size: %d bytes (%f ms)", frameSize, frameSizeMS)
				return
			}
			log.Debugf("[opus]: frameSize=%v frameSizeMS=%v", frameSize, frameSizeMS)
			data := make([]byte, 1000)
			n, err := enc.Encode(input, data)
			if err != nil {
				log.Errorf("Error while encoding opus: %v\n", err)
				continue
			}
			encoded := data[:n]
			ret.DataChan <- encoded
		}
	}()
	return ret, nil
}
