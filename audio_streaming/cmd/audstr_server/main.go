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
	"github.com/stts-se/tillstudpub/audiostreaming"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/cryptix/wav"
)

var upgrader = websocket.Upgrader{}

const outputDir = "data"

var latestAudioRawFileName = filepath.Join(outputDir, "latest.raw")
var latestAudioWavFileName = filepath.Join(outputDir, "latest.wav")
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

func parseFlags() config {
	var cmd = "audstr_server"

	var cfg config
	cfg.host = flag.String("host", "127.0.0.1", "Server host")
	cfg.port = flag.String("port", "7651", "Server port")
	cfg.saveRaw = flag.Bool("raw", false, "Save audio raw files")
	cfg.noWav = flag.Bool("nowav", false, "Skip wav output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd)
		flag.PrintDefaults()
	}
	flag.Parse()

	return cfg
}

// server config
type config struct {
	host, port     *string
	saveRaw, noWav *bool
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

	go receiveAudioStream(&handshake, conn)
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

func writeMessageToSocket(msg audiostreaming.Message, socket *websocket.Conn) (string, error) {
	jsb, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json %v : %v", msg, err)
	}
	if err = socket.WriteMessage(websocket.TextMessage, jsb); err != nil {
		return "", fmt.Errorf("failed to write json %v : %v", msg, err)
	}
	return string(jsb), nil
}

func readMessageFromSocket(socket *websocket.Conn) (audiostreaming.Message, string, error) {
	var res audiostreaming.Message
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

func initialiseAudioStream(conn *websocket.Conn) (audiostreaming.Handshake, error) {
	log.Printf("Opened audio stream")

	var handshake audiostreaming.Handshake
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
	returnMsg := audiostreaming.Message{
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

func createWavHeader(audioConfig *audiostreaming.AudioConfig) wav.File {
	// type File struct {
	// 	SampleRate      uint32
	// 	SignificantBits uint16
	// 	Channels        uint16
	// 	NumberOfSamples uint32
	// 	Duration        time.Duration
	// 	AudioFormat     uint16
	// 	SoundSize       uint32
	// 	Canonical       bool
	// 	BytesPerSecond  uint32
	// }

	// Create the headers for our new mono file
	// TODO: problematic conversion int => unitNN ?
	res := wav.File{
		Channels:        uint16(audioConfig.ChannelCount),
		SampleRate:      uint32(audioConfig.SampleRate),
		SignificantBits: uint16(audioConfig.BitDepth), // TODO: Significant Bits = Bit Depth?
	}
	if audioConfig.Encoding == "pcm" {
		res.AudioFormat = uint16(1)
	}
	return res
}

func receiveAudioStream(handshake *audiostreaming.Handshake, audioStreamSender *websocket.Conn) {
	defer audioStreamSender.Close() // ??
	var err error

	byteCount := 0

	var rawWriter bufferedFileWriter
	var audioRawFileName string
	if *cfg.saveRaw {
		audioRawFileName = filepath.Join(outputDir, fmt.Sprintf("%s.raw", handshake.UUID.String()))
		rawWriter, err = newBufferedFileWriter(audioRawFileName)
		if err != nil {
			log.Printf("Couldn't open raw audio file for writing: %v", err)
			return
		}
	}

	var wavWriter *wav.Writer
	var audioWavFileName string
	if !*cfg.noWav {
		wavHeader := createWavHeader(handshake.AudioConfig)
		audioWavFileName = filepath.Join(outputDir, fmt.Sprintf("%s.wav", handshake.UUID.String()))

		wavFile, err := os.Create(audioWavFileName)
		if err != nil {
			log.Printf("Couldn't open wav audio file for writing: %v", err)
			return
		}

		wavWriter, err = wavHeader.NewWriter(wavFile)
		if err != nil {
			log.Printf("Couldn't create wav audio file writer: %v", err)
			return
		}
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
			byteCount += len(bts)
			if *cfg.saveRaw {
				if _, err := rawWriter.writer.Write(bts); err != nil {
					log.Printf("Couldn't write raw audio to file: %v", err)
					break
				}
			}

			//--------------------- TODO -------------------------

			if !*cfg.noWav {
				if _, err := wavWriter.Write(bts); err != nil {
					log.Printf("Couldn't write wav audio to file: %v", err)
					break
				}
			}

			//log.Printf("Wrote %v bytes of audio to file", len(bts))
		}
	}

	if !*cfg.noWav {
		if err := wavWriter.Close(); err != nil {
			log.Printf("Couldn't close wav writer: %v", err)
			return
		}
		log.Printf("Saved wav audio file %s", audioWavFileName)

		// copy wav to latest.wav
		latestAudioFileMutex.Lock()
		if err := copyFile(audioWavFileName, latestAudioWavFileName); err != nil {
			log.Printf("Couldn't copy wav audio file to %s: %v", latestAudioWavFileName, err)
			return
		}
		log.Printf("Saved wav audio file %s", latestAudioWavFileName)
		latestAudioFileMutex.Unlock()
	} else {
		os.Remove(latestAudioWavFileName)
	}

	if *cfg.saveRaw {
		// save raw file
		if err := rawWriter.close(); err != nil {
			log.Printf("Couldn't save raw audio file: %v", err)
			return
		}
		log.Printf("Saved raw audio file %s (%v bytes)", audioRawFileName, byteCount)
		// copy wav to latest.wav
		latestAudioFileMutex.Lock()
		if err := copyFile(audioRawFileName, latestAudioRawFileName); err != nil {
			log.Printf("Couldn't copy raw audio file to %s: %v", latestAudioRawFileName, err)
			return
		}
		log.Printf("Saved raw audio file %s", latestAudioRawFileName)
		latestAudioFileMutex.Unlock()
	} else {
		os.Remove(latestAudioRawFileName)
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
