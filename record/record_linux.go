package record

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gurupras/dhwani_backend_p2p/alsautils"
	log "github.com/sirupsen/logrus"
	"github.com/yobert/alsa"
)

type ALSARecorder struct {
	identifier string
	port       int
	cards      []*alsa.Card
	card       *alsa.Card
	device     *alsa.Device
	stopped    int32
	wg         sync.WaitGroup
}

func (r *ALSARecorder) Start() error {
	go r.Loop()
	return nil
}

func (r *ALSARecorder) Loop() error {
	if err := r.device.Open(); err != nil {
		return fmt.Errorf("failed to open device: %v", err)
	}

	channels, err := r.device.NegotiateChannels(2)
	if err != nil {
		return fmt.Errorf("failed to negotiate channels: %v", err)
	}
	rate, err := r.device.NegotiateRate(44100)
	if err != nil {
		return fmt.Errorf("failed to negotiate rate: %v", err)
	}
	format, err := r.device.NegotiateFormat(alsa.S16_LE)
	if err != nil {
		return fmt.Errorf("failed to negotiate format: %v", err)
	}
	bufSize, err := r.device.NegotiateBufferSize(8192, 16384)
	if err != nil {
		return fmt.Errorf("failed to negotiate buffer size: %v", err)
	}
	err = r.device.Prepare()
	if err != nil {
		return fmt.Errorf("failed to prepare device: %v", err)
	}
	buf := r.device.NewBufferDuration(10 * time.Millisecond)

	log.Debugf("identifier=%v channels=%v rate=%v format=%v bufSize=%v port=%v", r.identifier, channels, rate, format, bufSize, r.port)

	r.wg.Add(1)
	atomic.StoreInt32(&r.stopped, 0)
	defer r.wg.Done()

	conn, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%v", r.port))
	if err != nil {
		return fmt.Errorf("failed to open udp connection: %v", err)
	}
	defer conn.Close()

	for {
		if atomic.LoadInt32(&r.stopped) == 1 {
			break
		}
		err = r.device.Read(buf.Data)
		if err != nil {
			return fmt.Errorf("failed to read sound data: %v", err)
		}
		n, err := conn.Write(buf.Data)
		if err != nil {
			return fmt.Errorf("failed to write sound data to udp: %v", err)
		}
		_ = n
		log.Debugf("Wrote %v/%v bytes", n, len(buf.Data))
	}
	return nil
}

func (r *ALSARecorder) Stop() error {
	atomic.StoreInt32(&r.stopped, 1)
	r.wg.Wait()
	defer alsa.CloseCards(r.cards)
	defer r.device.Close()
	return nil
}

func NewRecorderGst(identifier string, port int) (Recorder, error) {
	prog := "gst-launch-1.0"
	args := fmt.Sprintf("alsasrc device=%v latency-time=1500 buffer-time=10000 ! queue ! rawaudioparse ! audioresample ! opusenc frame-size=20 ! rtpopuspay ! udpsink host=127.0.0.1 port=%v", identifier, port)
	cmd := exec.Command(prog, strings.Split(args, " ")...)
	return &subprocessRecorder{
		cmdline: fmt.Sprintf("%v %v", cmd, args),
		proc:    cmd,
	}, nil
}

func NewRecorder(identifier string, port int) (Recorder, error) {
	// Find the device that matches this identifier
	// TODO: Remove this
	_, _ = alsautils.ListDevicesWithLib()

	cards, err := alsa.OpenCards()
	if err != nil {
		return nil, err
	}
	// var card *alsa.Card
	// var device *alsa.Device

	// for _, c := range cards {
	// 	cardDevices, err := c.Devices()
	// 	if err != nil {
	// 		log.Errorf("Failed to list devices of card: %v\n", err)
	// 		continue
	// 	}
	// 	for _, d := range cardDevices {
	// 		id := alsautils.GetIdentifier(c, d)
	// 		if id == identifier && d.Record {
	// 			devInfo := make([]string, 0)
	// 			devInfo = append(devInfo, fmt.Sprintf("\tNumber: %v", d.Number))
	// 			devInfo = append(devInfo, fmt.Sprintf("\tPath:   %v", d.Path))
	// 			devInfo = append(devInfo, fmt.Sprintf("\tPlay:   %v", d.Play))
	// 			devInfo = append(devInfo, fmt.Sprintf("\tRecord: %v", d.Record))
	// 			devInfo = append(devInfo, fmt.Sprintf("\tTitle:  %v", d.Title))
	// 			devInfo = append(devInfo, fmt.Sprintf("\tType:   %v", d.Type))
	// 			log.Debugf("Using device: %v\n%v\n", d, strings.Join(devInfo, "\n"))

	// 			card = c
	// 			device = d
	// 			break
	// 		}
	// 	}
	// }

	return &ALSARecorder{
		identifier: identifier,
		port:       port,
		cards:      cards,
		card:       card,
		device:     device,
		stopped:    0,
		wg:         sync.WaitGroup{},
	}, nil
}
