package main

import (
	"bufio"
	//"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	//"io"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

const outputDir = "data"

var latestAudioFileName = filepath.Join(outputDir, "latest.raw")
var latestJSONFileName = filepath.Join(outputDir, "latest.json")

// print serverMsg to server log, and return an http error with clientMsg and the specified error code (http.StatusInternalServerError, etc)
func httpError(w http.ResponseWriter, serverMsg string, clientMsg string, errCode int) {
	log.Println(serverMsg)
	http.Error(w, clientMsg, errCode)
}

func readFile(fName string) ([]string, error) {
	bytes, err := ioutil.ReadFile(fName)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(strings.TrimSpace(string(bytes)), "\n"), nil
}

func copyFile(fromFile, toFile string) error {
	sourceFile, err := os.Open(fromFile)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceFile.Close()

	newFile, err := os.Create(toFile)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, sourceFile); err != nil {
		return err
	}
	//log.Printf("Copied %d bytes.", bytesCopied)
	return nil
}

type language struct {
	Code, Description string
}

type encoding struct {
	Code, Description string
}

var languages = []language{
	{"sv-SE", "Swedish (Sweden)"},
	{"en-US", "English (US)"},
	{"da-DK", "Danish (Denmark)"},
	{"en", "English"},
	{"en-UK", "English (UK)"},
}

var encodings = []encoding{
	{"amr", ""},
	{"amr_wb", ""},
	{"flac", ""},
	{"pcm", ""},
	{"linear16", ""},
	{"mulaw", ""},
	{"ogg_opus", ""},
	{"ogg", ""},
	{"opus", ""},
	{"speex_hb", ""},
}

func getEncoding(encName string) (encoding, bool) {
	for _, enc := range encodings {
		if enc.Code == encName {
			return enc, true
		}
	}
	return encoding{}, false
}

func getLanguage(langName string) (language, bool) {
	for _, lang := range languages {
		if lang.Code == langName {
			return lang, true
		}
	}
	return language{}, false
}

func parseFlags() config {
	var cmd = filepath.Base(os.Args[0])

	var cfg config
	cfg.language = flag.String("language", "sv-SE", "Audio input language code")
	cfg.encoding = flag.String("encoding", "flac", "Audio input encoding")
	cfg.host = flag.String("host", "127.0.0.1", "Server host")
	cfg.port = flag.String("port", "7651", "Server port")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd)
		flag.PrintDefaults()
	}
	flag.Parse()

	if _, ok := getEncoding(*cfg.encoding); !ok {
		fmt.Fprintf(os.Stderr, "Invalid encoding: %s\n", *cfg.encoding)
		fmt.Fprintf(os.Stderr, "Valid encodings: ")
		for i, e := range encodings {
			fmt.Fprintf(os.Stderr, e.Code)
			if i < len(encodings)-1 {
				fmt.Fprintf(os.Stderr, ", ")
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}
	if _, ok := getLanguage(*cfg.language); !ok {
		fmt.Fprintf(os.Stderr, "Invalid language code: %s\n", *cfg.language)
		fmt.Fprintf(os.Stderr, "Valid language codes: ")
		for i, l := range languages {
			fmt.Fprintf(os.Stderr, l.Code)
			if i < len(languages)-1 {
				fmt.Fprintf(os.Stderr, ", ")
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}

	return cfg
}

// server config
type config struct {
	language, encoding, host, port *string
}

func openDataWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		msg := fmt.Sprintf("failed to upgrade to ws: %v", err)
		httpError(w, msg, "Failed to upgrade to ws", http.StatusBadRequest)
		return
	}
	log.Println("Registered audio stream sender socket")

	handshake, err := initialiseAudioStream(conn)
	if err != nil {
		log.Fatal(err) // TODO
	}

	go receiveAudioStream(*handshake.UUID, conn)

	//go listenForResults(*handshake.UUID, jsonWriter, conn)
}

type bufferedFileWriter struct {
	fileName string
	file     *os.File
	writer   *bufio.Writer
}

func newBufferedFileWriter(fName string) (bufferedFileWriter, error) {
	var res bufferedFileWriter
	f, err := os.Create(fName)
	if err != nil {
		return res, fmt.Errorf("Could not open %s for writing : %v", fName, err)
	}
	buf := bufio.NewWriter(f)
	return bufferedFileWriter{fileName: fName, file: f, writer: buf}, nil
}

func (w bufferedFileWriter) write(s string) error {
	if _, err := w.writer.Write([]byte(s)); err != nil {
		return fmt.Errorf("couldn't write to buffer : %v", err)
	}
	return nil
}

func (w bufferedFileWriter) writeJSON(source interface{}) error {
	jsb, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal json %v : %v", source, err)
	}
	if _, err := w.writer.Write(jsb); err != nil {
		return fmt.Errorf("couldn't write to buffer : %v", err)
	}
	if _, err := w.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("couldn't write to buffer : %v", err)
	}
	return nil
}

func (w bufferedFileWriter) close() error {
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("couldn't flush buffer : %v", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("couldn't save file : %v", err)
	}
	return nil
}

