package protocol

import (
	"github.com/google/uuid"
)

//Message is a struct for sending messages over websockets
type Message struct {
	Label     string     `json:"label"`
	Error     string     `json:"error,omitempty"`
	Handshake *Handshake `json:"handshake,omitempty"`
}

//AudioConfig contains settings for audio
type AudioConfig struct {
	SampleRate   int    `json:"sample_rate"`
	ChannelCount int    `json:"channel_count"`
	Encoding     string `json:"encoding"`
	BitDepth     int    `json:"bit_depth"`
}

//Handshake is a struct for sending handshakes over websockets
type Handshake struct {
	// sent from client to server
	AudioConfig *AudioConfig `json:"audio_config"`

	UserAgent string `json:"user_agent"`
	Timestamp string `json:"timestamp"`

	UserName string `json:"user_name"`
	Project  string `json:"project"`
	Session  string `json:"session"`

	UUID *uuid.UUID `json:"uuid"` // sent from server to client
}
