package audio

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

type AudioRTP struct {
	port     int
	listener *net.UDPConn
	Track    *webrtc.TrackLocalStaticRTP
	mutex    sync.Mutex
	wg       sync.WaitGroup
	running  bool
	stopped  bool
}

func (rtp *AudioRTP) Loop() {
	rtp.mutex.Lock()
	rtp.running = true
	rtp.stopped = false
	rtp.mutex.Unlock()
	inboundRTPPacket := make([]byte, 1600) // UDP MTU
	rtp.wg.Add(1)
	defer func() {
		rtp.mutex.Lock()
		defer rtp.mutex.Unlock()
		rtp.stopped = true
		rtp.wg.Done()
	}()

	defer func() {
		rtp.listener.Close()
	}()

	once := false
	for {
		if !once {
			log.Infof("Started audio RTP loop on port %v", rtp.port)
		}
		n, _, err := rtp.listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			log.Errorf("Error reading RTP packet: %v\n", err)
			return
		}
		if !once {
			log.Debugf("Received UDP packet")
			once = true
		}
		if _, err = rtp.Track.Write(inboundRTPPacket[:n]); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				// The peerConnection has been closed.
				break
			}
			panic(err)
		}
	}
	log.Debugf("Stopped audio RTP loop")
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
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
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
