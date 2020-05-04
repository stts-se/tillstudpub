# Overview

This is a simple proof of concept application for streaming user microphone audio from a client to a server. The audio is saved in the server as wav file, a binary "raw" audio file, and a JSON file, containing relevant metadata about the recording.

More information about the application can be found in docs/README.md
A technical background on audio streaming using JavaScript can be find in docs/TechnicalBackground.md

# Usage


To start the server, change directory to `audio_streaming` and run

 `go run cmd/audstr_server/main.go`

If you prefer precompiled executables, use the `audstr_server` command from a published release: https://github.com/stts-se/tillstudpub/releases.

Clients:

* JavaScript: Point your browser to http://localhost:7651

   To use the deprecated ScritpProcessorNode implementation, use http://localhost:7651?mode=scriptprocessornode

* `Go` command line client: See folder `cmd/audstr_client`

You can use the Go client to stream audio output via the sox `play` command:

   `rec -r 48000 -t raw -c 1 - 2> /dev/null  | go run cmd/audstr_client/main.go -channels 1 -sample_rate 48000 -encoding linear16 -host 127.0.0.1 -port 7651 -`

Instead of using `go run`, you can use the `audstr_client` command from a published release: https://github.com/stts-se/tillstudpub/releases.

End the recording with `CTRL-c`.


Recorded audio is saved in the `data` folder in the `audio_streaming` directory. The last recorded file is always saved in raw format as `data/latest.raw`, and with a wav header: `data/latest.wav` (the wav header is work in progress).



