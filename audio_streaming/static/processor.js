// Testing code and examples from https://gist.github.com/flpvsk/047140b31c968001dc563998f7440cc1
// but they need to be adapted.

class RecorderWorkletProcessor extends AudioWorkletProcessor {
  static get parameterDescriptors() {
    return [{
      name: 'isRecording',
      defaultValue: 0
    }];
  }

  constructor() {
    super();
    this._bufferSize = 2048;
    console.log("AudioWorkletProcessor bufferSize", this._bufferSize);
    this._buffer = new Float32Array(this._bufferSize);
    this._initBuffer();
  }

  _initBuffer() {
    this._bytesWritten = 0;
  }

  _isBufferEmpty() {
    return this._bytesWritten === 0;
  }

  _isBufferFull() {
    return this._bytesWritten === this._bufferSize;
  }

  _appendToBuffer(value) {
    if (this._isBufferFull()) {
      this._flush();
    }

    this._buffer[this._bytesWritten] = value;
    this._bytesWritten += 1;
  }

  _flush() {
    let buffer = this._buffer;
    if (this._bytesWritten < this._bufferSize) {
      buffer = buffer.slice(0, this._bytesWritten);
    }

    this.port.postMessage({
      eventType: 'data',
      audioBuffer: buffer
    });

    this._initBuffer();
  }

  _recordingStopped() {
    this.port.postMessage({
      eventType: 'stop'
    });
  }

  process(inputs, outputs, parameters) {
    // By default, the node has single input and output.
    const input = inputs[0];

    const channel = input[0]; // TODO: why only use first channel? I can't play the audio using more channels

    //for (let ci = 0; ci < input.length; ci++) {
    //  const channel = input[ci];
      for (let i = 0; i < channel.length; i++) {
        this._appendToBuffer(channel[i]);
      }
   //}

    return true;

  }

}

registerProcessor('recorder-worklet', RecorderWorkletProcessor);
