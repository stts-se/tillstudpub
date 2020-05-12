"use strict";

/** 
ADAPTED FROM: https://github.com/mdn/voice-change-o-matic/blob/gh-pages/scripts/app.js

Use with the following canvas (adapt sizes and positions if needed):
    <div style="position:relative; height: 120px; width: 500px" class="visualiser-wrapper">
        <canvas style="position:absolute; top:0px; left:0px; width: 100%; height: 120px" class="visualiser"></canvas>
        <span style="padding-top: 10px; padding-bottom: 10px; text-align: center; position:absolute; top:0px; left:0px; width: 100%; height: 120px">
            <image id="visualisermic" style="height: 100px" src=""></image>
        </span>
    </div>

Initialize:
    VISUALISER.init();
    
Connect to a media stream and a function that is used to turn visualisation on/off:
    VISUALISER.visualise(stream, shouldVisualiseFunc);

Media stream: https://developer.mozilla.org/en-US/docs/Web/API/MediaStream:

*/

var VISUALISER = {};

const visualiserMicOnSrc = "images/mic_red_microphone-3404243_1280.png"

VISUALISER.init = function () {
    // set up canvas context for visualizer    
    VISUALISER.canvas = document.querySelector('.visualiser');
    VISUALISER.canvasCtx = VISUALISER.canvas.getContext("2d");

    VISUALISER.updateCanvasSize();

    // draw the black rectangle
    VISUALISER.canvasCtx.fillStyle = 'rgb(0, 0, 0)';
    VISUALISER.canvasCtx.fillRect(0, 0, VISUALISER.innerWidth, VISUALISER.innerHeight);
}

VISUALISER.updateCanvasSize = function () {
    VISUALISER.intendedWidth = document.querySelector('.visualiser-wrapper').clientWidth;
    VISUALISER.canvas.setAttribute('width', VISUALISER.intendedWidth / 2);
    VISUALISER.innerWidth = VISUALISER.canvas.width;
    VISUALISER.innerHeight = VISUALISER.canvas.height;
}


VISUALISER.visualise = function (audioContext, stream, shouldVisualiseFunc) {
    VISUALISER.audioCtx = audioContext;
    VISUALISER.analyser = VISUALISER.audioCtx.createAnalyser();
    VISUALISER.analyser.minDecibels = -90;
    VISUALISER.analyser.maxDecibels = -10;
    VISUALISER.analyser.smoothingTimeConstant = 0.85;

    let source = VISUALISER.audioCtx.createMediaStreamSource(stream);
    source.connect(VISUALISER.analyser);
    VISUALISER.innerWidth = VISUALISER.canvas.width;
    VISUALISER.innerHeight = VISUALISER.canvas.height;

    VISUALISER.analyser.fftSize = 256;
    let bufferLengthAlt = VISUALISER.analyser.frequencyBinCount;
    let dataArrayAlt = new Uint8Array(bufferLengthAlt);

    VISUALISER.canvasCtx.clearRect(0, 0, VISUALISER.innerWidth, VISUALISER.innerHeight);

    let draw = function () {
        let drawVisual = requestAnimationFrame(draw);
        VISUALISER.analyser.getByteFrequencyData(dataArrayAlt);
        VISUALISER.canvasCtx.fillStyle = 'rgb(0, 0, 0)';
        VISUALISER.canvasCtx.fillRect(0, 0, VISUALISER.innerWidth, VISUALISER.innerHeight);

        let barWidth = (VISUALISER.innerWidth / bufferLengthAlt) * 2.5;
        let barHeight;
        let x = 0;

        if (shouldVisualiseFunc()) {
            document.getElementById("visualisermic").src = visualiserMicOnSrc;
            for (let i = 0; i < bufferLengthAlt; i++) {
                barHeight = dataArrayAlt[i];

                VISUALISER.canvasCtx.fillStyle = 'rgb(' + (barHeight + 100) + ',50,50)';
                VISUALISER.canvasCtx.fillRect(x, VISUALISER.innerHeight - barHeight / 2, barWidth, barHeight / 2);

                x += barWidth + 1;
            }
        } else {
            document.getElementById("visualisermic").src = "";
        }

    };

    draw();
}
