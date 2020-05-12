// SOURCE: https://github.com/pion/webrtc-voicemail

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media/oggwriter"
)

const sampleRate = 48000
const channelCount = 2

// func stopRecording(w http.ResponseWriter, r *http.Request) {
// }

func receiveRecording(w http.ResponseWriter, r *http.Request) {
	sdp := webrtc.SessionDescription{}
	if err := json.NewDecoder(r.Body).Decode(&sdp); err != nil {
		panic(err)
	}

	// Create a MediaEngine object to configure the supported codec
	m := webrtc.MediaEngine{}
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, sampleRate))

	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Received peer connection state change: %v\n", state)
	})

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("Received ICE connection state change: %v\n", state)
	})

	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		if track.Codec().Name != webrtc.Opus {
			return
		}

		uuid, err := generateUUID()
		if err != nil {
			panic(err)
		}
		fileName := fmt.Sprintf("data/%s.ogg", uuid)

		oggFile, err := oggwriter.New(fileName, sampleRate, channelCount)
		defer oggFile.Close()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Got Opus track, writing to disk as %s (%v Hz, %v channels) \n", fileName, sampleRate, channelCount)

		for {
			rtpPacket, err := track.ReadRTP()
			//fmt.Println(rtpPacket)
			if err != nil {
				panic(err)
			}
			if err := oggFile.WriteRTP(rtpPacket); err != nil {
				panic(err)
			}
		}
		fmt.Println("Reached the end of peerConnection.OnTrack")
	})

	// Allow us to receive 1 audio track.
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	if err = peerConnection.SetRemoteDescription(sdp); err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(answer, "", "  ")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(output); err != nil {
		panic(err)
	}
}

func main() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err = os.Mkdir("data", 0755); err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, r.URL.Path[1:]) })
	http.HandleFunc("/start", receiveRecording)
	//http.HandleFunc("/stop", stopRecording)

	port := 7659
	fmt.Printf("Server has started on :%v\n", port)
	panic(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func generateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.New(rand.NewSource(time.Now().UnixNano())).Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
