# Overview

This is a simple proof of concept application for streaming user microphone audio from a client to a server. The audio is saved in the server as a binary "raw" audio file, along with a JSON file, containing relevant metadata about the recording.

The communication between the client and the server is performed using websockets.

## Audio stream capture

The JavaScript client uses a [AudioWorkletNode](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorkletNode) to capture the input audio as a stream.

There is an option to use a [ScriptProcessorNode](https://developer.mozilla.org/en-US/docs/Web/API/ScriptProcessorNode) instead. This part of the Web Audio API has been deprecated due to some critical design flaws. More information on the motivation behind the move to AudioWorkletNode can be found here: https://hoch.io/assets/publications/icmc-2018-choi-audioworklet.pdf

Unfortunately, the newer AudioWorkletNode is not fully supported by Firefox yet (neither stable version 75 nor beta version 76 work as of April 29th, 2020). It works fine with Google Chrome (version 81) and Opera (version 68).

The ScriptProcessorNode implementation has been kept for testing and compatability issues.


## Browser settings for audio streaming

There seems to be some settings in the browser(s), that are difficult to control or even access for reading. Examples:

* sample rate: we can read it, but haven't figure out how to change it
* audio encoding: we currently cannot read or set this value in the browser
* channel count: we can set this value, but we are not sure how it's used (may differ between ScriptProcessorNode and AudioWorkletNode)

We would like to be able to better control these settings, or at the very least find out what they are, so that we can confidently send this information to the server along with the other metadata. This information is needed to save the raw audio data as a specific audio file (wav or similar).


## Saving audio with wav header

For now, we are just saving the audio data as-is, in a "raw" file. If you know what settings were used to record the raw audio data, you can play it back using the `play` command (on Linux), or using the import dialog in Audacity, for example.

The audio settings (and other metadata) are saved in the JSON file accompanying each raw file.

In the future, we will use these settings to have the server save the audio with a wav header. We haven't figured out the details about this yet, but there are libraries for this, so it should be fairly straightforward.


## Distortion

Currently, there is some distortion especially in the beginning of the audio files. This needs to be investigated further. It is possible that this will change once we move over to using AudioWorkletNode, but if not, this issue needs to be resolved.


# Technical description

There are two available clients for testing the application:

1. JavaScript for browser use
2. A `go` command line client

The client opens a websocket for each recording, and sends a "handshake" message to the server. If the server is up and running, and the handshake is correct and valid, the server responds with the same handshake message, adding a unique identifier (UUID). Once this handshake is received by the client, the audio stream capture is started, and sent in chunks (typically 1024 or 2048 bytes each) to the server, using the open websocket.

On the receiving end, the server continuously writes received bytes to a file buffer.

When the user stops the recording, the audio capture is terminated, and the websocket is closed. When the server receives the close message over the websocket, the buffered audio data is saved to disk as a binary "raw" audio file, along with a JSON file containing relevant metadata about the recording, including the audio parameters needed to play the file. The files are stored in the `data` directory on the server. Each file is given their unique (UUID) file name, with the extensions `.raw` and `.json`. The last files created are copied to "latest.raw" and "latest.json", as a convenience for testing.



# Usage

Simple server/client library for testing audio streaming using the MediaRecorder API.

To start the server, change directory to `audio_streaming` and run

 `go run cmd/audstr_server/main.go`

If you prefer precompiled executables, use the `audstr_server` command from a published release: https://github.com/stts-se/tillstudpub/releases.

Clients:

* JavaScript: Point your browser to http://localhost:7651

   To use the deprecated ScritpProcessorNode implementation, use http://localhost:7651?mode=scriptprocessornode

* `Go` command line client: See folder `cmd/audstr_client`

You can use the Go client to stream audio output via the sox `play` command:

   `rec -r 48000 -t flac -c 2 - 2> /dev/null  | go run cmd/audstr_client/main.go -channels 2 -sample_rate 48000 -encoding flac -host 127.0.0.1 -port 7651 -`

Instead of using `go run`, you can use the `audstr_client` command from a published release: https://github.com/stts-se/tillstudpub/releases.

End the recording with `CTRL-c`.


Recorded audio is saved in the `data` folder in the `audio_streaming` directory. The last recorded file is always saved in raw format as `data/latest.raw`, and with a wav header: `data/latest.wav` (the wav header is work in progress).

# raw files

To play a recorded `.raw` file, run `play` with the correct parameters, e.g.

 `play -e signed-integer -r 48000 -b 16 <rawfile>`


See also playraw_example.sh.

Hints on what parameters to use can be found in the `.json` files accompanying each `.raw` file.


# TODO

1. Upgrade audio stream capture to use AudioWorkletNode (work in progress)

2. Investigate how to read and set audio configuration settings in the browser

3. Saving audio data with a wav header

4. Are other options needed for what format the server should use? Once we have a wav file, we can always convert to another format if needed.

5. Investigate distortion issues, if it doesn't improve after the switch to AudioWorkletNode.



# Testing a browser's AudioWorkletNode compatibility

1. Start the server (see above)
2. Point your browser to http://localhost:7651/audioworklet
3. Make sure you have audio in (microphone) and audio out enabled and fully functioning (use headphones if you can, to avoid audio feedback)
4. Click the START button and start talking

If your voice echoes back, your browser supports AudioWorkletNode.

If you cannot hear anything, the AudioWorkletNode probably doesn't work (or there could be something wrong with your audio settings). Some more info is usually found in the Console output. For Firefox, you could get the following error message: `typeError: context.audioWorklet is undefined`.

If you have audio feedback issues, click STOP or reload the page.
