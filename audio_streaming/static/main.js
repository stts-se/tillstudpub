"use strict";

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
var channelCount = 1;

let bytesSent = 0; // for logging purposes

const bigMicOnSrc = "images/mic_red_microphone-3404243_1280.png"

let audioWS;

let user = document.getElementById("user");
let session = document.getElementById("session");
let project = document.getElementById("project");

let initStreamer; // = initStreamerWithScriptProcessor;

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

function isRecording() {
    return document.getElementById("recstop").disabled === false;
}

function loadUserSettings() { // TEMPLATE
    // TODO: Save settings between sessions
    let urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('project')) {
        project.value = urlParams.get("project");
    }
    if (urlParams.has('session')) {
        session.value = urlParams.get("session");
    }
    if (urlParams.has('user')) {
        user.value = urlParams.get("user");
    }
    
    let scriptProcessorNode = "scriptprocessornode";
    let audioWorkletNode =  "audioworkletnode";
    let mode = scriptProcessorNode;
    // streaming mode
    if (urlParams.has('mode')) {
	mode = urlParams.get("mode");
    }
    if (mode.toLowerCase() === scriptProcessorNode) {
	initStreamer = initStreamerWithScriptProcessor();
    } else if (mode.toLowerCase() === audioWorkletNode) {
	initStreamer = initStreamerWithAudioWorklet();
    } else {
	alert("Invalid mode: " + mode + "\nValid modes: " + scriptProcessorNode + " (default) or " + audioWorkletNode);
	disableEverything();
    }

    // log settings
    console.log("project", project.value);
    console.log("session", session.value);
    console.log("user", user.value);
    console.log("mode", mode);
    console.log("options can be set using URL params, e.g. http://localhost:7651/?mode=audioworkletmode to use audioworklet instead of scriptprocessormode");
}

function initSettings() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
}

function disableAll() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = true;
}

document.getElementById("recstart").addEventListener("click", function () {
    // init audio context/recorder first time recstart is clicked (it has to be initialized after user gesture, in order to work in Chrome)
    if (context === undefined || context === null) {
        if (!initStreamer()) {
            return;
        }
    } else {
        //audioContext.resume();
    }

    let wsURL = "ws://" + baseURL + "/ws/register";
    console.log(wsURL);
    audioWS = new WebSocket(wsURL);

    audioWS.onopen = function () {
        console.log("websocket opened");

        let handshake = {
            'label': 'handshake',
            'handshake': {
                'sample_rate': context.sampleRate,
                'channel_count': channelCount,
                'encoding': defaultEncoding(),
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
                document.getElementById("audiofeedbacktextspan").innerText = "";
                document.getElementById("bigmic").src = bigMicOnSrc;
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
    document.getElementById("audiofeedbacktextspan").innerText = "";
    document.getElementById("bigmic").src = "";
}

document.getElementById("recstop").addEventListener("click", function () {
    recStop();
});

function recStop() {
    //audioContext.suspend();
    if (recorder === null) {
        msg = "Cannot stop recording -- recorder is undefined"; // todo: only applicable to ScriptProcessor mode
        console.log(msg);
        alert(msg);
    }
    //logMessage("info", "recording stopped");
    console.log("recording stopped");
    console.log("sent " + bytesSent + " bytes in total");
    bytesSent = 0;

    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
    document.getElementById("audiofeedbacktextspan").innerText = "";
    document.getElementById("bigmic").src = "";
    //recorder.stop();
    audioWS.close();
    audioWS = null;
};


function defaultEncoding() {
    // let browser = navigator.userAgent;
    // if (browser.toLowerCase().indexOf("chrome") != -1) {
    //     return "flac";
    // } else if (browser.indexOf("mozilla") != -1 || browser.indexOf("firefox") != -1) {
    //     return "pcm";
    // }
    return "pcm";
}

window.onload = async function () {
    this.loadUserSettings();
    this.initSettings();
    VISUALISER.init(isRecording);
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
