class StreamingAPI {

    constructor(context, audioSource, isRecordingFunc) {
        this._bytesSent = 0;
        this._isRecordingFunc = isRecordingFunc;
    }

    get byteCount() {
        return this._bytesSent;
    }

    set websocket(audioWS) {
        console.log("set websocket called");
        this._audioWS = audioWS;
        console.log(this._audioWS);
    }

    reset() {
        this._bytesSent = 0;
        this._audioWS = null;
    }

    convertFloat32ToInt16(bufferIn) {
        var l = bufferIn.length;
        var res = new Int16Array(l);     
        const intMax = 32767;
        while (l--) {
            let f = bufferIn[l] * intMax;
            if (f > intMax) 
                f = intMax;
            else if (f < -intMax)
                f = -intMax;
            res[l] = f;
        }
        return res.buffer;
    }

    convertFloat32ToInt16OLD(bufferIn) {
        var l = buffer.length;
        var res = new Int16Array(l);
        while (l--) {
            res[l] = bufferIn[l] * 0xFFFF; //convert to 16 bit
        }
        return res.buffer;
    }

}

class ScriptProcessorAPI extends StreamingAPI {

    constructor(context, audioSource, isRecordingFunc) {
        super(context, audioSource, isRecordingFunc);
        let parent = this;

        let bufferSize = 2048;
        recorder = context.createScriptProcessor(bufferSize, channelCount, channelCount);
        //console.log("ScriptProcessor bufferSize", bufferSize);
        audioSource.connect(recorder)
        recorder.connect(context.destination);

        recorder.onaudioprocess = function (e) {
            if (!parent._isRecordingFunc()) return;
            //console.log("recorder.onaudioprocess", typeof e , e.inputBuffer.getChannelData(0).length);
            const buffer = e.inputBuffer.getChannelData(0);
            const sendable = convertFloat32ToInt16(buffer);
            parent._bytesSent = parent._bytesSent + sendable.byteLength;
            parent._audioWS.send(sendable);
        }
    }

}

class AudioWorkletAPI extends StreamingAPI {

    constructor(context, audioSource, isRecordingFunc) {
        super(context, audioSource, isRecordingFunc);
        let parent = this;

        const connect = async function (context, audioSource) {
            await context.audioWorklet.addModule('processor.js');
            const recorder = new AudioWorkletNode(context, 'recorder-worklet');
            audioSource.connect(recorder).connect(context.destination);

            recorder.port.onmessage = function (e) {
                if (!parent._isRecordingFunc()) return;
                if (e.data.eventType === 'data') {
                    //console.log("recorder.ondata", typeof e.data , e.data.audioBuffer.length);
                    const buffer = e.data.audioBuffer;
                    const sendable = convertFloat32ToInt16(buffer);
                    parent._bytesSent = parent._bytesSent + sendable.byteLength;
                    parent._audioWS.send(sendable);
                }
                if (e.data.eventType === 'stop') {
                    // recording has stopped
                }
            }
        }
        connect(context, audioSource);
    }

}
