# Overview

In this document, we describe the background for the `audio_streaming` demo application, available methods for audio streaming and technical challenges with different methods.

# Microphone capture in the web browser

## Recording a complete file before sending to the server

One way to do microphone recording in a web browser is using the JavaScript Web Audio API. You can obtain an audio stream from the microphone using the `getUserMedia()` function. Using a `MediaRecorder`, the audio can be saved as a Blob. When the recording is done, the Blob can be added to an HTML5 audio element, that can be played in the web browser.

The audio file can be sent to a HTTP server as a base64 encoded string, and decoded and saved on the server.

This method works when the user wants to record something, and only when done send it to the server. This could, for example, be a recording tool, where you read aloud and record a manuscript sentence presented in the browser.

How ever, this method is not useful for streaming audio to the server.



## Streaming over a websocket

A websocket can be thought of as a bi-directional HTTP connection. Usually, when a client calls an HTTP server, there is a request from the client, and a response from the server, and that's it. Subsequent calls must establish new connections to the server.

A websocket, on the other hand, is an HTTP connection that may stay open for any period of time, and on which both the client and the server may send data. Since a websocket is an upgraded HTTP connection, is has much of the same characteristics, such as guarantees against package loss, etc.

A difference between a websocket and a normal HTTP connection, is that there is not a well defined protocol for interaction between client and server (such as GET and POST, etc). After a websocket connection has been established, the client and the server may send text or binary data to each other in any form.


## Streaming using WebRTC

WebRTC is a peer-to-peer method for streaming audio and video. Since human cognition is more forgiving to missing samples than to latency, as little lag as possible is more important than being sure that the sound wave is complete and intact --- packet loss is tolerated.

Under the hood, WebRTC takes care of things like echo-canceling and noise reduction.

WebRTC also includes a DataChannel API, that can be used for transporting data losslessly between peers. However, a websocket may do the job equally good, if you do not need peer-to-peer capabilities.

**Pros**
* Supported by most browsers (?)
* ...

**Cons**
* Not "lossless" [TODO wording]
* ...

## Streaming using ScriptProcessorNode

Before the AudioWorket was introduced into the Web Audio API, the [ScriptProcessorNode](https://developer.mozilla.org/en-US/docs/Web/API/ScriptProcessorNode) could be used for streaming. This part of the Web Audio API has since been deprecated due to some critical design flaws. More information on the motivation behind the move to AudioWorklet can be found here: https://hoch.io/assets/publications/icmc-2018-choi-audioworklet.pdf

Unfortunately, the AudioWorklet is not fully supported by Firefox yet (neither stable version 75 nor beta version 76 work as of April 29th, 2020). It works fine with Google Chrome (version 81) and Opera (version 68).

TODO more background

**Pros**
* Supported by most browsers
* ...

**Cons**
* Not "lossless" [TODO wording]
* ...


## Streaming using AudioWorklet

For streaming audio with JavaScript only (no external libraries) the Web Audio API and [AudioWorklet](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet) is the ??? ??? method recommended by ??? ???.

TODO more background


**Pros**
* Designed to be "lossless" [TODO wording]
* ...

**Cons**
* Not supported by all browsers (browser support listed here: https://developer.mozilla.org/en-US/docs/Web/API/AudioWorkletNode)
* ...




## Browser settings for audio and streaming

There seems to be some settings in the browser(s), that are difficult to control or even access for reading. Examples:

* sample rate: we can read it, but haven't figure out how to change it
* audio encoding: we currently cannot read or set this value in the browser
* channel count: we can set this value, but we are not sure how it's used (may differ between ScriptProcessorNode and AudioWorklet)

We would like to be able to better control these settings, or at the very least find out what they are, so that we can confidently send this information to the server along with the other metadata. This information is needed to save the raw audio data as a specific audio file (wav or similar).
