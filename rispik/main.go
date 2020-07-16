package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stts-se/tillstudpub/rispik/logger"
	"github.com/stts-se/tillstudpub/rispik/protocol"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
	logger.Error(serverMsg)
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
		logger.Fatal(err)
	}
	defer sourceFile.Close()

	newFile, err := os.Create(toFile)
	if err != nil {
		logger.Fatal(err)
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, sourceFile); err != nil {
		return err
	}
	//logger.Debug("Copied %d bytes.", bytesCopied)
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
	logger.Infof("Registered audio stream sender socket")

	handshake, err := initialiseStream(conn)
	if err != nil {
		logger.Fatal(err) // TODO
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

func writeMessageToSocket(msg protocol.Message, socket *websocket.Conn) (string, error) {
	jsb, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json %v : %v", msg, err)
	}
	if err = socket.WriteMessage(websocket.TextMessage, jsb); err != nil {
		return "", fmt.Errorf("failed to write json %v : %v", msg, err)
	}
	return string(jsb), nil
}

func readMessageFromSocket(socket *websocket.Conn) (protocol.Message, string, error) {
	var res protocol.Message
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

func initialiseStream(conn *websocket.Conn) (protocol.Handshake, error) {
	logger.Info("Opened audio stream")

	var handshake protocol.Handshake
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
	logger.Infof("Received handshake from client: %s", js)

	handshake = *msg.Handshake
	handshake.UUID = &id
	returnMsg := protocol.Message{
		Label:     "handshake",
		Handshake: &handshake,
	}

	if err = writer.writeJSON(handshake); err != nil {
		return handshake, fmt.Errorf("failed to write handshake to file %v : %v", handshake, err)
	}
	if err := writer.close(); err != nil {
		return handshake, fmt.Errorf("couldn't save json file : %v", err)
	}
	logger.Infof("Saved json file %s", jsonFileName)
	latestJSONFileMutex.Lock()
	if err := copyFile(jsonFileName, latestJSONFileName); err != nil {
		return handshake, fmt.Errorf("couldn't copy json file to %s : %v", latestJSONFileName, err)
	}
	logger.Infof("Saved audio file %s", latestJSONFileName)

	latestJSONFileMutex.Unlock()

	js, err = writeMessageToSocket(returnMsg, conn)
	if err != nil {
		return handshake, fmt.Errorf("received non-handshake message from websocket : %v", err)
	}
	logger.Infof("Sent handshake message to client: %s", js)

	if msg.Error != "" {
		return handshake, fmt.Errorf("received non-handshake message from websocket : %s", msg.Error)
	}

	return handshake, nil
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

	r.HandleFunc("/ws/list_audio_files_for_user", listAudioFilesForUser)
	r.HandleFunc("/get_audio_for_uuid/{uuid}", getAudio)

	r.HandleFunc("/ws/register", openDataWebsocket)
	r.HandleFunc("/ws/admin", openAdminWebsocket)

	// code in login.go
	r.HandleFunc("/list/users", listUsers)
	r.HandleFunc("/list/projects", listProjects)
	r.HandleFunc("/list/sessions", listSessions)

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	r.StrictSlash(true)
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", *cfg.host, *cfg.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//fmt.Fprintf(os.Stderr, "Server started on %s\n", srv.Addr)
	logger.Infof("Server started on %s\n", srv.Addr)

	logger.Fatal(srv.ListenAndServe())
	fmt.Println("No fun")

}
