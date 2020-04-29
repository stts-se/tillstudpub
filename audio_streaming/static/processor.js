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
    console.log("process called");
  
    const isRecordingValues = parameters.isRecording;
    //console.log(isRecordingValues);

    // echo input
    // for (let channel = 0; channel < output[0].length; ++channel) { outputs[0][channel].set(input[0]channel]); }

    for (
      let dataIndex = 0;
      dataIndex < isRecordingValues.length;
      dataIndex++
    ) {
      const shouldRecord = isRecordingValues[dataIndex] === 1;
      if (!shouldRecord && !this._isBufferEmpty()) {
        this._flush();
        this._recordingStopped();
      }

      if (shouldRecord) {
        this._appendToBuffer(inputs[0][0][dataIndex]);
      }
    }

    return true;
  }

}

registerProcessor('recorder-worklet', RecorderWorkletProcessor);

/**
 * A simple bypass node demo.
 *
 * @class BypassProcessor
 * @extends AudioWorkletProcessor
 */

class BypassProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
  }

  process(inputs, outputs) {
    //console.log("process called");
      
    // By default, the node has single input and output.
    const input = inputs[0];
    const output = outputs[0];

    for (let channel = 0; channel < output.length; ++channel) {
      output[channel].set(input[channel]);
    }

    return true;
  }
}

registerProcessor('bypass-processor', BypassProcessor);
