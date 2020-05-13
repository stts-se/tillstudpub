package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func validAudioFileExtension(ext string) bool {
	return ext == "wav"
	//return (ext == "opus" || ext == "mp3" || ext == "wav")
}

// save the original audio from the client + a set of additional versions
func writeAudioFile(audioDir string, input SessionAudioInput) (string, error) {

	var err error

	// writeMutex declared in recserver.go
	writeMutex.Lock()
	defer writeMutex.Unlock()

	_, err = os.Stat(audioDir)

	if os.IsNotExist(err) {
		// create subdir input_audio to keep original audio from client
		err = os.MkdirAll(audioDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("writeAudioFile: failed to create dir : %v", err)
		}
	}

	if strings.TrimSpace(input.Audio.FileType) == "" {
		msg := fmt.Sprintf("input audio for '%s' has no associated file type", input.UttID)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}

	var ext string
	for _, e := range []string{"webm", "wav", "ogg", "mp3"} {
		if strings.Contains(input.Audio.FileType, e) {
			ext = e
			break
		}
	}

	if ext == "" {
		msg := fmt.Sprintf("unknown file type for '%s': %s", input.UttID, input.Audio.FileType)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}

	// generate next running number for file with same recordingID. Starts at "0001"
	// always returns, with default returnvaule "0001"
	runningNum := "0001" // generateNextFileNum(audioDir, input.UttID) // TODO?
	audioFile := filepath.Join(audioDir, fmt.Sprintf("%s_%s.%s", input.UttID, runningNum, ext))

	var audio []byte
	audio, err = base64.StdEncoding.DecodeString(input.Audio.Data)
	if err != nil {
		msg := fmt.Sprintf("failed audio base64 decoding : %v", err)
		log.Println(msg)
		return "", fmt.Errorf("%s : %v", msg, err)
	}

	// (1) Save original audio input file (whatever extension/format)
	err = ioutil.WriteFile(audioFile, audio, 0644)
	if err != nil {
		msg := fmt.Sprintf("failed to write audio file : %v", err)
		log.Println(msg)
		return "", fmt.Errorf("%s : %v", msg, err)
	}

	log.Printf("[writeAudioFile] Saved %s", audioFile)

	// (2) ALWAYS convert to wav 16kHz MONO
	// ffmpegConvert function is defined in ffmpegConvert.go
	audioFileWav := strings.TrimSuffix(audioFile, "."+ext)
	audioFileWav = fmt.Sprintf("%s.%s", audioFileWav, cleanExt(audioExt))
	err = ffmpegConvert(audioFile, audioFileWav, false)
	if err != nil {
		msg := fmt.Sprintf("writeAudioFile failed converting from %s to %s : %v", audioFile, audioFileWav, err)
		log.Print(msg)
		return "", fmt.Errorf(msg)
	}
	log.Printf("[writeAudioFile] Saved %s", audioFileWav)

	// if !validAudioFileExtension(audioExt) {
	// 	msg := fmt.Sprintf("writeAudioFile unknown default extension: %s", audioExt)
	// 	log.Print(msg)
	// 	return "", fmt.Errorf(msg)
	// }
	return audioFileWav, nil
}
