package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/stts-se/tillstudpub/rispik/logger"

	"github.com/gorilla/mux"
)

// http://<host>:<port>/get_audio_for_uuid/d2e7b980-c6c4-11ea-86e0-8c89a580ea9c
func getAudio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	logger.Infof("got UUID '%s'", uuid)

	fn := filepath.Join(outputDir, uuid+".wav")

	bts, err := ioutil.ReadFile(fn)
	if err != nil {
		logger.Errorf("failed to read audio file : %v", err)
		http.Error(w, "failed to find audiofile", http.StatusBadRequest)
		return
	}

	audioBase64 := base64.StdEncoding.EncodeToString(bts)
	fmt.Fprint(w, audioBase64)
}
