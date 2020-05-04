class AudioAPI {

}

class ScriptProcessorAPI extends AudioAPI {

    connect(context, audioSource) {
        let bufferSize = 2048;
        recorder = context.createScriptProcessor(bufferSize, channelCount, channelCount);
        console.log("ScriptProcessor bufferSize", bufferSize);
        audioSource.connect(recorder)
        recorder.connect(context.destination);

        recorder.onaudioprocess = function (e) {
            if (!isRecording()) return;
            //console.log("recorder.onaudioprocess", typeof e , e.inputBuffer.getChannelData(0).length);
            var left = e.inputBuffer.getChannelData(0);
            let sendable = convertFloat32ToInt16(left);
            bytesSent = bytesSent + sendable.byteLength;
            audioWS.send(sendable);
        }
    }

}

class AudioWorkletAPI extends AudioAPI {

    async connect(context, audioSource) {
        await context.audioWorklet.addModule('processor.js');
        const recorder = new AudioWorkletNode(context, 'recorder-worklet');
        audioSource.connect(recorder).connect(context.destination);

        recorder.port.onmessage = function (e) {
            if (e.data.eventType === 'data') {
                //console.log("recorder.ondata", typeof e.data , e.data.audioBuffer.length);
                const audioData = e.data.audioBuffer;
                if (!isRecording()) return;
                var left = e.data.audioBuffer;
                let sendable = convertFloat32ToInt16(left);
                bytesSent = bytesSent + sendable.byteLength;
                audioWS.send(sendable);
            }
            if (e.data.eventType === 'stop') {
                // recording has stopped
            }
        }
    }
}
