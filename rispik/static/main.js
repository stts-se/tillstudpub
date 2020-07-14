"use strict";

navigator.getUserMedia = (navigator.getUserMedia ||
    navigator.webkitGetUserMedia ||
    //navigator.mediaDevices.getUserMedia ||
    navigator.mozGetUserMedia ||
    navigator.msGetUserMedia);


function shouldVisualise() {
    return document.getElementById("recstop").disabled === false;
}

document.getElementById("recstart").addEventListener("click", function () {
   console.log("start recording");
    document.getElementById("recstop").disabled = false;
    document.getElementById("recstart").disabled = true;
});

document.getElementById("recstop").addEventListener("click", function () {
    console.log("stopped recording");
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
});

window.onload = async function () {

    //console.log("INITIALISING VISUALISER");

    VISUALISER.init();
    let mediaAccess = navigator.mediaDevices.getUserMedia({ 'audio': true, 'video': false });
    mediaAccess.then(function (stream) {
        console.log("navigator.mediaDevices.getUserMedia was called")
	let audioCtx = window.AudioContext || window.webkitAudioContext;
	let context = new audioCtx();
        VISUALISER.visualise(context, stream, shouldVisualise);
    });
}



///////////////////////////////////////////////////////////////
/// https://github.com/stts-se/tillstudpub/blob/master/audio_streaming/static/main.js

// for streaming, the code is inspired by https://github.com/gabrielpoca/browser-pcm-stream/blob/master/public/recorder.js

const baseURL = window.location.host;
const baseURLWithProtocol = window.location.protocol + '//' + baseURL;// + window.location.pathname.replace(/\/$/g,"");

const keyCodeEnter = 13;
const keyCodeSpace = 32;
const keyCodeEscape = 27;

let mediaAccess;
let recorder;
let micDetected = true;

let context;
const channelCount = 1;
const audioEncoding = "pcm";
const bitDepth = 16;

let streamingAPI

let audioWS;

let user = document.getElementById("user");
let session = document.getElementById("session");
let project = document.getElementById("project");

// START: UTIL
function addClass(element, theClass) {
    element.setAttribute("class", element.getAttribute("class") + " " + theClass);
}

function addStyle(element, theStyle) {
    element.setAttribute("style", element.getAttribute("style") + "; " + theStyle);
}

function removeStyle(element, theStyle) {
    let allS = element.getAttribute("style");
    if (nonEmptyString(allS)) {
        let newSs = [];
        let allSs = allS.split(/ *; +/);
        for (let i = 0; i < allSs.length; i++) {
            const thisS = allSs[i];
            let key = thisS.split(/ *: */)[0].trim();
            if (key !== theStyle) {
                newSs.push(thisS);
            }
        }
        element.setAttribute("style", newSs.join(" "));
    }
}

function removeClass(element, theClass) {
    let allC = element.getAttribute("class");
    if (nonEmptyString(allC)) {
        let newCs = [];
        let allCs = allC.split(/ +/);
        for (let i = 0; i < allCs.length; i++) {
            const thisC = allCs[i];
            if (thisC !== theClass) {
                newCs.push(thisC);
            }
        }
        element.setAttribute("class", newCs.join(" "));
    }
}

function nonEmptyString(s) {
    return s !== undefined && s !== null && s.trim().length > 0;
}

