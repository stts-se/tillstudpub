package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

func sendFile(c *websocket.Conn, audioFile string) {
	f, err := os.Open(audioFile)
	if err != nil {
		log.Fatalf("Couldn't open audio file %s: %v", audioFile, err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	stride := 1024
	buf := make([]byte, 0, stride)
	nTotal := 0
	log.Println("Reading from file")
	for {
		n, err := io.ReadFull(r, buf[:cap(buf)])
		buf = buf[:n]
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				log.Printf("Got error: %v", err)
				break
			}
		}

		//fmt.Println("read n bytes...", n)
		nTotal += n
		err = c.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			log.Fatalf("Couldn't write bytes to websocket: %v", err)
		}
	}
	log.Printf("Read %v bytes from file %s", nTotal, audioFile)
}

func sendStdin(c *websocket.Conn) {
	stride := 1024
	nTotal := 0
	// Pipe stdin to the API.
	buf := make([]byte, stride)
	log.Println("Reading from stdin")
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			nTotal += n
			err = c.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Fatalf("Couldn't write bytes to websocket: %v", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Could not read from stdin: %v", err)
			continue
		}
	}
	log.Printf("Read %v bytes from stdin", nTotal)
}

type Message struct {
	Label string `json:"label"`
	Error string `json:"error,omitempty"`
	//Transcript *Transcript `json:"transcript,omitempty"`
	Handshake *Handshake `json:"handshake,omitempty"`
}

type Handshake struct {
	// sent from client to server
	SampleRate   int    `json:"sample_rate"`
	ChannelCount int    `json:"channel_count"`
	Encoding     string `json:"encoding,omitempty"`
	UserAgent    string `json:"user_agent"`
	Timestamp    string `json:"timestamp,omitempty"`

	UserName string `json:"user,omitempty"`
	Project  string `json:"project,omitempty"`
	Session  string `json:"session,omitempty"`

	UUID *uuid.UUID `json:"uuid,omitempty"` // sent from server to client
}

type Word struct {
	Word      string `json:"word,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
}

// type Transcript struct {
// 	Words     *[]Word `json:"words,omitempty"`
// 	Text      string  `json:"text,omitempty"`
// 	IsFinal   bool    `json:"is_final"`
// 	StartTime string  `json:"start_time,omitempty"`
// 	EndTime   string  `json:"end_time,omitempty"`
// }

func listenToResults(c *websocket.Conn) {
	for {
		var result = Message{}
		err := c.ReadJSON(&result)
		if err != nil {
			log.Printf("Failed to read websocket : %v", err)
			break
		}
		// if result.Label == "transcript" {
		// 	if result.Transcript.IsFinal {
		// 		fmt.Println()
		// 		fmt.Println(result.Transcript.Text)
		// 	} else {
		// 		fmt.Printf("\r                                                          \r> %s", result.Transcript.Text)
		// 	}
		// }
	}
}

func writeMessageToSocket(msg Message, socket *websocket.Conn) error {
	jsb, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal json %v : %v", msg, err)
	}
	if err = socket.WriteMessage(websocket.TextMessage, jsb); err != nil {
		return fmt.Errorf("failed to write json %v : %v", msg, err)
	}
	return nil
}

func readMessageFromSocket(socket *websocket.Conn) (Message, error) {
	var res Message
	mType, bts, err := socket.ReadMessage()
	if err != nil {
		return res, fmt.Errorf("could not read from websocket : %v", err)
	}
	if mType == websocket.TextMessage {
		if err := json.Unmarshal(bts, &res); err != nil {
			return res, fmt.Errorf("could not parse json : %v", err)
		}
	}
	return res, nil
}

func doHandshakes(c *websocket.Conn) error {
	msg := Message{
		Label: "handshake",
		Handshake: &Handshake{
			Timestamp:    time.Now().String(),
			SampleRate:   *sampleRate,
			Encoding:     *encoding,
			ChannelCount: *channelCount,
			UserAgent:    "gocli",
			UserName:     *userName,
			Session:      *session,
			Project:      *project,
		},
	}
	if err := writeMessageToSocket(msg, c); err != nil {
		return fmt.Errorf("failed to write handshake to server %v : %v", msg, err)
	}
	log.Println("Waiting for server to return handshake")
	msg, err := readMessageFromSocket(c)
	if err != nil {
		return fmt.Errorf("failed to read handshake from server : %v", err)
	}
	if msg.Error != "" {
		return fmt.Errorf("got error from server : %s", msg.Error)
	}
	log.Println("Server handshake received")
	return nil
}

var channelCount, sampleRate *int
var encoding, userName, project, session *string

func main() {

	var cmd = filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] audioFile\n", cmd)
		fmt.Fprintf(os.Stderr, " to record from stdin, use audioFile -\n")
		fmt.Fprintf(os.Stderr, "\nSample usage: rec -r 48000 -t flac -c 2 - 2> /dev/null  | go run . -channels 2 -sample_rate 48000 -encoding flac -host 127.0.0.1 -port 7651 -\n")
		fmt.Fprintf(os.Stderr, "\nFlags\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var defaultUser string
	if usr, err := user.Current(); err == nil {
		defaultUser = usr.Username
	}

	host := flag.String("host", "127.0.0.1", "Server host")
	port := flag.String("port", "7651", "Server port")
	channelCount = flag.Int("channels", 2, "Number of channels")
	sampleRate = flag.Int("sample_rate", 48000, "Sample rate")
	encoding = flag.String("encoding", "flac", "Audio encoding")
	userName = flag.String("user", defaultUser, "User name")
	session = flag.String("session", "", "Session name")
	project = flag.String("project", "", "Project name")

	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
	}

	audioFile := flag.Args()[0]

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%s", *host, *port), Path: "/ws/register"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Dial error: %v", err)
	}
	defer c.Close()

	if err := doHandshakes(c); err != nil {
		log.Fatalf("Handshake failed: %v", err)
	}

	if audioFile == "-" {
		go sendStdin(c)
	} else {
		go sendFile(c, audioFile)
	}

	go listenToResults(c)
	select {}

}
