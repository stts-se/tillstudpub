package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// === CONSTANTS
const timestampFmt = "2006-01-02 15:04:05 CET"

// === FIELDS

var writeMutex sync.Mutex

const baseDir = "."
const textFileName = "text.txt"
const audioExt = "wav"
const lockFileName = ".lock"
const buildInfoFile = "buildinfo.txt"

var projectsDir = filepath.Join(baseDir, "projects")

// This is filled in by main, listing the URLs handled by the router,
// so that these can be shown in the generated docs.
var walkedURLs []string

type Audio struct {
	FileType string `json:"file_type,omitempty"`
	Data     string `json:"data,omitempty"`
}

type SessionAudioInput struct {
	Project   string `json:"project"`
	Session   string `json:"session"`
	User      string `json:"user"`
	Audio     Audio  `json:"audio,omitempty"`
	Text      string `json:"text"`
	UttID     string `json:"uttid"`
	Timestamp string `json:"timestamp"`
}

type ResponseJSON struct {
	Label   string `json:"label"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ProjectInfoJSON struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type AudioPromptJSON struct {
	Project      string `json:"name"`
	UttID        string `json:"uttid"`
	UttIndex     int    `json:"uttindex"`
	Instructions string `json:"instructions"`
	Text         string `json:"text"`
	Audio        string `json:"audio"`
	FileType     string `json:"file_type"`
}

// print serverMsg to server log, and return an http error with clientMsg and the specified error code (http.StatusInternalServerError, etc)
func httpError(w http.ResponseWriter, serverMsg string, clientMsg string, errCode int) {
	log.Println(serverMsg)
	http.Error(w, clientMsg, errCode)
}

func JSONResponse(w http.ResponseWriter, label string, msg string) {
	sendJS := ResponseJSON{
		Label:   label,
		Content: msg,
	}
	jsb, err := json.Marshal(sendJS)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal %v : %v", sendJS, err)
		log.Println(msg)
		return
	}
	fmt.Fprintf(w, string(jsb))
}

func generateDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>%s</title></head><body>", "STTS PromptRec: Doc")
	for _, url := range walkedURLs {
		fmt.Fprintf(w, "%s<br/>\n", url)
	}
	fmt.Fprintf(w, "</body></html>")
}

func getBuildInfo(prefix string, lines []string, defaultValue string) []string {
	for _, l := range lines {
		fs := strings.Split(l, ": ")
		if fs[0] == prefix {
			return fs
		}
	}
	return []string{prefix, defaultValue}
}

func mimeType(ext string) string {
	ext = cleanExt(ext)
	if ext == "mp3" {
		return "audio/mpeg"
	}
	return fmt.Sprintf("audio/%s", ext)
}

func prettyMarshal(thing interface{}) ([]byte, error) {
	var res []byte

	j, err := json.Marshal(thing)
	if err != nil {
		return res, err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, j, "", "\t")
	if err != nil {
		return res, err
	}
	res = prettyJSON.Bytes()
	return res, nil
}

func findUtterance(id string, utts []utterance) (utterance, error) {
	for _, utt := range utts {
		if utt.id == id {
			return utt, nil
		}
	}
	return utterance{}, fmt.Errorf("no such utterance id: %d", id)
}

func getNextPromptID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	currID := vars["utt_id"]
	utts, err := readProject(project)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}
	for i, utt := range utts {
		if i > 0 && utts[i-1].id == currID {
			JSONResponse(w, "id", utt.id)
			return
		}
	}
}

func getPrevPromptID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	currID := vars["utt_id"]
	utts, err := readProject(project)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}
	for i, utt := range utts {
		if i > 0 && utt.id == currID {
			JSONResponse(w, "id", utts[i-1].id)
			return
		}
	}
}

func utt2JSON(project string, utt utterance) AudioPromptJSON {
	var res AudioPromptJSON
	data := base64.StdEncoding.EncodeToString(utt.audio)
	res.FileType = mimeType(audioExt)
	res.Audio = data
	res.Project = project
	res.UttID = utt.id
	res.UttIndex = utt.index
	res.Text = utt.text
	res.Instructions = utt.instructions
	return res
}

func getPrompt0(project string, uttID string) (AudioPromptJSON, error) {
	utts, err := readProject(project)
	if err != nil {
		return AudioPromptJSON{}, err
	}
	utt, err := findUtterance(uttID, utts)
	if err != nil {
		return AudioPromptJSON{}, err
	}
	return utt2JSON(project, utt), nil
}

func getFirstPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	utts, err := readProject(project)
	if err != nil {
		msg := fmt.Sprintf("Couldn't load project: %v", err)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}
	if len(utts) == 0 {
		msg := "No prompts for project"
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}
	utt := utts[0]
	res := utt2JSON(project, utt)
	resJSON, err := prettyMarshal(res)
	if err != nil {
		msg := fmt.Sprintf("getPrompt: failed to create JSON from struct : %v", res)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

func getPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	uttID := vars["utt_id"]
	res, err := getPrompt0(project, uttID)
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}
	resJSON, err := prettyMarshal(res)
	if err != nil {
		msg := fmt.Sprintf("getPrompt: failed to create JSON from struct : %v", res)
		httpError(w, msg, msg, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

func generateAbout(w http.ResponseWriter, r *http.Request) {

	var buildInfoLines = []string{}
	bytes, err := ioutil.ReadFile(filepath.Clean(buildInfoFile))
	if err != nil {
		log.Printf("failed loading file : %v", err)
	} else {
		buildInfoLines = strings.Split(strings.TrimSpace(string(bytes)), "\n")
	}

	res := [][]string{}
	res = append(res, []string{"Application name", "PromptRec"})

	// build timestamp
	res = append(res, getBuildInfo("Build timestamp", buildInfoLines, "n/a"))
	user, err := user.Current()
	if err != nil {
		log.Printf("failed reading system user name : %v", err)
	}

	// built by username
	res = append(res, getBuildInfo("Built by", buildInfoLines, user.Name))

	// git commit id and branch
	commitIDLong, err := exec.Command("git", "rev-parse", "HEAD").Output()
	var commitIDAndBranch = "unknown"
	if err != nil {
		log.Printf("couldn't retrieve git commit hash: %v", err)
	} else {
		commitID := string([]rune(string(commitIDLong)[0:7]))
		branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err != nil {
			log.Printf("couldn't retrieve git branch: %v", err)
		} else {
			commitIDAndBranch = fmt.Sprintf("%s on %s", commitID, strings.TrimSpace(string(branch)))
		}
	}
	res = append(res, getBuildInfo("Git commit", buildInfoLines, commitIDAndBranch))

	// git release tag
	releaseTag, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		log.Printf("couldn't retrieve git release/tag: %v", err)
		releaseTag = []byte("unknown")
	}
	res = append(res, getBuildInfo("Release", buildInfoLines, string(releaseTag)))

	res = append(res, []string{"Started", start.Format(timestampFmt)})
	res = append(res, []string{"Host", host})
	//res = append(res, []string{"Port", port})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>%s</title></head><body>", "STTS PromptRec: About")
	fmt.Fprintf(w, "<table><tbody>")
	for _, l := range res {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>\n", l[0], l[1])
	}
	fmt.Fprintf(w, "</tbody></table>")
	fmt.Fprintf(w, "</body></html>")
}

func userSessionIsLocked(project string, session string, user string) bool {
	dir := filepath.Join(projectsDir, project, session, user)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	lf := filepath.Join(dir, lockFileName)
	if _, err := os.Stat(lf); os.IsNotExist(err) {
		return false
	}
	return true
}

func openSession0(project string, session string, user string) error {
	dir := filepath.Join(projectsDir, project, session, user)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0750)
		if err != nil {
			return fmt.Errorf("failed to create session dir : %v", err)
		}
		log.Printf("Created session dir %s\n", dir)
	}
	lf := filepath.Join(dir, lockFileName)
	if _, err := os.Stat(lf); !os.IsNotExist(err) {
		return fmt.Errorf("session is locked by another user")
	}
	err := ioutil.WriteFile(lf, []byte("locked"), 0644)
	if err != nil {
		return fmt.Errorf("failed to create lock file : %v", err)
	}
	return nil
}

func closeSession0(project string, session string, user string) error {
	dir := filepath.Join(projectsDir, project, session, user)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("session doesn't exist")
	}
	lf := filepath.Join(dir, lockFileName)
	if _, err := os.Stat(lf); os.IsNotExist(err) {
		return fmt.Errorf("session is not open (missing lock file)")
	}

	// remove lock file
	var err = os.Remove(lf)
	if err != nil {
		return fmt.Errorf("failed to delete lock file : %v", err)
	}
	log.Printf("Removed session lock file %s", lf)

	// remove user folder if empty
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		err = os.Remove(dir)
		if err != nil {
			return err
		}
		log.Printf("Removed empty user folder %s", dir)
	}

	// remove session folder if empty
	sDir := filepath.Join(projectsDir, project, session)
	files, err = ioutil.ReadDir(sDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		err = os.Remove(sDir)
		if err != nil {
			return err
		}
		log.Printf("Removed empty session folder %s", sDir)
	}
	return nil
}

func init() { // called once only (golang built-in feature)
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		err := os.MkdirAll(projectsDir, 0750)
		if err != nil {
			log.Fatalf("failed to create projects dir : %v\n", err)
		}
		log.Printf("Created projects dir %s", projectsDir)
	}
	projects, err := listProjects0()
	if err != nil {
		log.Fatalf("failed to list project : %v\n", err)
	}
	projectSinPlu := "project"
	if len(projects) != 1 {
		projectSinPlu += "s"
	}
	pNames := []string{}
	for _, p := range projects {
		pNames = append(pNames, fmt.Sprintf("%s (%v)", p.Name, p.Size))
	}
	if len(projects) == 0 {
		log.Printf("Found %d %s", len(projects), projectSinPlu)
	} else {
		log.Printf("Found %d %s: %s", len(projects), projectSinPlu, strings.Join(pNames, ", "))
	}
}

func readFile(fName string) (string, error) {
	bytes, err := ioutil.ReadFile(filepath.Clean(fName))
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(bytes), "\n"), nil
}

type utterance struct {
	id           string
	index        int // starts at 1, for progress
	text         string
	instructions string
	audio        []byte
}

func readProjectText(project string) ([]utterance, error) {
	res := []utterance{}
	hasInstructions := false
	s, err := readFile(filepath.Join(projectsDir, project, textFileName))
	if err != nil {
		return []utterance{}, fmt.Errorf("couldn't read text file : %v", err)
	}
	lineNo := 0
	for _, line := range strings.Split(s, "\n") {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		lineNo++
		fs := strings.Split(line, "\t")
		if len(fs) != 2 && len(fs) != 3 {
			return []utterance{}, fmt.Errorf("expected 2-3 fields in project text file, found %d : %s", len(fs), project)
		}
		id := fs[0]
		text := fs[1]
		instr := ""
		if len(fs) == 3 {
			hasInstructions = true
			instr = fs[2]
		}
		res = append(res, utterance{id: id, text: text, instructions: instr, index: lineNo})
	}
	for _, u := range res {
		if hasInstructions && u.instructions == "" {
			return []utterance{}, fmt.Errorf("project contains a mix of utterances with/without instructions: %s", project)
		}
		if !hasInstructions && u.instructions != "" {
			return []utterance{}, fmt.Errorf("project contains a mix of utterances with/without instructions: %s", project)
		}
	}
	log.Printf("Loaded %d text items for project %s", len(res), project)
	return res, nil
}

func cleanExt(ext string) string {
	return strings.TrimPrefix(ext, ".")
}

func contains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

func readProject(project string) ([]utterance, error) {
	utts, err := readProjectText(project)
	if err != nil {
		return []utterance{}, err
	}
	return utts, nil
}

var ignoreProjectRE = regexp.MustCompile("(?i)(IGNORE|BAK|BKP|BLOCK)")

func hasIgnoreLabel(project os.FileInfo) (bool, error) {
	if ignoreProjectRE.MatchString(project.Name()) {
		return true, nil
	}
	files, err := ioutil.ReadDir(filepath.Join(projectsDir, project.Name()))
	for _, file := range files {
		if ignoreProjectRE.MatchString(file.Name()) {
			return true, nil
		}
	}
	if err != nil {
		err = fmt.Errorf("hasIgnoreLabel: failed to list files in project folder : %v", err)
		return false, err
	}
	return false, nil
}

func listProjects0() ([]ProjectInfoJSON, error) {
	res := []ProjectInfoJSON{}

	files, err := ioutil.ReadDir(projectsDir)
	if err != nil {
		err = fmt.Errorf("listProjects: failed to list project folders : %v", err)
		return res, err
	}
	for _, project := range files {
		if !project.Mode().IsDir() {
			continue
		}
		ignore, err := hasIgnoreLabel(project)
		if err != nil {
			return res, err
		}
		if ignore {
			log.Printf("Skipping project %s", project.Name())
			continue
		}
		utts, err := readProject(project.Name())
		if err != nil {
			log.Printf("Invalid project %s : %v ", project.Name(), err)
		} else {
			res = append(res, ProjectInfoJSON{Name: project.Name(), Size: len(utts)})
		}
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res, nil
}

func listProjects(w http.ResponseWriter, r *http.Request) {
	res, err := listProjects0()
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		httpError(w, msg, "failed to list projects", http.StatusInternalServerError)
		return
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("listProjects: failed to marshal list of strings : %v", err)
		httpError(w, msg, "failed to return list of projects", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(resJSON))
}

func openSession(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	project := params["project"]
	sessionName := params["session"]
	userName := params["user"]

	overwrite := r.URL.Query()["overwrite"][0]

	if userName == "" {
		msg := "No username provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	if sessionName == "" {
		msg := "No sessionname provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	if project == "" {
		msg := "No project provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	sessionPath := fmt.Sprintf("%s/%s/%s", project, sessionName, userName)

	if userSessionIsLocked(project, sessionName, userName) {
		msg := "session is locked by another user"
		log.Println(msg)
		JSONResponse(w, "error", fmt.Sprintf("Couldn't start session %s: %v", sessionPath, msg))
		return
	}

	if overwrite == "" || overwrite == "false" {
		nExists, nTotal, err := userSessionExists(project, sessionName, userName)
		if err != nil {
			msg := "Couldn't check for pre-existing user session"
			httpError(w, msg, msg, http.StatusBadRequest)
			return
		}
		if nExists > 0 {
			fileSinOrPlu := "file"
			if nTotal != 1 {
				fileSinOrPlu = fileSinOrPlu + "s"
			}
			JSONResponse(w, "session_exists", fmt.Sprintf("%d of %d %s", nExists, nTotal, fileSinOrPlu))
			return
		}
	}

	err := openSession0(project, sessionName, userName)
	if err != nil {
		serverMsg := fmt.Sprintf("%v", err)
		log.Println(serverMsg)
		JSONResponse(w, "error", fmt.Sprintf("Couldn't start session %s: %v", sessionPath, err))
		return
	}
	JSONResponse(w, "info", fmt.Sprintf("Opened session %s", sessionPath))
}

func closeSession(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	project := params["project"]
	sessionName := params["session"]
	userName := params["user"]

	if userName == "" {
		msg := "No username provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	if sessionName == "" {
		msg := "No sessionname provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	if project == "" {
		msg := "No project provided"
		httpError(w, msg, msg, http.StatusBadRequest)
		return
	}
	sessionPath := fmt.Sprintf("%s/%s/%s", project, sessionName, userName)
	err := closeSession0(project, sessionName, userName)
	if err != nil {
		serverMsg := fmt.Sprintf("%v", err)
		log.Println(serverMsg)
		JSONResponse(w, "error", fmt.Sprintf("Couldn't close session %s: %v", sessionPath, err))
		return
	}
	JSONResponse(w, "info", fmt.Sprintf("Closed session %s", sessionPath))
}

func checkAudioInput(input SessionAudioInput) error {
	var errMsg []string

	if strings.TrimSpace(input.Project) == "" {
		errMsg = append(errMsg, "no value for 'project'")
	}
	if strings.TrimSpace(input.Session) == "" {
		errMsg = append(errMsg, "no value for 'session'")
	}
	if strings.TrimSpace(input.User) == "" {
		errMsg = append(errMsg, "no value for 'user'")
	}
	if strings.TrimSpace(input.UttID) == "" {
		errMsg = append(errMsg, "no value for 'uttid'")
	}
	if strings.TrimSpace(input.Timestamp) == "" {
		errMsg = append(errMsg, "no value for 'timestamp'")
	}
	if len(input.Audio.Data) == 0 {
		errMsg = append(errMsg, "no 'audio.data'")
	}
	if strings.TrimSpace(input.Audio.FileType) == "" {
		errMsg = append(errMsg, "no value for 'audio.file_type'")
	}

	if len(errMsg) > 0 {
		return fmt.Errorf("missing values in input JSON: %s", strings.Join(errMsg, " : "))
	}

	return nil
}

func saveAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		return
	}
	params := mux.Vars(r)
	verb := params["verb"]
	if verb == "true" {
		saveAudio0(w, r, true)
	} else {
		saveAudio0(w, r, false)
	}

}

// verbMode includes all component results, instead of just one single selected result
func saveAudio0(w http.ResponseWriter, r *http.Request, verbMode bool) {
	var body []byte
	var err error
	if strings.Contains(r.Header.Get("content-type"), "application/x-www-form-urlencoded") {
		for key := range r.Form {
			body = []byte(key)
		}
	} else {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("failed to read request body : %v", err)
			log.Println(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}

	input := SessionAudioInput{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal incoming JSON : %v", err)
		log.Println("[promptrec] " + msg)
		log.Printf("[promptrec] incoming JSON string : %s\n", string(body))
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = checkAudioInput(input)
	if err != nil {
		msg := fmt.Sprintf("incoming JSON was incomplete: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	prettyInput := input
	prettyInput.Audio.Data = ""
	log.Println("[promptrec] SessionAudioInput", prettyInput)

	audioDir := filepath.Join(projectsDir, input.Project, input.Session, input.User)

	audioFile, err := writeAudioFile(audioDir, input)
	if err != nil {
		msg := fmt.Sprintf("failed writing audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	log.Printf("Saved audio file %s", audioFile)

	err = writeJSONInfoFile(audioFile, input)
	if err != nil {
		msg := fmt.Sprintf("failed writing info file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	JSONResponse(w, "info", fmt.Sprintf("Saved audio file %s", input.UttID))

}

var afterUnderscoreRE = regexp.MustCompile("_.*$")

// returns # existing files, # total expected files for this project, and an error (if any)
func userSessionExists(project string, session string, user string) (int, int, error) {
	utts, err := readProject(project)
	if err != nil {
		return 0, 0, err
	}

	dir := filepath.Join(projectsDir, project, session, user)
	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to read user session folder %s : %v", dir, err)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list files for user session %s : %v", dir, err)
	}
	foundAudio := make(map[string]bool)
	foundJSON := make(map[string]bool)
	for _, file := range files {
		fName := file.Name()
		ext := cleanExt(path.Ext(fName))
		baseName := strings.TrimSuffix(filepath.Base(fName), ext)
		baseName = afterUnderscoreRE.ReplaceAllString(baseName, "")
		if ext == audioExt {
			foundAudio[baseName] = true
		} else if ext == "json" {
			foundJSON[baseName] = true
		}
	}
	nFound := 0
	for _, utt := range utts {
		if foundAudio[utt.id] && foundJSON[utt.id] {
			nFound++
		}
	}
	return nFound, len(utts), nil
}

func clearSessions() error {
	if err := clearLockedSessions(); err != nil {
		return fmt.Errorf("couldn't clear locked sessions : %v", err)
	}
	if err := clearEmptySessions(); err != nil {
		return fmt.Errorf("couldn't clear empty sessions : %v", err)
	}
	return nil
}

func clearLockedSessions() error {
	projects, err := ioutil.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to list project folders : %v", err)
	}
	for _, project := range projects {
		if !project.Mode().IsDir() {
			continue
		}
		pDir := filepath.Join(projectsDir, project.Name())
		sessions, err := ioutil.ReadDir(pDir)
		if err != nil {
			return fmt.Errorf("failed to list sessions for project : %v", err)
		}
		for _, session := range sessions {
			if !session.Mode().IsDir() {
				continue
			}
			sDir := filepath.Join(projectsDir, project.Name(), session.Name())
			users, err := ioutil.ReadDir(sDir)
			if err != nil {
				return fmt.Errorf("failed to list users for session : %v", err)
			}
			for _, user := range users {
				if !user.Mode().IsDir() {
					continue
				}
				lf := filepath.Join(projectsDir, project.Name(), session.Name(), user.Name(), lockFileName)
				if _, err := os.Stat(lf); !os.IsNotExist(err) {
					err := os.Remove(lf)
					if err != nil {
						return fmt.Errorf("failed to delete lock file %s : %v", lf, err)
					}
					log.Printf("Deleted lock file %s", lf)
				}
			}
		}
	}
	return nil
}

func clearEmptySessions() error {
	projects, err := ioutil.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to list project folders : %v", err)
	}
	for _, project := range projects {
		if !project.Mode().IsDir() {
			continue
		}
		pDir := filepath.Join(projectsDir, project.Name())
		sessions, err := ioutil.ReadDir(pDir)
		if err != nil {
			return fmt.Errorf("failed to list sessions for project : %v", err)
		}
		for _, session := range sessions {
			if !session.Mode().IsDir() {
				continue
			}
			sDir := filepath.Join(projectsDir, project.Name(), session.Name())
			users, err := ioutil.ReadDir(sDir)
			if err != nil {
				return fmt.Errorf("failed to list users for session : %v", err)
			}
			for _, user := range users {
				uDir := filepath.Join(projectsDir, project.Name(), session.Name(), user.Name())
				if !user.Mode().IsDir() {
					continue
				}
				userSessionFiles, err := ioutil.ReadDir(uDir)
				if err != nil {
					return fmt.Errorf("failed to list files for user : %v", err)
				}
				if len(userSessionFiles) == 0 {
					err = os.Remove(uDir)
					if err != nil {
						return fmt.Errorf("failed to remove empty user dir : %v", err)
					}
					log.Printf("Deleted empty user dir %s", uDir)
				}
			}
			sessionFiles, err := ioutil.ReadDir(sDir)
			if err != nil {
				return fmt.Errorf("failed to list files for user : %v", err)
			}
			if len(sessionFiles) == 0 {
				err = os.Remove(sDir)
				if err != nil {
					return fmt.Errorf("failed to remove empty session dir : %v", err)
				}
				log.Printf("Deleted empty session dir %s", sDir)
			}
		}
	}
	return nil
}

var start = time.Now()
var host = "localhost"
var port = "3092"

func main() {

	if (len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-h")) || (len(os.Args) != 3 && len(os.Args) != 2 && len(os.Args) != 1) {
		fmt.Fprintf(os.Stderr, "Usage:\n promptrec (host) (port)\n\n")
		fmt.Fprintf(os.Stderr, "Sample usage:\n")
		fmt.Fprintf(os.Stderr, " promptrec\n")
		fmt.Fprintf(os.Stderr, " promptrec 127.0.0.1\n")
		fmt.Fprintf(os.Stderr, " promptrec 127.0.0.1 3092\n")
		os.Exit(0)
	}

	if len(os.Args) >= 2 {
		host = os.Args[1]
	}
	if len(os.Args) >= 3 {
		port = os.Args[2]
	}

	if !ffmpegEnabled() {
		log.Printf("Exiting! %s is required! Please install.", ffmpegCmd)
		os.Exit(1)
	}

	if err := clearSessions(); err != nil {
		log.Fatalf("Couldn't clear sessions : %v", err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/projects/list", listProjects).Methods("GET")
	r.HandleFunc("/project/get_prompt/{project}/{utt_id}", getPrompt).Methods("GET")
	r.HandleFunc("/project/get_first_prompt/{project}", getFirstPrompt).Methods("GET")
	r.HandleFunc("/project/get_prev_prompt_id/{project}/{utt_id}", getPrevPromptID).Methods("GET")
	r.HandleFunc("/project/get_next_prompt_id/{project}/{utt_id}", getNextPromptID).Methods("GET")
	r.HandleFunc("/session/start/{project}/{session}/{user}", openSession).Methods("GET")
	r.HandleFunc("/session/close/{project}/{session}/{user}", closeSession).Methods("GET")

	r.HandleFunc("/save/", saveAudio).Methods("POST", "OPTIONS")

	r.HandleFunc("/doc/", generateDoc).Methods("GET")
	r.HandleFunc("/about/", generateAbout).Methods("GET")

	docs := make(map[string]string)
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		if info, ok := docs[t]; ok {
			t = fmt.Sprintf("%s - %s", t, info)
		}
		walkedURLs = append(walkedURLs, t)
		return nil
	})
	if err != nil {
		msg := fmt.Sprintf("failure to walk URLs : %v", err)
		log.Println(msg)
		return
	}

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("promptrec server started on %s:%s", host, port)
	log.Fatal(srv.ListenAndServe())

}
