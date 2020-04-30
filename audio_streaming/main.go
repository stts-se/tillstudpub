package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/cryptix/wav"
)

var upgrader = websocket.Upgrader{}

const outputDir = "data"

//var latestAudioWavFileName = filepath.Join(outputDir, "latest.wav")
var latestAudioRawFileName = filepath.Join(outputDir, "latest.raw")
var latestJSONFileName = filepath.Join(outputDir, "latest.json")
var latestAudioFileMutex = &sync.Mutex{}
var latestJSONFileMutex = &sync.Mutex{}

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

//Message is a struct for sending messages over websockets
type Message struct {
	Label     string     `json:"label"`
	Error     string     `json:"error,omitempty"`
	Handshake *Handshake `json:"handshake,omitempty"`
}

//AudioConfig contains settings for audio
type AudioConfig struct {
	SampleRate   int    `json:"sample_rate"`
	ChannelCount int    `json:"channel_count"`
	Encoding     string `json:"encoding"`
}

//Handshake is a struct for sending handshakes over websockets
type Handshake struct {
	// sent from client to server
	AudioConfig *AudioConfig `json:"audio_config"`

	StreamingMethod string `json:"streaming_method"`
	UserAgent       string `json:"user_agent"`
	Timestamp       string `json:"timestamp"`

	UserName string `json:"user_name"`
	Project  string `json:"project"`
	Session  string `json:"session"`

	UUID *uuid.UUID `json:"uuid"` // sent from server to client
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

	jsonFileName := filepath.Join(outputDir, fmt.Sprintf("%s.json", id.String()))
	if writer, err = newBufferedFileWriter(jsonFileName); err != nil {
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
	if err := writer.close(); err != nil {
		return handshake, fmt.Errorf("couldn't save json file : %v", err)
	}
	log.Printf("Saved json file %s", jsonFileName)
	latestJSONFileMutex.Lock()
	if err := copyFile(jsonFileName, latestJSONFileName); err != nil {
		return handshake, fmt.Errorf("couldn't copy json file to %s : %v", latestJSONFileName, err)
	}
	log.Printf("Saved audio file %s", latestJSONFileName)

	latestJSONFileMutex.Unlock()

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

	audioRawFileName := filepath.Join(outputDir, fmt.Sprintf("%s.raw", id.String()))

	audioWavFileName := filepath.Join(outputDir, fmt.Sprintf("%s.wav", id.String()))

	//audioWavFileName := filepath.Join(outputDir, fmt.Sprintf("%s.wav", id.String()))
	audioW, err := newBufferedFileWriter(audioRawFileName)
	if err != nil {
		log.Printf("Couldn't open raw audio file for writing: %v", err)
		return
	}

	// ------------------------- TODO ---------------------------
	//TODO No harwired values for wav
	// Create the headers for our new mono file
	meta := wav.File{
		Channels:        1,
		SampleRate:      44100,
		SignificantBits: 16,
	}

	f, err := os.Create(audioWavFileName)
	if err != nil {
		log.Printf("Couldn't open wav audio file for writing: %v", err)
		return
	}

	wavWriter, err := meta.NewWriter(f)
	if err != nil {
		log.Printf("Couldn't create wav audio file writer: %v", err)
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
				log.Printf("Couldn't write raw audio to file: %v", err)
				break
			}

			//--------------------- TODO -------------------------

			if _, err := wavWriter.Write(bts); err != nil {
				log.Printf("Couldn't write wav audio to file: %v", err)
				break
			}

			//log.Printf("Wrote %v bytes of audio to file", len(bts))
		}
	}

	wavWriter.Close()
	f.Close()

	// save raw file
	if err := audioW.close(); err != nil {
		log.Printf("Couldn't save raw audio file: %v", err)
		return
	}
	log.Printf("Saved raw audio file %s", audioRawFileName)

	latestAudioFileMutex.Lock()
	if err := copyFile(audioRawFileName, latestAudioRawFileName); err != nil {
		log.Printf("Couldn't copy raw audio file to %s: %v", latestAudioRawFileName, err)
		return
	}
	log.Printf("Saved raw audio file %s", latestAudioRawFileName)
	latestAudioFileMutex.Unlock()

	// save wav file
	//if err := writeWavFile(handshake, audioRawFileName, audioWavFileName); err != nil {
	//	log.Printf("Couldn't save wav audio file: %v", err)
	//	return
	//}
	//log.Printf("Saved wav audio file %s", audioWavFileName)
	//
	//latestAudioFileMutex.Lock()
	//if err := copyFile(audioWavFileName, latestAudioWavFileName); err != nil {
	//	log.Printf("Couldn't copy wav audio file to %s: %v", latestAudioWavFileName, err)
	//	return
	//}
	//log.Printf("Saved wav audio file %s", latestAudioWavFileName)
	//latestAudioFileMutex.Unlock()
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
