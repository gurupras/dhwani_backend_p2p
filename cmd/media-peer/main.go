package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	p2p "github.com/gurupras/dhwani_backend_p2p"
	"github.com/gurupras/dhwani_backend_p2p/rtp/audio"
	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

var audioRTP *audio.AudioRTP
var ID string
var serverConn *p2p.ServerConn

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	reader(ws)
}

func reader(conn *websocket.Conn) {
	log.Infof("Starting reader\n")

	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		log.Debugf("Received message: type=%v msg=%v\n", messageType, string(p))

		var msg map[string]interface{}
		if err = json.Unmarshal(p, &msg); err != nil {
			log.Errorf("Unexpectedly received non-JSON message: %v\n", string(p))
			continue
		}

		action := msg["action"].(string)
		log.Debugf("action=%v\n", action)
		switch action {
		case "ping":
			{
				data := make(map[string]interface{})
				data["action"] = "pong"
				data["timestamp"] = time.Now().UnixMilli()
				b, _ := json.Marshal(data)
				if err = conn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Errorf("Failed to send pong: %v\n", err)
				}
				log.Debugf("Sending back pong")
			}
		case "start-rtp-server":
			{
				rawPort, ok := msg["data"]
				var port int
				if !ok {
					port = 3131
				} else {
					port = int(rawPort.(float64))
				}
				log.Debugf("port: %v\n", port)
				if audioRTP != nil {
					audioRTP.Stop()
				}
				audioRTP = audio.SetupExternalRTP(port)
				go audioRTP.Loop()

				answerData := make(map[string]interface{})
				answerData["action"] = "started-rtp-server"
				answerData["data"] = ID
				b, _ := json.Marshal(answerData)
				if err = conn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Errorf("Failed to send started-rtp-server: %v\n", err)
					return
				}
				log.Debugf("Started RTP server on port=%v\n", port)
			}
		}
	}
}

func encodeToString(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func main() {
	log.SetLevel(log.DebugLevel)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	var err error

	// This peer is going to let the server know that it has a unique ID that is 9 digits long
	// ID = encodeToString(9)
	ID = "111111111"
	log.Infof("ID=%v\n", ID)

	serverConn, err = p2p.NewServerConnection(ID)
	if err != nil {
		log.Fatalf("Failed to set up server connection")
	}
	go serverConn.Loop()

	// TODO: Remove this
	audioRTP = audio.SetupExternalRTP(3131)
	go audioRTP.Loop()

	pcMap := make(map[string]*webrtc.PeerConnection)

	serverConn.OnSignal(func(sp p2p.SignalPacket) {
		from := sp.From
		dataBytes, err := base64.StdEncoding.DecodeString(sp.Data)
		if err != nil {
			log.Errorf("Bad offer. Expected base64 encoded data: %v\n", err)
			return
		}
		switch sp.Type {
		case "offer":
			{
				offer := webrtc.SessionDescription{}
				if err = json.Unmarshal(dataBytes, &offer); err != nil {
					log.Errorf("Bad offer. Failed to unmarshal JSON: %v\n", err)
					return
				}
				peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{
						{
							URLs: []string{"stun:ice-us-east-269qnzqlg2.twoseven.xyz:4558"},
						},
						{
							URLs: []string{
								"turn:ice-us-east-269qnzqlg2.twoseven.xyz:4558?transport=udp",
								"turn:ice-us-east-269qnzqlg2.twoseven.xyz:4558?transport=tcp",
							},
							Username:       "1642471464:twoseven",
							Credential:     "PaQln41l2CoGu+W++dk7rH0eIrI=",
							CredentialType: webrtc.ICECredentialTypePassword,
						},
					},
				})
				if err != nil {
					panic(err)
				}
				pcMap[from] = peerConnection
				rtpSender, err := peerConnection.AddTrack(audioRTP.Track)
				if err != nil {
					panic(err)
				}

				// Read incoming RTCP packets
				// Before these packets are returned they are processed by interceptors. For things
				// like NACK this needs to be called.
				go func() {
					rtcpBuf := make([]byte, 1500)
					for {
						if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							return
						}
					}
				}()

				// Set the handler for ICE connection state
				// This will notify you when the peer has connected/disconnected
				peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
					fmt.Printf("Connection State has changed %s \n", connectionState.String())

					if connectionState == webrtc.ICEConnectionStateFailed {
						if closeErr := peerConnection.Close(); closeErr != nil {
							panic(closeErr)
						}
					}
				})

				peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
					if i == nil {
						return
					}
					candidateData := make(map[string]interface{})
					candidateData["type"] = "candidate"
					candidateData["candidate"] = i.ToJSON()
					b, _ := json.Marshal(candidateData)
					answerData := make(map[string]interface{})
					answerData["action"] = "signal"
					answerData["from"] = ID
					answerData["to"] = sp.From
					answerData["data"] = base64.StdEncoding.EncodeToString(b)

					b, _ = json.Marshal(answerData)
					if err = serverConn.WriteMessage(websocket.TextMessage, b); err != nil {
						log.Errorf("Failed to send ice-candidate: %v\n", err)
						return
					} else {
						// log.Debugf("Sent ICE candidate. to=%v\n", from)
					}
				})

				// Set the remote SessionDescription
				if err = peerConnection.SetRemoteDescription(offer); err != nil {
					panic(err)
				}

				// Create answer
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}

				// Create channel that is blocked until ICE Gathering is complete
				gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

				// Sets the LocalDescription, and starts our UDP listeners
				if err = peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}

				// Block until ICE Gathering is complete, disabling trickle ICE
				// we do this because we only can exchange one signaling message
				// in a production application you should exchange ICE Candidates via OnICECandidate
				<-gatherComplete
				log.Debugf("ICE gathering complete\n")

				answerData := make(map[string]interface{})
				answerData["action"] = "signal"
				answerData["from"] = ID
				answerData["to"] = sp.From
				answer.SDP = strings.Replace(answer.SDP, "useinbandfec=1", "useinbandfec=1; maxaveragebitrate=2560000", 1)
				b, _ := json.Marshal(answer)
				answerData["data"] = base64.StdEncoding.EncodeToString(b)
				b, _ = json.Marshal(answerData)
				if err = serverConn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Errorf("Failed to send answer: %v\n", err)
					return
				}
				// log.Debugf("Sent back answer to peer=%v answer=%v\n", sp.From, answer.SDP)
			}
		case "candidate":
			{
				var obj map[string]interface{}
				candidateBytes, err := json.Marshal(obj["candidate"])
				if err != nil {
					log.Errorf("Failed to marshal 'candidate' key: %v\n", err)
					return
				}
				candidate := webrtc.ICECandidateInit{}
				if err = json.Unmarshal(candidateBytes, &candidate); err != nil {
					log.Errorf("Bad candidate. Failed to unmarshal JSON: %v\n", err)
					return
				}
				peerConnection, ok := pcMap[from]
				if !ok {
					log.Errorf("Unknown peer '%v'. Ignored candidate\n", from)
					return
				}
				peerConnection.AddICECandidate(candidate)
			}
		}
	})
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/ws", wsHandler)

	log.Fatal(http.ListenAndServe(":4234", nil))

}
