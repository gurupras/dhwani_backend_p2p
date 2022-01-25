package audio

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
	log "github.com/sirupsen/logrus"
)

type AudioRTP struct {
	Port     int
	listener *net.UDPConn
	Track    *webrtc.TrackLocalStaticSample
	mutex    sync.Mutex
	wg       sync.WaitGroup
	running  bool
	stopped  bool
}

func (artp *AudioRTP) Loop() {
	artp.mutex.Lock()
	artp.running = true
	artp.stopped = false
	artp.mutex.Unlock()
	inboundRTPPacket := make([]byte, 1600) // UDP MTU
	artp.wg.Add(1)
	defer func() {
		artp.mutex.Lock()
		defer artp.mutex.Unlock()
		artp.stopped = true
		artp.wg.Done()
	}()

	defer func() {
		artp.listener.Close()
	}()

	once := false
	audioBuilder := samplebuilder.New(3, &codecs.OpusPacket{}, 48000)

	for {
		packet := &rtp.Packet{}
		if !once {
			log.Infof("Started audio RTP loop on port %v", artp.Port)
		}
		n, _, err := artp.listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			log.Errorf("Error reading RTP packet: %v\n", err)
			return
		}
		if !once {
			log.Debugf("Received UDP packet")
			once = true
		}

		if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
			log.Fatalf("Failed to unmarshal RTP packet: %v\n", err)
		}
		audioBuilder.Push(packet)
		for {
			sample := audioBuilder.Pop()
			if sample == nil {
				break
			}

			if writeErr := artp.Track.WriteSample(*sample); writeErr != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					// The peerConnection has been closed.
					break
				}
				panic(err)
			}
		}
	}
	// log.Debugf("Stopped audio RTP loop")
}

func (rtp *AudioRTP) Stop() {
	rtp.mutex.Lock()
	// TODO: Corner-cases.
	rtp.running = false
	rtp.stopped = false
	rtp.listener.Close()
	rtp.mutex.Unlock()
	rtp.wg.Wait()
}

func SetupExternalRTP(port int) *AudioRTP {
	// Open a UDP Listener for RTP Packets on port
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		panic(err)
	}

	// Create audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		panic(err)
	}

	// Read RTP packets forever and send them to the WebRTC Client
	return &AudioRTP{
		port,
		listener,
		audioTrack,
		sync.Mutex{},
		sync.WaitGroup{},
		false,
		true,
	}
}
