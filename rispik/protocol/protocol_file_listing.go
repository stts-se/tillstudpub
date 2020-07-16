package protocol

// FileListingRequest is sent by client to request their audio files for specified project/session/user
type FileListingRequest struct {
	User    string `json:"user"`
	Project string `json:"project"`
	Session string `json:"session"`
}

// FileInfo is sent by server to client when listing audio files on server.
// type FileInfo struct {
// 	Date string `json:"date"`
// 	//UUID *uuid.UUID `json:"uuid"`
// 	UUID string `json:"uuid"`
// 	//Length int    `json:"length"`
// }

// // FileInfoList is a list of FileInfo.
// type FileInfoList struct {
// 	Files []FileInfo `json:"files"`
// }