type Message struct {
	Label      string      `json:"label"`
	Error      string      `json:"error,omitempty"`
	Transcript *Transcript `json:"transcript,omitempty"`
	Handshake  *Handshake  `json:"handshake,omitempty"`
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

type Transcript struct {
	Words     *[]Word `json:"words,omitempty"`
	Text      string  `json:"text,omitempty"`
	IsFinal   bool    `json:"is_final"`
	StartTime string  `json:"start_time,omitempty"`
	EndTime   string  `json:"end_time,omitempty"`
}

func writeMessageToSocket(msg Message, socket *websocket.Conn) (string, error) {
	jsb, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json %v : %v", msg, err)
	}
	if err = socket.WriteMessage(websocket.TextMessage, jsb); err != nil {
		return "", fmt.Errorf("failed to write json %v : %v", msg, err)
	}
	return string(jsb), nil
}

func readMessageFromSocket(socket *websocket.Conn) (Message, string, error) {
	var res Message
	mType, bts, err := socket.ReadMessage()
	if err != nil {
		return res, "", fmt.Errorf("could not read from websocket : %v", err)
	}
	if mType == websocket.TextMessage {
		if err := json.Unmarshal(bts, &res); err != nil {
			return res, "", fmt.Errorf("could not parse json : %v", err)
		}
	}
	return res, string(bts), nil
}

func initialiseAudioStream(conn *websocket.Conn) (Handshake, error) {
	log.Printf("Opened audio stream")

	var handshake Handshake
	var id uuid.UUID
	var err error
	var writer bufferedFileWriter

	id, err = uuid.NewUUID()
	if err != nil {
		return handshake, fmt.Errorf("couldn't create uuid : %v", err)
	}

	//ctx := context.Background()

	if writer, err = newBufferedFileWriter(filepath.Join(outputDir, fmt.Sprintf("%s.json", id.String()))); err != nil {
		return handshake, fmt.Errorf("couldn't connect recogniser context : %v", err)
	}

	// Receive handshake with config/settings, and return handshake confirmation + uuid (or error)
	msg, js, err := readMessageFromSocket(conn)
	if err != nil {
		return handshake, fmt.Errorf("couldn't read from websocket : %v", err)
	}
	if msg.Label != "handshake" {
		return handshake, fmt.Errorf("received non-handshake message from websocket")
	}
	log.Printf("Received handshake from client: %s", js)

	handshake = *msg.Handshake
	handshake.UUID = &id
	returnMsg := Message{
		Label:     "handshake",
		Handshake: &handshake,
	}

	if err = writer.writeJSON(handshake); err != nil {
		return handshake, fmt.Errorf("failed to write handshake to file %v : %v", handshake, err)
	}

	js, err = writeMessageToSocket(returnMsg, conn)
	if err != nil {
		return handshake, fmt.Errorf("received non-handshake message from websocket : %v", err)
	}
	log.Printf("Sent handshake message to client: %s", js)

	if msg.Error != "" {
		return handshake, fmt.Errorf("received non-handshake message from websocket : %s", msg.Error)
	}

	return handshake, nil
}

func receiveAudioStream(id uuid.UUID, audioStreamSender *websocket.Conn) {
	defer audioStreamSender.Close() // ??
	var err error

	audioFileName := filepath.Join(outputDir, fmt.Sprintf("%s.raw", id.String()))
	audioW, err := newBufferedFileWriter(audioFileName)
	if err != nil {
		log.Printf("Couldn't open audio file for writing: %v", err)
		return
	}

	log.Println("Audio stream open for input")

	for {
		mType, bts, err := audioStreamSender.ReadMessage()
		if err != nil {
			if err != nil {
				log.Printf("Could not read from websocket: %v", err)
				break
			}
		}

		if mType == websocket.CloseMessage {
			log.Println("Recevied close from client")
			break
		}

		if mType != websocket.BinaryMessage {
			log.Printf("Skipping non-binary message from websocket")
			continue
		}

		if len(bts) > 0 {
			if _, err := audioW.writer.Write(bts); err != nil {
				log.Printf("Couldn't write audio to file: %v", err)
				break
			}
		}
	}

	if err := audioW.close(); err != nil {
		log.Printf("Couldn't save audio file: %v", err)
	} else {
		log.Printf("Saved audio file %s", audioFileName)
		if err := copyFile(audioFileName, latestAudioFileName); err != nil {
			log.Printf("Couldn't copy audio file to %s: %v", latestAudioFileName, err)
		} else {
			log.Printf("Saved audio file %s", latestAudioFileName)
		}
	}
}

var cfg config

func main() {

	// Create the out put dir if it doesn't exist
	_, sErr := os.Stat(outputDir)
	if os.IsNotExist(sErr) {
		os.Mkdir(outputDir, os.ModePerm)
	}

	cfg = parseFlags()

	r := mux.NewRouter()

	r.HandleFunc("/ws/register", openDataWebsocket)

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	r.StrictSlash(true)
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", *cfg.host, *cfg.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Server started on %s\n", srv.Addr)

	log.Fatal(srv.ListenAndServe())
	fmt.Println("No fun")

}
