package record

import (
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	require := require.New(t)

	r, err := NewRecorder("hw:1,0", 4444)
	require.Nil(err)
	require.NotNil(r)

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4444})
	if err != nil {
		panic(err)
	}

	pkt := make([]byte, 1600)

	r.Start()

	start := time.Now()
	read := 0
	for {
		now := time.Now()
		if now.Sub(start) > 300*time.Millisecond {
			break
		}
		n, _, err := listener.ReadFrom(pkt)
		require.Nil(err)
		read += n
	}
	require.Greater(read, 0)
	r.Stop()
}
