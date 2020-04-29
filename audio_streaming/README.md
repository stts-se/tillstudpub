# Overview

This is a simple proof of concept application for streaming user microphone audio from a client to a server, where the audio is saved as a binary "raw" audio file, along with a JSON file, containing relevant metadata about the recording.


## Technical description

There are two available clients for testing the application:

1. Javascript for browser use
2. A `go` command line client

The client opens a websocket for each recording, and sends a "handshake" message to the server. If the server is up and running, and the handshake is correct and valid, the server reponds with the same handshake message, adding a unique (UUID). Once this handshake is received by the client, the audio stream capture is started, and sent in 1024 or 2048 byte chunks to the server, using the open websocket.

On the receiving end, the server continuosly writes received bytes to a file buffer.

When the user stops the recording, the audio capture is terminated, and the websocket is closed. When the server receives the close message over the websocket, the buffered audio data is saved to disk as a binary "raw" audio file, along with a JSON file containing relevant metadata about the recording, including the audio parameters needed to play the file. The files are stored in the "data" directory on the server. Each file is given their unique (UUID) file namne, with the extensions `.raw` and `.json`. The last files created are copied to "latest.raw" and "latest.json", as a convenience for testing.


## Remaining issues

### Audio capture: ScriptProcessorNode vs. AudioWorkletNode

The current version of the Javascript client uses a [ScriptProcessorNode](https://developer.mozilla.org/en-US/docs/Web/API/ScriptProcessorNode) to capture the input audio as a stream.

Since this part of the Web Audio API is deprecated, we are working on an update using [AudioWorkletNode](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorkletNode) instead.

The ScriptProcessorNode was deprecated due to some critical design flaws. More information on the motivation behind the switch to AudioWorkletNode can be found here: https://hoch.io/assets/publications/icmc-2018-choi-audioworklet.pdf

The downside of using AudioWorkletNode is that it is not fully supported by Firefox yet (we have tested using stable version 75 and beta version 76, and none of these work). It works fine with Google Chrome (81) and Opera (version 68), however.

We are working on a version using AudioWorkletMode for audio streaming, but it's not fully functioning yet.


### Browser settings for audio streaming

There seems to be some "secret" settings in the browser(s), that are difficult to control or even access for reading. This includes for example audio encoding and sample rate.

We would like to be able to control these settings, or at the very least find out what they are, so that we can send this information to the server along with the other metadata. This information is needed to save the raw audio data as a specific audio file (wav or similar).


### Saving audio as non-raw data

For now, we are just saving the audio data as-is, in a "raw" file. If you know what settings were used to record the raw audio data, you can play it back using the `play` command (on Linux), or using the import dialog in Audacity.

In the future, we will use these settings to create a non-raw audio file (wav or similar). We haven't figured out the details about this yet, but it should not be too complicated.



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



## Test AudioWorkletNode

1. Start the server (see above)
2. Point your browser to http://localhost:7651/audioworklet
3. Make sure you have audio in (microphone) and audio out enabled and fully functioning (use headphones if you can, to avoid audio feedback)
4. Click the START button and start talking

If your voice echoes back, your browser supports AudioWorkletNode.

If you cannot hear anything, the AudioWorkletNode probably doesn't work (or there could be something wrong with your audio settings). Some more info is usually found in the Console output. For Firefox, you could get the following error message: `typeError: context.audioWorklet is undefined`.

If you have audio feedback issues, click STOP or reload the page.
