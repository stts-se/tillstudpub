* Overview

This is a test of streaming user microphone audio from the browser to a server, where the audio is currently saved as a binary "raw" audio file, along with a JSON file, containing the audio parameters needed to play the file.

The files are stored in the "data" directory on the server. Each file is given a unique (UUID) file namne, with the extensions `.raw` and `.json`. The last file created is copied to "latest.raw" and "latest.json", as a conveninence for testing.

* Usage

Simple server/client library for testing audio streaming using the MediaRecorder API.

To start the server, run

 `go run . `

Clients:

* Javascript: Point your browser to http://localhost:7651

* `Go` command line client: See folder `gocli`

Recorded audio is saved in the `data` folder. The last recorded file is always saved as `data/latest.raw`. To play a recorded `.raw` files, run play with the correct parameters, e.g.

 `play -e signed-integer -r 44100 -b 16 <rawfile>`

(See playraw_example.sh.)



