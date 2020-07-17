package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptix/wav"
	"github.com/gorilla/websocket"

	"github.com/stts-se/tillstudpub/rispik/logger"
	"github.com/stts-se/tillstudpub/rispik/protocol"
)

func createWavHeader(audioConfig *protocol.AudioConfig) wav.File {
	// type File struct {
	// 	SampleRate      uint32
	// 	SignificantBits uint16
	// 	Channels        uint16
	// 	NumberOfSamples uint32
	// 	Duration        time.Duration
	// 	AudioFormat     uint16
	// 	SoundSize       uint32
	// 	Canonical       bool
	// 	BytesPerSecond  uint32
	// }

	// Create the headers for our new mono file
	res := wav.File{
		Channels:        uint16(audioConfig.ChannelCount),
		SampleRate:      uint32(audioConfig.SampleRate),
		SignificantBits: uint16(audioConfig.BitDepth), // TODO: Significant Bits = Bit Depth?
	}
	if audioConfig.Encoding == "pcm" || audioConfig.Encoding == "linear16" {
		res.AudioFormat = uint16(1)
	}
	logger.Infof("Created wav header: %#v", res)
	return res
}

func receiveAudioStream(metaData *protocol.AudioMetaData, audioStreamSender *websocket.Conn) {
	defer audioStreamSender.Close() // ??
	var err error

	byteCount := 0

	var rawWriter bufferedFileWriter
	var audioRawFileName string
	if *cfg.saveRaw {
		audioRawFileName = filepath.Join(outputDir, fmt.Sprintf("%s.raw", metaData.UUID.String()))
		rawWriter, err = newBufferedFileWriter(audioRawFileName)
		if err != nil {
			logger.Errorf("Couldn't open raw audio file for writing: %v", err)
			return
		}
	}

	var wavWriter *wav.Writer
	var audioWavFileName string
	if !*cfg.noWav {
		wavHeader := createWavHeader(metaData.AudioConfig)
		audioWavFileName = filepath.Join(outputDir, fmt.Sprintf("%s.wav", metaData.UUID.String()))

		wavFile, err := os.Create(audioWavFileName)
		if err != nil {
			logger.Errorf("Couldn't open wav audio file for writing: %v", err)
			return
		}

		wavWriter, err = wavHeader.NewWriter(wavFile)
		if err != nil {
			logger.Errorf("Couldn't create wav audio file writer: %v", err)
			return
		}
	}

	logger.Infof("Audio stream open for input")

	for {
		mType, bts, err := audioStreamSender.ReadMessage()
		if err != nil {
			// TODO: This seems to happen everytime the websocket is closed, but everything works fine? Are we not sending a proper close?
			msg := fmt.Sprintf("%v", err)
			if strings.Contains(msg, "close 1000") {
				logger.Info("Received implicit close from client")
				break
			}
			logger.Errorf("Could not read from websocket: %v", err)
			break
		}

		if mType == websocket.CloseMessage { // this never happens
			logger.Info("Recevied explicit close from client")
			break
		}

		if mType != websocket.BinaryMessage {
			logger.Info("Skipping non-binary message from websocket")
			continue
		}

		if len(bts) > 0 {
			byteCount += len(bts)
			if *cfg.saveRaw {
				if _, err := rawWriter.writer.Write(bts); err != nil {
					logger.Errorf("Couldn't write raw audio to file: %v", err)
					break
				}
			}

			//--------------------- TODO -------------------------

			if !*cfg.noWav {
				if _, err := wavWriter.Write(bts); err != nil {
					logger.Errorf("Couldn't write wav audio to file: %v", err)
					break
				}
			}

			//logger.Infof("Wrote %v bytes of audio to file", len(bts))
		}
	}

	if !*cfg.noWav {
		if err := wavWriter.Close(); err != nil {
			logger.Errorf("Couldn't close wav writer: %v", err)
			return
		}
		logger.Infof("Saved wav audio file %s", audioWavFileName)

		// copy wav to latest.wav
		latestAudioFileMutex.Lock()
		if err := copyFile(audioWavFileName, latestAudioWavFileName); err != nil {
			logger.Errorf("Couldn't copy wav audio file to %s: %v", latestAudioWavFileName, err)
			return
		}
		logger.Infof("Saved wav audio file %s", latestAudioWavFileName)
		latestAudioFileMutex.Unlock()
	} else {
		os.Remove(latestAudioWavFileName)
	}

	if *cfg.saveRaw {
		// save raw file
		if err := rawWriter.close(); err != nil {
			logger.Errorf("Couldn't save raw audio file: %v", err)
			return
		}
		logger.Infof("Saved raw audio file %s (%v bytes)", audioRawFileName, byteCount)
		// copy wav to latest.wav
		latestAudioFileMutex.Lock()
		if err := copyFile(audioRawFileName, latestAudioRawFileName); err != nil {
			logger.Errorf("Couldn't copy raw audio file to %s: %v", latestAudioRawFileName, err)
			return
		}
		logger.Infof("Saved raw audio file %s", latestAudioRawFileName)
		latestAudioFileMutex.Unlock()
	} else {
		os.Remove(latestAudioRawFileName)
	}

}
