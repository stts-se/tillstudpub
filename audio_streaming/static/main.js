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
let speechDetected = false;

let channelCount = 1;

const bigMicOnSrc = "images/mic_red_microphone-3404243_1280.png"

let audioWS;

let username = document.getElementById("username");
let sessionname = document.getElementById("session");
let projectname = document.getElementById("project");

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
	for (let i=0; i<allSs.length;i++) {
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
	for (let i=0; i<allCs.length;i++) {
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
    	projectname.value=urlParams.get("project");
    }
    if (urlParams.has('session')) {
    	sessionname.value=urlParams.get("session");
    }
    if (urlParams.has('user')) {
    	username.value=urlParams.get("user");
    }
}

function initSettings() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
}

function disableAll() {
    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = true;
}

document.getElementById("recstart").addEventListener("click", function() {
    //currentBlob = null;

    let wsURL = "ws://" + baseURL + "/ws/register";
    console.log(wsURL);
    audioWS = new WebSocket(wsURL);

    let context;

    audioWS.onopen = function() {
	console.log("websocket opened");

	let audioContext = window.AudioContext || window.webkitAudioContext;
	context = new audioContext();
	
	let handshake = {
	    'label': 'handshake',
	    'handshake': {
		'sample_rate': context.sampleRate,
		'channel_count': channelCount,
		'encoding': defaultEncoding(),
		//'user': user, ... etc
		'user_agent' : navigator.userAgent,
		'timestamp': new Date().toISOString(),
		'user_name': username.value,
		'project': project.value,
		'session': session.value,
	    },
	};
	console.log("sending handshake to server", handshake);
	audioWS.send(JSON.stringify(handshake));
    }
    audioWS.onerror = function() {
	logMessage("error","Couldn't open data socket");
    }
    audioWS.onmessage = function(evt) { // TODO: messages should be sent over a separate message channel
	let resp = JSON.parse(evt.data);
	if (resp["label"] === "handshake") {
		console.log("received streaming handshake from server", resp);
	    if (resp["error"] != undefined && resp["error"] != "") {
		logMessage("error", "received error from server handshake: " + resp["error"]);
		recStop();
	    }
	    else {
		startStreaming(context);
	    }
	}
    }
    audioWS.onclose = function() {
	console.log("websocket closed");
    }
    
    document.getElementById("audiofeedbacktextspan").innerText = "";
    document.getElementById("bigmic").src = bigMicOnSrc;

    //logMessage("info", "recording started");
    console.log("recording started");

    document.getElementById("recstop").disabled = false;
    document.getElementById("recstart").disabled = true;
    // recorder.start();
});

function startStreaming(context) {
    //console.log("supported constraints", navigator.mediaDevices.getSupportedConstraints());
    let constraints = {audio: true};

    navigator.mediaDevices.getUserMedia(constraints)
    // on success:
	.then(function(stream) {
	    
	    VISUALISER.init(isRecording);		
	    VISUALISER.connect(stream);
	    
	    let audioInput = context.createMediaStreamSource(stream);
	    var bufferSize = 1024;
	    recorder = context.createScriptProcessor(bufferSize, channelCount, channelCount);
	    audioInput.connect(recorder)
	    recorder.connect(context.destination);
	    
	    recorder.onaudioprocess = function(e){
		//console.log("recorder.onaudioprocess");
		if(!isRecording()) return;
		//logMessage("info", "recording");
		var left = e.inputBuffer.getChannelData(0);
		let sendable = convertoFloat32ToInt16(left);
		//console.log("streaming " + sendable.byteLength + " bytes of audio");
		audioWS.send(sendable);
	    }

	})

    // on error:
	.catch(function(err) {
    	    console.log("error from getUserMedia", err);
    	    micDetected = false;
    	    logMessage("error", "No microphone detected. Please verify that your microphone is properly connected.");
    	    alert("Couldn't initialize recorder: " + err.message + "\n\nPlease verify that your microphone is properly connected.");
	});

}

document.getElementById("recstop").addEventListener("click", function() {
    recStop();
});
    
function recStop() {						    
    if (recorder === null) {
	msg = "Cannot record -- recorder is undefined";
	console.log(msg);
	alert(msg);
    }
    //logMessage("info", "recording stopped");
    console.log("recording stopped");

    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
    document.getElementById("audiofeedbacktextspan").innerText = "";
    document.getElementById("bigmic").src = "";
    //recorder.stop();
    audioWS.close();
    audioWS = null;
};


function convertoFloat32ToInt16(buffer) {
    var l = buffer.length;
    var buf = new Int16Array(l)

    while (l--) {
	buf[l] = buffer[l]*0xFFFF; //convert to 16 bit
    }
    return buf.buffer
}


function defaultEncoding() {
    let browser = navigator.userAgent;
    if (browser.toLowerCase().indexOf("chrome") != -1) {
	return "flac";
    } else if (browser.indexOf("mozilla") != -1 || browser.indexOf("firefox") != -1) {
	return "pcm";
    }
    return "pcm";
}

window.onload = async function () {
    
    loadUserSettings();
    initSettings();

    VISUALISER.init(isRecording);

    if (!navigator.mediaDevices.getUserMedia)
        navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia ||
        navigator.mozGetUserMedia || navigator.msGetUserMedia;
    
    if (!navigator.mediaDevices.getUserMedia) {
	disableEverything();
	alert('getUserMedia not supported in this browser.');
    }

}

window.onbeforeunload = function() {
    if (isRecording()) {
	return "Are you sure you want to navigate away?";
    }
}

window.onunload = function() {
    // console.log("window.onunload");
    // TODO: can we do something here? close the session?
}

document.addEventListener("keydown", function(evt) { // TEMPLATE FOR KEYBOARD SHORTCUTS
    //console.log("window.keydown", evt.keyCode);
    // if (evt.keyCode === keyCodeSpace) {
    // }
});
