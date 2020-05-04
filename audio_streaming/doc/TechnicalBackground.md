# Overview

In this document, we describe the background for the `audio_streaming` demo application, available methods for audio streaming and technical challenges with different methods.


## Streaming technologies


### AudioWorklet

For streaming audio with JavaScript only (no external libraries) the Web Audio API and [AudioWorklet](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet) is the ??? ??? method recommended by ??? ???.

TODO more background


**Pros**
* Designed to be lossless
* ...

**Cons**
* Not supported by all browsers
* ...

### ScriptProcessorNode

Before the AudioWorket was introduced into the Web Audio API, the [ScriptProcessorNode](https://developer.mozilla.org/en-US/docs/Web/API/ScriptProcessorNode) could be used for streaming. This part of the Web Audio API has since been deprecated due to some critical design flaws. More information on the motivation behind the move to AudioWorklet can be found here: https://hoch.io/assets/publications/icmc-2018-choi-audioworklet.pdf

Unfortunately, the AudioWorklet is not fully supported by Firefox yet (neither stable version 75 nor beta version 76 work as of April 29th, 2020). It works fine with Google Chrome (version 81) and Opera (version 68).

TODO more background

**Pros**
* Supported by most browsers
* ...

**Cons**
* Not lossless
* ...


### WebRTC

Background TODO

**Pros**
* Supported by most browsers (?)
* ...

**Cons**
* Not lossless
* ...



## Browser settings for audio streaming

There seems to be some settings in the browser(s), that are difficult to control or even access for reading. Examples:

* sample rate: we can read it, but haven't figure out how to change it
* audio encoding: we currently cannot read or set this value in the browser
* channel count: we can set this value, but we are not sure how it's used (may differ between ScriptProcessorNode and AudioWorklet)

We would like to be able to better control these settings, or at the very least find out what they are, so that we can confidently send this information to the server along with the other metadata. This information is needed to save the raw audio data as a specific audio file (wav or similar).
