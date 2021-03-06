# visualiser

A javascript library for visualing when the microphone is on/off (i.e., when audio is captured)

Adapted from https://github.com/mdn/voice-change-o-matic/blob/gh-pages/scripts/app.js

Use with the following canvas (adapt sizes and positions if needed):

    <div style="position:relative; height: 120px; width: 500px" class="visualiser-wrapper">
        <canvas style="position:absolute; top:0px; left:0px; width: 100%; height: 120px" class="visualiser"></canvas>
        <span style="padding-top: 10px; padding-bottom: 10px; text-align: center; position:absolute; top:0px; left:0px; width: 100%; height: 120px">
            <image id="visualisermic" style="height: 100px; display: none" src="mic_red_microphone-3404243_1280.png"></image>
        </span>
    </div>

Initialize:
    VISUALISER.init();
    
Create an audio context, and connect to a media stream and a function that is used to turn visualisation on/off:
    VISUALISER.visualise(audioContext, stream, shouldVisualiseFunc);

Audio context: https://developer.mozilla.org/en-US/docs/Web/API/AudioContext
Media stream: https://developer.mozilla.org/en-US/docs/Web/API/MediaStream:


## Demo

To see a simple demo, clone this repository, and open the file `demo.html` in a browser. You can toggle the visualiser by clicking the start/stop buttons.

For ideas on how to use the visualiser in an application, have a look at the sample code in files `demo.html` and `demo.js`.

### Screenshot with visualisation enabled
![](demo_screenshot_mic_on.png?raw=true)

### Screenshot with visualisation disabled
![](demo_screenshot_mic_off.png?raw=true)
