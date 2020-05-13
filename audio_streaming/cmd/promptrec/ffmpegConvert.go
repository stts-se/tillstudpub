package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const ffmpegCmd = "ffmpeg"

func ffmpegEnabled() bool {
	_, pErr := exec.LookPath(ffmpegCmd)
	if pErr != nil {
		log.Printf("recserver.FfmpegEnabled(): External '%s' command does not exist!", ffmpegCmd)
		return false
	}
	return true
}

func ffmpegConvert(inFilePath, outFilePath string, removeInputFile bool) error {

	_, pErr := exec.LookPath(ffmpegCmd)
	if pErr != nil {
		log.Printf("ffmpegConvert failure : %v\n", pErr)
		return fmt.Errorf("ffmpegConvert failed to find the external 'ffmpeg' command : %v", pErr)
	}

	// '-y' means write over if output file already exists
	//without resampling and conversion to mono
	//cmd := exec.Command(ffmpegCmd, "-y", "-i", inFilePath, outFilePath)
	//with resampling and conversion to mono below
	sampleRate := "16000"
	cmd := exec.Command(ffmpegCmd, "-y", "-i", inFilePath, "-ac", "1", "-ar", sampleRate, outFilePath)
	var out bytes.Buffer
	var sterr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &sterr

	err := cmd.Run()
	if err != nil {
		log.Printf("%s\n", sterr.String())
		return fmt.Errorf("ffmpegConvert failed running '%s': %v\n", cmd.Path, err)

	}
	if removeInputFile {
		err := os.Remove(inFilePath)
		if err != nil {
			log.Printf("failed to remove input file : %v\n", err)
		}
	}

	return nil
}
