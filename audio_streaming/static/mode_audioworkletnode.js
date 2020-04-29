function initStreamerWithAudioWorklet() {
    if (!navigator.mediaDevices.getUserMedia)
        navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia ||
            navigator.mozGetUserMedia || navigator.msGetUserMedia;

    if (!navigator.mediaDevices.getUserMedia) {
        disableEverything();
        alert('getUserMedia not supported in this browser.');
        return false;
    }


    let audioCtx = window.AudioContext || window.webkitAudioContext;
    context = new audioCtx();
    navigator.mediaDevices.getUserMedia({ audio: true })
        // on success:
        .then(async function (stream) {
            VISUALISER.init(isRecording);
            VISUALISER.connect(stream);

            let audioSource = context.createMediaStreamSource(stream);
            await context.audioWorklet.addModule('processor.js');
            const recorder = new AudioWorkletNode(context, 'recorder-worklet');
            audioSource.connect(recorder).connect(context.destination);

            recorder.port.onmessage = function (e) {
                if (e.data.eventType === 'data') {
                    console.log("recorder.ondata");
                    const audioData = e.data.audioBuffer;
                    if (!isRecording()) return;
                    var left = e.inputBuffer.getChannelData(0);
                    let sendable = convertFloat32ToInt16(left);
                    bytesSent = bytesSent + sendable.byteLength;
                    audioWS.send(sendable);
                }
                if (e.data.eventType === 'stop') {
                    // recording has stopped
                }
            };
            //let time = new Date().getTime(); // ??? 
            //let duration = 1; // ??? 
            //recorder.parameters.get('isRecording').setValueAtTime(1, time);
            //recorder.parameters.get('isRecording').setValueAtTime(0, time + duration);
            //yourSourceNode.start(time); // ??? 
            return true;
        })

        // on error:
        .catch(function (err) {
            console.log("error from getUserMedia", err);
            micDetected = false;
            logMessage("error", "No microphone detected. Please verify that your microphone is properly connected.");
            alert("Couldn't initialize recorder: " + err.message + "\n\nPlease verify that your microphone is properly connected.");
            disableEverything();
            return false;
        });

}

