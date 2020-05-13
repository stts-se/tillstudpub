package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func writeJSONInfoFile(audioFile string, input SessionAudioInput) error {

	prettyInput := input
	prettyInput.Audio.Data = ""

	audioExt := cleanExt(filepath.Ext(audioFile))
	jsonFile := strings.TrimSuffix(audioFile, "."+audioExt)
	jsonFile = fmt.Sprintf("%s.%s", jsonFile, "json")

	// writeMutex declared in recserver.go
	writeMutex.Lock()
	defer writeMutex.Unlock()

	prettyJSON, err := prettyMarshal(prettyInput)
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to create info JSON : %v", err)
	}
	jsonHandle, err := os.Create(jsonFile)
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to create info file : %v", err)
	}
	defer jsonHandle.Close()

	_, err = jsonHandle.WriteString(string(prettyJSON) + "\n")
	if err != nil {
		return fmt.Errorf("writeJSONInfoFile: failed to write info file : %v", err)
	}
	log.Printf("[writeJSONInfoFile] Saved %s", jsonFile)

	return nil
}
