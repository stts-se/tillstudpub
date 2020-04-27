"use strict";

/** 
SOURCE: https://github.com/mdn/voice-change-o-matic/blob/gh-pages/scripts/app.js

Use with the following HTML canvas:
  <div class="visualiser-wrapper">
     <canvas class="visualiser" width="640" height="100"></canvas> 
  </div>

And in Javascript:
  VISUALISER.init(isRecordingFunc);
  VISUALISER.visualise(stream);

*/

var VISUALISER = {};

VISUALISER.init = function(shouldVisualiseFunc) {
    VISUALISER.audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    VISUALISER.shouldVisualiseFunc = shouldVisualiseFunc;

    //set up the different audio nodes we will use for the app    
    VISUALISER.analyser = VISUALISER.audioCtx.createAnalyser();
    VISUALISER.analyser.minDecibels = -90;
    VISUALISER.analyser.maxDecibels = -10;
    VISUALISER.analyser.smoothingTimeConstant = 0.85;
    
    // set up canvas context for visualizer    
    VISUALISER.canvas = document.querySelector('.visualiser');
    VISUALISER.canvasCtx = VISUALISER.canvas.getContext("2d");

    VISUALISER.updateCanvasSize();

    // draw visualisation area
    VISUALISER.innerWidth = VISUALISER.canvas.width;
    VISUALISER.innerHeight = VISUALISER.canvas.height;
    
    VISUALISER.analyser.fftSize = 256;
    let bufferLengthAlt = VISUALISER.analyser.frequencyBinCount;
    let dataArrayAlt = new Uint8Array(bufferLengthAlt);
    
    VISUALISER.canvasCtx.clearRect(0, 0, VISUALISER.innerWidth, VISUALISER.innerHeight);
    
    let draw = function() {
	let drawVisual = requestAnimationFrame(draw);	
	VISUALISER.analyser.getByteFrequencyData(dataArrayAlt);	
	VISUALISER.canvasCtx.fillStyle = 'rgb(0, 0, 0)';
	VISUALISER.canvasCtx.fillRect(0, 0, VISUALISER.innerWidth, VISUALISER.innerHeight);
	
	let barWidth = (VISUALISER.innerWidth / bufferLengthAlt) * 2.5;
	let barHeight;
	let x = 0;
	
	if (VISUALISER.shouldVisualiseFunc()) { 
	    for(let i = 0; i < bufferLengthAlt; i++) {
		barHeight = dataArrayAlt[i];
		
		VISUALISER.canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
		VISUALISER.canvasCtx.fillRect(x,VISUALISER.innerHeight-barHeight/2,barWidth,barHeight/2);
		
		x += barWidth + 1;
	    }
	};
    };
    
    draw(); 
}


VISUALISER.updateCanvasSize = function() {
    VISUALISER.intendedWidth = document.querySelector('.visualiser-wrapper').clientWidth;
    VISUALISER.canvas.setAttribute('width',VISUALISER.intendedWidth / 2);
    VISUALISER.innerWidth = VISUALISER.canvas.width;
    VISUALISER.innerHeight = VISUALISER.canvas.height;    
}

VISUALISER.connect = function(stream) {
    let source = VISUALISER.audioCtx.createMediaStreamSource(stream);
    source.connect(VISUALISER.analyser);
}