function logMessage(title, text, stacktrace) {
    if (stacktrace !== undefined) {
        //const stack = new Error().stack;
        console.log("logMessage", title, text, stacktrace.stack);
    } else {
        console.log("logMessage", title, text);
    }
    document.getElementById("messages").textContent = title + ": " + text;
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
// END: UTIL


async function initStreamer() {
    console.log("initStreamer called");
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
    await navigator.mediaDevices.getUserMedia({ audio: true })
        // on success:
        .then(async function (stream) {
            VISUALISER.visualise(context, stream, isRecording);
            let audioSource = context.createMediaStreamSource(stream);
            streamingAPI = new AudioWorkletAPI(context, audioSource, bitDepth, isRecording);
            
	    console.log("initStreamer created");
            //streamingAPI.connect(context,audioSource);
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
    console.log("initStreamer function return");
    return true;
}

function isRecording() {
    return document.getElementById("recstop").disabled === false;
}

// TODO: dummy values
function loadUserSettings() { // TEMPLATE
// TODO: dummy values
    //project.value = "rispik"
    //session.value = "test_recording"
    //user.value = "test_user"
    

    project.value = document.getElementById("project").innerText;
    session.value = document.getElementById("session").innerText;
    user.value = document.getElementById("user").innerText;
    
    // TODO: Save settings between sessions
    // let urlParams = new URLSearchParams(window.location.search);
    // if (urlParams.has('project')) {
    //     project.value = urlParams.get("project");
    // }
    // if (urlParams.has('session')) {
    //     session.value = urlParams.get("session");
    // }
    // if (urlParams.has('user')) {
    //     user.value = urlParams.get("user");
    // }

    console.log("Settings");
    console.log("- project:", project.value);
    console.log("- session:", session.value);
    console.log("- user:", user.value);
}

function initSettings() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
}

document.getElementById("recstart").addEventListener("click", async function () {
    // init audio context/recorder first time recstart is clicked (it has to be initialized after user gesture, in order to work in Chrome)
    if (context === undefined || context === null) {
	let ok = await initStreamer();
        if (!ok) {
            return;
        }
    } else {
        //context.resume();
    }

    let wsURL = "ws://" + baseURL + "/ws/register";
    console.log(wsURL);
    audioWS = new WebSocket(wsURL);
    console.log("streamingAPI", streamingAPI);
    streamingAPI.websocket = audioWS;

    audioWS.onopen = function () {
        console.log("websocket opened");

        let handshake = {
            'label': 'handshake',
            'handshake': {
                'audio_config': {
                    'sample_rate': context.sampleRate,
                    'channel_count': channelCount,
                    'encoding': audioEncoding,
                    'bit_depth': streamingAPI.bitDepth,
                },
                'user_agent': navigator.userAgent,
                'timestamp': new Date().toISOString(),
                'user_name': user.value,
                'project': project.value,
                'session': session.value,
            },
        };
        console.log("sending handshake to server", handshake);
        audioWS.send(JSON.stringify(handshake));
    }
    audioWS.onerror = function () {
        logMessage("error", "Couldn't open data socket");
    }
    audioWS.onmessage = function (evt) { // TODO: messages should be sent over a separate message channel
        let resp = JSON.parse(evt.data);
        if (resp["label"] === "handshake") {
            console.log("received streaming handshake from server", resp);
            if (resp["error"] != undefined && resp["error"] != "") {
                logMessage("error", "received error from server handshake: " + resp["error"]);
                recStop();
            }
            else {
                document.getElementById("recstop").disabled = false;
                document.getElementById("recstart").disabled = true;
                console.log("recording started");
            }
        }
    }
    audioWS.onclose = function () {
        console.log("websocket closed");
    }
});

function disableEverything() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = true;
}

document.getElementById("recstop").addEventListener("click", function () {
    recStop();
});

function recStop() {
    //context.suspend();
    if (recorder === null) {
        msg = "Cannot stop recording -- recorder is undefined";
        console.log(msg);
        alert(msg);
    }
    //logMessage("info", "recording stopped");
    console.log("recording stopped");
    console.log("sent " + streamingAPI.byteCount + " bytes in total");

    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;

    //recorder.stop();
    audioWS.close();
    audioWS = null;
    streamingAPI.reset();
};


window.onload = async function () {
    this.loadUserSettings();
    this.initSettings();
    VISUALISER.init();
}

window.onbeforeunload = function () {
    if (isRecording()) {
        return "Are you sure you want to navigate away?";
    }
}

window.onunload = function () {
    // console.log("window.onunload");
    // TODO: can we do something here? close the session?
}

document.addEventListener("keydown", function (evt) { // TEMPLATE FOR KEYBOARD SHORTCUTS
    //console.log("window.keydown", evt.keyCode);
    // if (evt.keyCode === keyCodeSpace) {
    // }
});
