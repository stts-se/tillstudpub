package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	//"github.com/google/uuid"
	"github.com/stts-se/tillstudpub/rispik/logger"
	"github.com/stts-se/tillstudpub/rispik/protocol"

	//"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// socket for listing audio files saved on server for particular project/user/session
func listAudioFilesForUser(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		msg := fmt.Sprintf("failed to upgrade to ws: %v", err)
		httpError(w, msg, "Failed to upgrade to ws", http.StatusInternalServerError)
		return
	}
	//TODO
	logger.Info("Registered file listing sender socket")

	var res protocol.FileListingRequest
	mType, bts, err := conn.ReadMessage()
	if err != nil {
		logger.Errorf("listAudioFileForUser(): could not read from websocket : %v", err)

		// TODO send error to client

		return //res, "", fmt.Errorf("listAudioFileForUser(): could not read from websocket : %v", err)
	}
	if mType == websocket.TextMessage {
		if err := json.Unmarshal(bts, &res); err != nil {
			//TODO
			logger.Errorf("listAudioFileForUser(): could not parse json : %v", err)
			return //res, "", fmt.Errorf("listAudioFileForUser(): could not parse json : %v", err)
		}

		logger.Infof("got client request %#v", res)

		files, err := getFileList(res)
		if err != nil {
			//TODO
			logger.Errorf("fiasco : %v", err)
			return
		}

		for _, f := range files {

			jsn, err := json.Marshal(f)
			if err != nil {
				//TODO
				logger.Error(err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, jsn); err != nil {
				//TODO
				logger.Error(err)
				return
			}
			//TODO
			logger.Infof("printed to websocket : %#v", f)
		}

	}

}

func getFileList(r protocol.FileListingRequest) ([]protocol.Handshake, error) {
	res, err := listAudioFiles(outputDir)
	if err != nil {
		return res, err
	}

	res = filterUser(r.User, res)
	res = filterSession(r.Session, res)
	res = filterProject(r.Project, res)

	return res, nil
}

func listAudioFiles(dataPath string) ([]protocol.Handshake, error) {
	var res []protocol.Handshake

	jsonFiles, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		return res, fmt.Errorf("listAudioFiles failed to list files in dir '%s' : %v", outputDir, err)
	}

	for _, jsf := range jsonFiles {
		if strings.HasSuffix(jsf, "latest.json") {
			continue
		}

		jsonBts, err := ioutil.ReadFile(jsf)
		if err != nil {
			return res, fmt.Errorf("failed to read json file : %v", err)
		}

		handShake := protocol.Handshake{}
		err = json.Unmarshal(jsonBts, &handShake)
		if err != nil {
			return res, fmt.Errorf("failed to unmarshal json file : %v", err)
		}

		//logger.Infof("%#v", handShake)
		res = append(res, handShake)
	}

	return res, nil
}

func filterUser(userName string, files []protocol.Handshake) []protocol.Handshake {
	var res []protocol.Handshake
	for _, f := range files {
		if strings.ToLower(f.UserName) == strings.ToLower(userName) {
			res = append(res, f)
		}
	}

	return res
}

func filterProject(projectName string, files []protocol.Handshake) []protocol.Handshake {
	var res []protocol.Handshake
	for _, f := range files {
		if strings.ToLower(f.Project) == strings.ToLower(projectName) {
			res = append(res, f)
		}
	}

	return res
}

func filterSession(sessionName string, files []protocol.Handshake) []protocol.Handshake {
	var res []protocol.Handshake
	for _, f := range files {
		if strings.ToLower(f.Session) == strings.ToLower(sessionName) {
			res = append(res, f)
		}
	}

	return res
}
