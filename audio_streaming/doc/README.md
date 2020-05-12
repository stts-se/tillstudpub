# Overview

`audio_streaming` is a simple proof of concept application for streaming user microphone audio from a client to a server. The audio is saved in the server as a binary "raw" audio file, along with a JSON file, containing relevant metadata about the recording.

In this document, we describe the application itself and how it works.

# Technical description of the application

The communication between the client and the server is performed using websockets.

There are two available clients for testing the application:

1. JavaScript for browser use
2. A `go` command line client

The client opens a websocket for each recording, and sends a "handshake" message to the server. If the server is up and running, and the handshake is correct and valid, the server responds with the same handshake message, adding a unique identifier (UUID). Once this handshake is received by the client, the audio stream capture is started. The audio input is processed in chunks of 2048 bytes, converted to 16-bit depth, and sent to the server using the open websocket.

On the receiving end, the server continuously writes received bytes to a file buffer.

When the user stops the recording, the audio capture is terminated, and the websocket is closed. When the server receives the close message over the websocket, the buffered audio data is saved to disk as a wav file, along with a JSON file containing relevant metadata about the recording, including the audio parameters needed to play the file. The files are stored in the `data` directory on the server. Each file is given their unique (UUID) file name, with the extensions `.wav` and `.json`. The last files created are copied to "latest.wav" and "latest.json", as a convenience for testing.


# Supported streaming technologies

More background on the different technologies can be found in the accompanying technical_report.pdf

## ScriptProcessorNode

The ScriptProcessorNode was introduced to meet developers' need to process audio streams in the Web Audio API. Unlike other parts of the Web Audio API, the processing is run in the main thread, which can cause delays. It has since been deprecated and replaced by AudioWorklet (below).


## AudioWorklet

The AudioWorklet has been developed to handle some critical design flaws in the ScriptProcessorNode.

The default settings for this application is to use the AudioWorklet.

The implementation in this demo has been tested with the following browsers:
* Google Chrome - version 81 - supported, working
* Opera - version 68 - supported, working
* Firefox - version 75 - not supported
* Firefox - beta version 76 - will be supported in version 76, currently not working 


## WebRTC
TODO (separate demo)

# Saved audio format

The audio is saved as a wav file. The wav output is still under development, and may be faulty in some cases.

If the wav output seems faulty, the server can be started with an option to save the raw audio data in a `.raw` file. To play a recorded `.raw` file on Linux systems, run `play` with the correct parameters, e.g.

 `play -e signed-integer -r 48000 -b 16 -c 1 <rawfile>`

On Windows, you can for example use the Import function in Audacity.

Hints on what parameters to use can be found in the JSON files accompanying each `.raw` file.


# Audio quality and packet loss

Currently, there seems to be some packet loss for both streaming methods used. For AudioWorklet, the packet loss is small, but still...



# Usage

See audio_streaming/README.md





# Remaining issues

* Investigate how to read and set audio configuration settings in the browser

* Investigate remaining packet loss issues


---

# Testing a browser's AudioWorklet compatibility

Not all browsers support AudioWorklet yet (see above). Here's how to do a quick test:

1. Start the server (see above)
2. Point your browser to http://localhost:7651/audioworklet
3. Make sure you have audio in (microphone) and audio out enabled and fully functioning (use headphones if you can, to avoid audio feedback)
4. Click the START button and start talking

If your voice echoes back, your browser supports AudioWorklet.

If you cannot hear anything, the AudioWorklet probably doesn't work (or there could be something wrong with your audio settings). Some more info is usually found in the Console output. For Firefox, you could get the following error message: `typeError: context.audioWorklet is undefined`.

If you have audio feedback issues, click STOP or reload the page.
