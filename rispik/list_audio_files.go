package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/stts-se/tillstudpub/rispik/protocol"
)

// TODO remove
const outputDir = "data"

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

		//log.Printf("%#v", handShake)
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

// TODO remove
func main() {

	fileList, err := listAudioFiles(outputDir)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fileList = filterSession("ddd", fileList)

	log.Printf("%#v\n", fileList)
}
