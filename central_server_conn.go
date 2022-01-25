package dhwani_backend_p2p

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/recws-org/recws"
	log "github.com/sirupsen/logrus"
)

const SERVER_HOST = "dhwani-backend.gurupras.me"

type SignalPacket struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type ServerConnection interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	Close() error
}

type reconnectingConnWrapper struct {
	*recws.RecConn
}

func (r *reconnectingConnWrapper) Close() error {
	r.RecConn.Close()
	return nil
}

type ServerConn struct {
	ServerConnection
	signalCallbacks []func(SignalPacket)
	wg              sync.WaitGroup
	mutex           sync.Mutex
	started         bool
	stopped         bool
}

func (s *ServerConn) Close() error {
	s.ServerConnection.Close()
	return nil
}

func (s *ServerConn) OnSignal(cb func(SignalPacket)) func() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	length := len(s.signalCallbacks)
	s.signalCallbacks = append(s.signalCallbacks, cb)
	return func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.signalCallbacks = append(s.signalCallbacks[:length], s.signalCallbacks[length+1:]...)
		log.Debugf("Removed signal callback")
	}
}

func (s *ServerConn) Loop() {
	s.mutex.Lock()
	s.started = true
	s.stopped = false
	s.mutex.Unlock()
	s.wg.Add(1)
	defer s.wg.Done()

	once := false
	for {
		s.mutex.Lock()
		if s.stopped {
			break
		}
		s.mutex.Unlock()
		if !once {
			log.Infof("Starting server connection loop ...\n")
			once = true
		}
		_, rawMessage, err := s.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				log.Debugf("Connection closed")
				break
			} else {
				// log.Errorf("Failed to read message from server connection: %v\n", err)
			}
			continue
		}
		var msg map[string]interface{}
		if err = json.Unmarshal(rawMessage, &msg); err != nil {
			log.Errorf("Failed to parse JSON from message '%v': %v\n", string(rawMessage), err)
			continue
		}
		action := msg["action"].(string)
		switch action {
		case "signal":
			{
				sp := SignalPacket{
					From: msg["from"].(string),
					To:   msg["to"].(string),
					Type: msg["type"].(string),
					Data: msg["data"].(string),
				}
				for _, cb := range s.signalCallbacks {
					cb(sp)
				}
			}
		}
	}
	log.Warnf("Server connection loop terminated ...\n")
}

func (s *ServerConn) Stop() {
	s.mutex.Lock()
	if !s.started {
		s.mutex.Unlock()
		return
	}
	s.stopped = true
	s.mutex.Unlock()
	s.wg.Wait()
}

func NewServerConnection(id string, autoReconnect bool) (*ServerConn, error) {
	var err error
	u := url.URL{Scheme: "wss", Host: SERVER_HOST, Path: "/ws"}
	q := u.Query()
	q.Set("id", id)
	u.RawQuery = q.Encode()
	log.Infof("connecting to %s\n", u.String())

	var conn ServerConnection
	if autoReconnect {
		ws := recws.RecConn{
			KeepAliveTimeout: 10000 * time.Second,
			NonVerbose:       true,
		}
		ws.Dial(u.String(), nil)
		conn = &reconnectingConnWrapper{&ws}
	} else {
		conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
	}

	return &ServerConn{
		conn,
		nil,
		sync.WaitGroup{},
		sync.Mutex{},
		false,
		false,
	}, nil
}
