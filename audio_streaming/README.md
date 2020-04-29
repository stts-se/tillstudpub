# Overview

This is a test of streaming user microphone audio from the browser to a server, where the audio is currently saved as a binary "raw" audio file, along with a JSON file, containing the audio parameters needed to play the file.

The files are stored in the "data" directory on the server. Each file is given a unique (UUID) file namne, with the extensions `.raw` and `.json`. The last file created is copied to "latest.raw" and "latest.json", as a convenience for testing.

## Technical description

The current demo version uses a [ScriptProcessorNode](https://developer.mozilla.org/en-US/docs/Web/API/ScriptProcessorNode) to catch the input audio and stream to server.

Since this part of the Web Audio API is deprecated, we are working on an update using [AudioWorkletNode](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorkletNode) instead.

The reason for switching to the TODO ... More information on the motivation behind the AudioWorkletNode can be found here: https://hoch.io/assets/publications/icmc-2018-choi-audioworklet.pdf

The downside of using AudioWorkletNode is that it is not fully supported by Firefox yet. It works with Google Chrome, however.


# Usage

Simple server/client library for testing audio streaming using the MediaRecorder API.

To start the server, run

 `go run . `

Clients:

* Javascript: Point your browser to http://localhost:7651

* `Go` command line client: See folder `gocli`

Recorded audio is saved in the `data` folder. The last recorded file is always saved as `data/latest.raw`. To play a recorded `.raw` file, run `play` with the correct parameters, e.g.

 `play -e signed-integer -r 44100 -b 16 <rawfile>`

See also playraw_example.sh.

Hints on what parameters to use can be found in the `.json` files accompanying each `.raw` file.



