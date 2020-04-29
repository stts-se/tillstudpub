function convertFloat32ToInt16(buffer) {
    var l = buffer.length;
    var buf = new Int16Array(l)

    while (l--) {
        buf[l] = buffer[l] * 0xFFFF; //convert to 16 bit
    }
    return buf.buffer
}

function initStreamerWithScriptProcessor() {
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
        .then(function (stream) {
            VISUALISER.init(isRecording);
            VISUALISER.connect(stream);

            let audioInput = context.createMediaStreamSource(stream);
            var bufferSize = 1024;
            recorder = context.createScriptProcessor(bufferSize, channelCount, channelCount);
            audioInput.connect(recorder)
            recorder.connect(context.destination);

            recorder.onaudioprocess = function (e) {
                if (!isRecording()) return;
                var left = e.inputBuffer.getChannelData(0);
                let sendable = convertFloat32ToInt16(left);
                bytesSent = bytesSent + sendable.byteLength;
                audioWS.send(sendable);
            }
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
