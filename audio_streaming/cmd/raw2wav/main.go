package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cryptix/wav"
)

func main() {

	var cmd = "raw2wav"

	//handshake := flag.String("handshake", "", "JSON handshake `file`")
	channelCount := flag.Int("channels", 1, "Number of channels")
	sampleRate := flag.Int("sample_rate", 48000, "Sample rate")
	//encoding := flag.String("encoding", "flac", "Audio encoding")
	significantBits := flag.Int("bits", 16, "significant bits")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <input raw audio file> <output wav file>\n", cmd)
		fmt.Fprintf(os.Stderr, "\nFlags\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()

	args := flag.Args()

	if len(args) != 2 {
		flag.Usage()
	}

	inFile := args[0]
	outFile := args[1]

	bts, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatalf("IO error reading file %s : %v", inFile, err)
	}
	fmt.Fprintf(os.Stderr, "Loaded input file %s\n", inFile)

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("IO error creating file %s : %v", outFile, err)
	}

	// Create the headers for our new file
	meta := wav.File{
		Channels:        uint16(*channelCount),
		SampleRate:      uint32(*sampleRate),
		SignificantBits: uint16(*significantBits),
	}

	writer, err := meta.NewWriter(f)
	if err != nil {
		log.Fatalf("IO error creating writer for file %s: %v", outFile, err)
	}

	// Write to file

	_, err = writer.Write(bts)
	if err != nil {
		log.Fatalf("IO error writing to file %s: %v", outFile, err)
	}

	err = writer.Close()
	if err != nil {
		log.Fatalf("IO error upon closing file s%: %v", outFile, err)
	}
	fmt.Fprintf(os.Stderr, "Saved file %s\n", outFile)
}
