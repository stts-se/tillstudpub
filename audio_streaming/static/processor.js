// Adapted from https://gist.github.com/flpvsk/047140b31c968001dc563998f7440cc1

class RecorderWorkletProcessor extends AudioWorkletProcessor {
    constructor() {
      super();
	this._bufferSize = 2048;
	this._channels = 1;
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

	if (this._channelCount === 1) {
	    const channel = input[0];
	    for (let i = 0; i < channel.length; i++) {
		this._appendToBuffer(channel[i]);
	    }
	} else {
	    for (let ci = 0; ci < input.length; ci++) {
		let channel = input[ci];
		for (let i = 0; i < channel.length; i++) {
		    this._appendToBuffer(channel[i]);
		}
	    }
	}
	

	return true;

    }

}

registerProcessor('recorder-worklet', RecorderWorkletProcessor);
