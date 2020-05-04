# Overview

`audio_streaming` is a simple proof of concept application for streaming user microphone audio from a client to a server. The audio is saved in the server as a binary "raw" audio file, along with a JSON file, containing relevant metadata about the recording.

In this document, we describe the application itself and how it works.

# Technical description of the application

The communication between the client and the server is performed using websockets.

There are two available clients for testing the application:

1. JavaScript for browser use
2. A `go` command line client

The client opens a websocket for each recording, and sends a "handshake" message to the server. If the server is up and running, and the handshake is correct and valid, the server responds with the same handshake message, adding a unique identifier (UUID). Once this handshake is received by the client, the audio stream capture is started, and sent in chunks (typically 1024 or 2048 bytes each) to the server, using the open websocket.

On the receiving end, the server continuously writes received bytes to a file buffer.

When the user stops the recording, the audio capture is terminated, and the websocket is closed. When the server receives the close message over the websocket, the buffered audio data is saved to disk as a binary "raw" audio file, along with a JSON file containing relevant metadata about the recording, including the audio parameters needed to play the file. The files are stored in the `data` directory on the server. Each file is given their unique (UUID) file name, with the extensions `.raw` and `.json`. The last files created are copied to "latest.raw" and "latest.json", as a convenience for testing.


# Supported streaming technologies

More background on the different technologies can be found in the accompanying TechnicalOverview.md

## AudioWorklet

The default settings for this application is to use the AudioWorklet

## ScriptProcessorNode
TODO

## WebRTC
TODO

# Saved audio format

For now, we are saving the audio data as-is, in a "raw" file, and also as a "wav" file.

The wav output is still under development, and may be faulty in some cases. If this happens, the raw file can be used to play the audio. To play a recorded `.raw` file on Linux systems, run `play` with the correct parameters, e.g.

 `play -e signed-integer -r 48000 -b 16 -c 1 <rawfile>`

See also playraw_example.sh.

On Windows, you can for example use the Import function in Audacity.

Hints on what parameters to use can be found in the JSON files accompanying each `.raw` file.

# Distortion

Currently, there is some distortion especially in the beginning of the audio files. This needs to be investigated further. It is possible that this will change once we move over to using AudioWorkletN, but if not, this issue needs to be resolved.



# Usage

See audio_streaming/README.md





# TODO

* Investigate how to read and set audio configuration settings in the browser

* Saving audio data with a wav header -- a first implementation exists, but it needs to be tested further

* Are other options needed for what format the server should use? Once we have a wav file, we can always convert to another format if needed.

* Investigate remaining spike/distortion issues



# Testing a browser's AudioWorklet compatibility

Not all browsers support AudioWorklet yet (see above)

1. Start the server (see above)
2. Point your browser to http://localhost:7651/audioworklet
3. Make sure you have audio in (microphone) and audio out enabled and fully functioning (use headphones if you can, to avoid audio feedback)
4. Click the START button and start talking

If your voice echoes back, your browser supports AudioWorklet.

If you cannot hear anything, the AudioWorklet probably doesn't work (or there could be something wrong with your audio settings). Some more info is usually found in the Console output. For Firefox, you could get the following error message: `typeError: context.audioWorklet is undefined`.

If you have audio feedback issues, click STOP or reload the page.
