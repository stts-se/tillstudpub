/**
 * A simple bypass node demo.
 *
 * @class BypassProcessor
 * @extends AudioWorkletProcessor
 */

class BypassProcessor extends AudioWorkletProcessor {
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
  
  convertFloat32ToInt16(buffer) {
    var l = buffer.length;
    var buf = new Int16Array(l)

    while (l--) {
      buf[l] = buffer[l] * 0xFFFF; //convert to 16 bit
    }
    return buf.buffer
  }

  process(inputs, outputs, parameters) {
    const isRecordingValues = parameters.isRecording;

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



  //process(inputs, outputs) {
  //    const input = inputs[0];
  //  const output = outputs[0];
  //
  //  let sendable = this.convertFloat32ToInt16(input);
  //for (let channel = 0; channel < output.length; ++channel) {
  // output[channel].set(input[channel]);
  // }
  //
  //return true;
  //}
  //}

  registerProcessor('bypass-processor', BypassProcessor);
