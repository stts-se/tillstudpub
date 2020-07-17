package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stts-se/tillstudpub/rispik/login"
)

func writeJSONResponse(w http.ResponseWriter, source interface{}) error {
	//logger.Infof("login debug %#v\n", source)
	jsb, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal json %v : %v", source, err)
	}
	if _, err := w.Write(jsb); err != nil {
		return fmt.Errorf("couldn't write to buffer : %v", err)
	}
	if _, err := w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("couldn't write to buffer : %v", err)
	}
	return nil
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	err := writeJSONResponse(w, login.Users)
	if err != nil {
		cMsg := "Couldn't write JSON"
		sMsg := fmt.Sprintf("%s: %v", cMsg, err)
		httpError(w, sMsg, cMsg, http.StatusInternalServerError)
	}
}
func listSessions(w http.ResponseWriter, r *http.Request) {
	err := writeJSONResponse(w, login.Sessions)
	if err != nil {
		cMsg := "Couldn't write JSON"
		sMsg := fmt.Sprintf("%s: %v", cMsg, err)
		httpError(w, sMsg, cMsg, http.StatusInternalServerError)
	}
}

func listProjects(w http.ResponseWriter, r *http.Request) {
	err := writeJSONResponse(w, login.Projects)
	if err != nil {
		cMsg := "Couldn't write JSON"
		sMsg := fmt.Sprintf("%s: %v", cMsg, err)
		httpError(w, sMsg, cMsg, http.StatusInternalServerError)
	}
}
