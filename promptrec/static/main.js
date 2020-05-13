"use strict";

const baseURL = window.location.protocol + '//' + window.location.host;// + window.location.pathname.replace(/\/$/g,"");

const keyCodeEnter = 13;
const keyCodeSpace = 32;
const keyCodeEscape = 27;
const okStarCode = "&#9733;";
const okStarColor = "green"; // "#ffac33;";
const blueStarCode = "&#9733;";
const greyStarCode = "&#9733;";
const greyStarColor = "grey";
const incompleteStarCode = "&#9733;";
const incompleteStarColor = "red";

let projectname = document.getElementById("projectname");
let sessionname = document.getElementById("sessionname");
let username = document.getElementById("username");
let sessionstart = document.getElementById("sessionstart");
let sessionclose = document.getElementById("sessionclose");

let createProjectTab = document.getElementById("createproject");
let recordSessionTab = document.getElementById("recordsession");

let recaudio = document.getElementById("recaudio");

let noProjectsPlaceholder = "No projects";

let autorecDelay = 10; // (ms) short break between autoplay and autorec

let mediaAccess;
let recorder;
let currentBlob;
let speechDetected = false;

let micDetected = true;

let projects = {};
let status = {};

navigator.getUserMedia = (navigator.getUserMedia ||
                          navigator.webkitGetUserMedia ||
			  //navigator.mediaDevices.getUserMedia ||
                          navigator.mozGetUserMedia ||
                          navigator.msGetUserMedia);


async function loadProjects() {
    let url = baseURL + "/projects/list/"
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    projects = await r.json()
	    projectname.innerHTML = "";
	    let names = [];
	    document.getElementById("projectcount").innerText = "0";
	    if (projects.length > 0) {
		document.getElementById("projectcount").innerText = projects.length;
		for (let i=0; i<projects.length; i++) {
		    let option = document.createElement("option");
		    let set = projects[i];
		    option.value = set.name;
		    option.innerText = set.name + " (" + set.size + ")";
		    names.push(set.name + " (" + set.size + ")");
		    projectname.appendChild(option);
		}
		logMessage("info", "Listed projects: " + names.join(", "));
		sessionname.disabled = false;
		validateRegisterButton();
	    } else {
		let option = document.createElement("option");
		let name = noProjectsPlaceholder;
		option.value = name;
		option.innerText = name;
		projectname.appendChild(option);
		projectname.disabled = true;
	    }
	} else {
	    logMessage("error","Couldn't list projects.");
	}
    });
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

function addClass(element, theClass) {
    element.setAttribute("class", element.getAttribute("class") + " " + theClass);
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

function validateRegisterButton() {
    //console.log("validateRegisterButton","uName:", username.value, "sName:", sessionname.value, "pName:", projectname.value);
    const valid = micDetected && nonEmptyString(username.value) && nonEmptyString(sessionname.value) && !projectname.disabled && nonEmptyString(projectname.value);
    if (valid) {
	sessionstart.disabled = false;
    } else {
	sessionstart.disabled = true;
    }
    if (sessionOpen) {
	sessionclose.disabled = false;
    } else {
	sessionclose.disabled = true;
    }
}


sessionclose.addEventListener("click", function() {
    closeSession();
});

sessionstart.addEventListener("click", function() {
    openSession(projectname.value, sessionname.value, username.value, false);
});

async function getFirstPrompt(pn, id) {
    let url = baseURL + "/project/get_first_prompt/" + pn;
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();
	    if (resp.label === "error") {
		logMessage(resp.label, resp.content);
	    } else {
		displayPrompt(resp, id);
	    }
	} else {
	    console.log("getPrompt failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't start session.");
	}
    });
}

async function getPrompt(pn, id) {
    let url = baseURL + "/project/get_prompt/" + pn + "/" + id;
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();
	    if (resp.label === "error") {
		logMessage(resp.label, resp.content);
	    } else {
		displayPrompt(resp, id);
	    }
	} else {
	    console.log("getPrompt failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't start session.");
	}
    });
}

function displayPrompt(json, id) {
    currentBlob = null;
    speechDetected = false;
    
    // https://stackoverflow.com/questions/16245767/creating-a-blob-from-a-base64-string-in-javascript#16245768
    let byteCharacters = atob(json.audio);  
    
    let byteNumbers = new Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
	byteNumbers[i] = byteCharacters.charCodeAt(i);
    }
    let byteArray = new Uint8Array(byteNumbers);
    
    document.getElementById("promptid").innerText = json.uttid;
    document.getElementById("promptindex").innerText = json.uttindex;
    let leftStars ="";
    console.log
    for (let i=1; i<json.uttindex; i++) {
	let starCode = incompleteStarCode;
	let starColor = incompleteStarColor;
	if (status[i] === "ok") {
	    starCode = okStarCode;
	    starColor = okStarColor;
	}
	leftStars = leftStars + "<span style='color: " + starColor + "' title='Utt index " + i + "'>" + starCode + "</span>";
    }
    let nRightStars = document.getElementById("promptcount").innerHTML - json.uttindex;
    let rightStars = "";
    for (let i=0; i<nRightStars; i++) {
	rightStars = rightStars + greyStarCode;
    }
    document.getElementById("promptstarsdone").innerHTML = leftStars;
    document.getElementById("promptstarsthis").innerHTML = blueStarCode;
    document.getElementById("promptstarsremaining").innerHTML = rightStars;
    if (json.instructions !== "") {
	document.getElementById("promptinstructions").innerText = json.instructions;
	removeClass(document.getElementById("promptinstructionslabel"), "hidden");
	removeClass(document.getElementById("promptinstructions"), "hidden");
    } else {
	document.getElementById("promptinstructions").innerText = "";
	addClass(document.getElementById("promptinstructionslabel"), "hidden");
	addClass(document.getElementById("promptinstructions"), "hidden");
    }
    document.getElementById("prompttext").innerText = json.text;
    document.getElementById("prompttext").title = json.uttid + ": " + json.text;

    document.getElementById("recstart").disabled = false;
    document.getElementById("recstop").disabled = true;
    document.getElementById("recplay").disabled = true;
    document.getElementById("recsave").disabled = true;
    document.getElementById("recsaveandnext").disabled = true;

    enableNavigationButtons(true);
    if (document.getElementById("autorec").checked) {
	sleep(autorecDelay);
	document.getElementById("recstart").click();
    }

}

document.getElementById("nextprompt").addEventListener("click", function() {
    let doNext = function() { getNextPrompt(projectname.value, document.getElementById("promptid").innerText) };
    if (currentBlob !== null && currentBlob.saved !== true) {
	if (confirm("You have unsaved audio. Continue anyway?")) {
	    status[document.getElementById("promptindex").innerText] = "incomplete";
	    doNext();
	}
    } else if (currentBlob === null || currentBlob === undefined) {
	if (confirm("You haven't recorded this prompt. Continue anyway?")) {
	    status[document.getElementById("promptindex").innerText] = "incomplete";
	    doNext();
	}
    } else {
	status[document.getElementById("promptindex").innerText] = "ok";
	doNext();
    }
});

document.getElementById("prevprompt").addEventListener("click", function() {
    let doPrev = function() { getPrevPrompt(projectname.value, document.getElementById("promptid").innerText) };
    if (currentBlob !== null && currentBlob.saved !== true) {
	if (confirm("You have unsaved audio. Continue anyway?")) {
	    status[document.getElementById("promptindex").innerText] = "incomplete";
	    doPrev();
	}
    } else {
	status[document.getElementById("promptindex").innerText] = "ok";
	doPrev();
    }
});

async function getNextPromptID(pn, id) {
    let url = baseURL + "/project/get_next_prompt_id/" + pn + "/" + id;
    let nextID = await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();	    
	    if (resp.label === "id") {
		return resp.content;
	    } else {
		logMessage(resp.label, resp.content);		
	    }
	} else {
	    console.log("getNextPromptID failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't get next prompt.");
	    return -1;
	}
    });
    return nextID;
}

async function getPrevPromptID(pn, id) {
    let url = baseURL + "/project/get_prev_prompt_id/" + pn + "/" + id;
    let nextID = await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();	    
	    if (resp.label === "id") {
		return resp.content;
	    } else {
		logMessage(resp.label, resp.content);		
	    }
	} else {
	    console.log("getPrevPromptID failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't get next prompt.");
	    return -1;
	}
    });
    return nextID;
}

async function getNextPrompt(pn, oldID) {
    let id = await getNextPromptID(pn, oldID);
    if (id < 0) {
	logMessage("info","Already at last prompt");
	return;
    }
    let url = baseURL + "/project/get_prompt/" + pn + "/" + id;
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();	    
	    if (resp.label === "error") {
		logMessage(resp.label, resp.content);
	    } else {
		displayPrompt(resp, id);
	    }
	} else {
	    console.log("getNextPrompt failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't get next prompt.");
	}
    });
}

async function getPrevPrompt(pn, oldID) {
    let id = await getPrevPromptID(pn, oldID);
    if (id < 0) {
	logMessage("info","Already at last prompt");
	return;
    }
    let url = baseURL + "/project/get_prompt/" + pn + "/" + id;
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();	    
	    if (resp.label === "error") {
		logMessage(resp.label, resp.content);
	    } else {
		displayPrompt(resp, id);
	    }
	} else {
	    console.log("getPrevPrompt failed: ", url, r.status, r.statusText);
	    logMessage("error","Couldn't get prev prompt.");
	}
    });
}

function getProjectSize(pn) {
    let items = Object.keys(projects);
    for (let i=0; i<items.length; i++) {
	let p = projects[i];
	if (p.name === pn) {
	    return p.size;
	}
    }
}

async function openSession(pn, sn, un, overwrite) {
    let sessionPath = pn + "/" + sn + "/" + un + "/";
    let url = new URL(baseURL + "/session/start/" + sessionPath);
    let params = {"overwrite": overwrite};
    url.search = new URLSearchParams(params).toString();
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();
	    if (resp.label === "error") {
		alert(resp.content);
		logMessage(resp.label, resp.content);
	    } else if (resp.label === "session_exists") {
		if (confirm("Session data already exists on server (" + resp.content + ").\nOverwrite session?")) {
		    openSession(pn, sn, un, true);
		} else {
		    logMessage("error", "Session already exists on server");
		}
	    } else {
		document.getElementById("recstart").disabled = false;
		sessionstart.disabled = true;
		sessionclose.disabled = false;
		sessionOpen = true;
		document.getElementById("opensessioninfo").innerText = "Open session: " + sessionPath;
		let n = getProjectSize(pn);
		document.getElementById("promptcount").innerText = n;
		document.getElementById("promptstarsdone").innerHTML = "";
		document.getElementById("promptstarsthis").innerHTML = "";
		document.getElementById("promptstarsremaining").innerHTML = "";
		username.disabled = true;
		sessionname.disabled = true;
		projectname.disabled = true;
		removeClass(document.getElementById("main"), "hidden");
		VISUALISER.updateCanvasSize();
		getFirstPrompt(pn);
		//window.focus();
	    }
	} else {
	    console.log("SESSION START FAIL FOR", url, r.status, r.statusText);
	    logMessage("error","Couldn't start session.");
	}
    });

}

async function closeSession() {
    let sessionPath = projectname.value + "/" + sessionname.value + "/" + username.value + "/";
    let url = baseURL + "/session/close/" + sessionPath
    await fetch(url).then(async function(r) {
	if (r.ok) {
	    let resp = await r.json();
	    logMessage(resp.label, resp.content);
	    if (resp.label === "error") {
		console.log("SESSION CLOSE FAIL FOR", url, r.status, r.statusText);
	    } else {
		document.getElementById("recstart").disabled = true;
		document.getElementById("recstop").disabled = true;
		document.getElementById("recplay").disabled = true;
		document.getElementById("recsave").disabled = true;
		document.getElementById("recsaveandnext").disabled = true;
		sessionstart.disabled = false;
		sessionclose.disabled = true;
		sessionOpen = false;
		document.getElementById("opensessioninfo").innerText = "No open session";
		username.disabled = false;
		sessionname.disabled = false;
		projectname.disabled = false;    
		addClass(document.getElementById("main"),"hidden");
		try {
		    recorder.stop(); // TODO: cancel/abort?
		} catch (error) {
		    console.log(error);
		}
	    }
	} else {
	    console.log("SESSION CLOSE FAIL FOR", url, r.status, r.statusText);
	    logMessage("error","Couldn't close session.");
	}
    });

}

projectname.addEventListener("click", validateRegisterButton);
username.addEventListener("click", validateRegisterButton);
sessionname.addEventListener("click", validateRegisterButton);
projectname.addEventListener("change", validateRegisterButton);
username.addEventListener("change", validateRegisterButton);
sessionname.addEventListener("change", validateRegisterButton);
projectname.addEventListener("keyup", validateRegisterButton);
username.addEventListener("keyup", validateRegisterButton);
sessionname.addEventListener("keyup", validateRegisterButton);
username.addEventListener("keypress", function(evt) { if (evt.keyCode == keyCodeEnter) { sessionstart.click(); } });
sessionname.addEventListener("keypress", function(evt) { if (evt.keyCode == keyCodeEnter) { sessionstart.click(); } });
projectname.addEventListener("keypress", function(evt) { if (evt.keyCode == keyCodeEnter) {  sessionstart.click(); } });

document.getElementById("autooptions_toggle_visibility").addEventListener("click", function(evt) {
    if (evt.target.innerText === "Show auto options") {
	removeClass(document.getElementById("autooptionscontent"), "hidden");
	evt.target.innerText = "Hide auto options";
    } else {
	addClass(document.getElementById("autooptionscontent"), "hidden");
	evt.target.innerText = "Show auto options";
    }
});

function enableNavigationButtons(enable) {
    if (enable) {
	if (document.getElementById("promptindex").innerText === document.getElementById("promptcount").innerText) {
	    document.getElementById("nextprompt").disabled = true;
	} else {
	    document.getElementById("nextprompt").disabled = false;
	}
	if (parseInt(document.getElementById("promptindex").innerText) === 1) {
	    document.getElementById("prevprompt").disabled = true;
	} else {
	    document.getElementById("prevprompt").disabled = false;
	}
    } else {
	document.getElementById("nextprompt").disabled = true;
	document.getElementById("prevprompt").disabled = true;
    }
}


document.getElementById("recstart").addEventListener("click", function() {
    if (recorder == null) {
	msg = "Cannot record -- recorder is undefined";
	console.log(msg);
	alert(msg);
    }
    addStyle(document.getElementById("recstart"), "background-color:  #800000");
    document.getElementById("recstart").innerHTML = "&#128309;"
    enableNavigationButtons(false);

    console.log("start recording");

    this.disabled = true;
    document.getElementById("recstop").disabled = false;
    document.getElementById("recplay").disabled = true;
    document.getElementById("recsave").disabled = true;
    document.getElementById("recsaveandnext").disabled = true;
    
    // stopAndSendButton.disabled = false;
    recorder.start();
});


document.getElementById("recstop").addEventListener("click", function() {
    stopRecording();
});

function stopRecording() {						    
    if (recorder === null) {
	msg = "Cannot record -- recorder is undefined";
	console.log(msg);
	alert(msg);
    }
    console.log("stopped recording");

    document.getElementById("recstop").disabled = true;
    document.getElementById("recstart").disabled = false;
    document.getElementById("recplay").disabled = false;
    document.getElementById("recsave").disabled = false;
    document.getElementById("recsaveandnext").disabled = false;
    recorder.stop();
};

document.getElementById("recplay").addEventListener("click", function() {
    document.getElementById("recaudio").play();
});
document.getElementById("recaudio").addEventListener("ended", function() {
    if (document.getElementById("autoreplay").checked && document.getElementById("autosend").checked) {
	sendAndReceiveBlob();
    }
});

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

document.getElementById("recsave").addEventListener("click", function() {
    sendAndReceiveBlob();
});

document.getElementById("recsaveandnext").addEventListener("click", function() {
    if (document.getElementById("nextprompt").disabled) {
	var afterSave = function() {
	    logMessage("info", "At last prompt");
	};
	sendAndReceiveBlob(afterSave);
    } else {
	var afterSave = function() {
	    getNextPrompt(projectname.value, document.getElementById("promptid").innerText);
	};
	sendAndReceiveBlob(afterSave);
    }
});

function sendAndReceiveBlob(onAfterFunc) {
    console.log("sendAndReceiveBlob()");

    if (!speechDetected) {
	status[document.getElementById("promptindex").innerText] = "incomplete";
	if (!confirm("Little or no speech was detected in the recorded audio.\nContinue anyway?")) {
	    return;
	}
    }
    status[document.getElementById("promptindex").innerText] = "ok";
    
    let onLoadEndFunc = function(data) {
	console.log("onLoadEndFunc|STATUS", data.target.status, data.target.statusText);
	if (data.target.status === 200) {
	    let json = JSON.parse(data.target.responseText);
	    logMessage(json.label, json.content);
	    currentBlob.saved = true;
	    document.getElementById("promptstarsthis").innerHTML = "<span style='color: " + okStarColor + "; text-decoration: underline " + okStarColor + "'>" + okStarCode + "</span>";

	    if (document.getElementById("autonext").checked) {
		document.getElementById("nextprompt").click();
	    }
	    if (onAfterFunc !== undefined)
		onAfterFunc();
	} else {
	    logMessage("error", data);
	    return;
	}
    };
    let txt = document.getElementById("prompttext").innerText;
    let uttid = document.getElementById("promptid").innerText;
    let timestamp = new Date().toUTCString();
    
    AUDIO.sendBlob(currentBlob,
		   projectname.value,
		   sessionname.value,
		   username.value,
		   txt,
		   uttid,
		   timestamp,
		   onLoadEndFunc);
}


function updateAudio(blob) {
    console.log("updateAudio()", blob.size);
    currentBlob = blob;
    let audio = document.getElementById('recaudio');
    audio.src = URL.createObjectURL(blob);

    document.getElementById("recstop").disabled = false;
    document.getElementById("recplay").disabled = false;
    document.getElementById("recsave").disabled = false;
    document.getElementById("recsaveandnext").disabled = false;


};

let sessionOpen = false;
function shouldVisualise() {
    return document.getElementById("recstop").disabled === false;
}

document.getElementById("autoall").addEventListener("change", function(evt) {
    setAllAutoOptions(evt.target.checked);
});

function setAllAutoOptions(checked) {
    let autoOptions = document.getElementById("autooptions");
    let inputs = autoOptions.getElementsByTagName("input");
    for (let i=0; i<inputs.length; i++) {
	let ele = inputs[i];
	if (!ele.disabled) 
	    ele.checked = checked;
    }	
}

function loadUserSettings() {

    // TODO: Save settings between sessions

    let urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('project')) {
	projectname.value=urlParams.get("project");
    }
    if (urlParams.has('user')) {
	username.value=urlParams.get("user");
    }
    if (urlParams.has('session')) {
	sessionname.value=urlParams.get("session");
    }

    // auto options
    if (urlParams.has('autoall')) {
	let value = (urlParams.get('autoall') === "true");
	document.getElementById("autoall").checked = value;
	setAllAutoOptions(value);
    }
    if (urlParams.has('autorec')) {
	document.getElementById("autorec").checked = (urlParams.get('autorec') === "true");
    }
    if (urlParams.has('autoreplay')) {
	document.getElementById("autoreplay").checked = (urlParams.get('autoreplay') === "true");
    }
    if (urlParams.has('autosend')) {
	document.getElementById("autosend").checked = (urlParams.get('autosend') === "true");
    }
    if (urlParams.has('autosendonnext')) {
    	document.getElementById("autosendonnext").checked = (urlParams.get('autosendonnext') === "true");
    }
    if (urlParams.has('autostop')) {
	document.getElementById("autostop").checked = (urlParams.get('autostop') === "true");
    }
    if (urlParams.has('autonext')) {
	document.getElementById("autonext").checked = (urlParams.get('autonext') === "true");
    }
}


window.onload = async function () {
    
    await loadProjects();
    loadUserSettings();
    sessionname.focus();
   
    mediaAccess = navigator.mediaDevices.getUserMedia({'audio': true, 'video': false});
    mediaAccess.then(function(stream) {
	console.log("navigator.mediaDevices.getUserMedia was called")

	window.AudioContext = window.AudioContext || window.webkitAudioContext;
	var audioContext = new AudioContext();
	VISUALISER.init();
	VISUALISER.visualise(audioContext, stream, shouldVisualise);

	recorder = new MediaRecorder(stream);
	recorder.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data);
	    if (document.getElementById("autoreplay").checked) {
		document.getElementById("recplay").click();
	    } else if (document.getElementById("autosend").checked) {
		sendAndReceiveBlob();
	    }
	});
	recorder.onstop = function(evt) {
	    console.log("recorder.onstop");
	    removeStyle(document.getElementById("recstart"), "background-color:  #800000");
	    document.getElementById("recstart").innerHTML = "&#128308;"
	    document.getElementById("recstop").disabled = true;
	    enableNavigationButtons(true);
	}

	// SETUP VAD
	var source = audioContext.createMediaStreamSource(stream);
	var options = {
	    source: source,
	    fftSize: 256,
	    voice_stop: function() {
		if (document.getElementById("autostop").checked && !document.getElementById("recstop").disabled) {
	    	    console.log("vad: voice_stop => autostop recording");
		    stopRecording();
		}
	    }, 
	    voice_start: function() {
	    	console.log("vad: voice_start");
		speechDetected = true;
	    }, 
	}; 
	let vad = new VAD(options);

    });
    
    mediaAccess.catch(function(err) {
	console.log("error from getUserMedia", err);
	micDetected = false;
	sessionstart.disabled = true;
	logMessage("error", "No microphone detected. Please verify that your microphone is properly connected.");
	alert("Couldn't initialize recorder: " + err.message + "\n\nPlease verify that your microphone is properly connected.");
    });

    //setupVAD();
    
    validateRegisterButton();
    
}

window.onbeforeunload = function() {
    if (sessionOpen) {
	return "Are you sure you want to navigate away?";
    }
}

window.onunload = function() {
    // console.log("window.onunload");
    // TODO: can we do something here? close the session?
}

document.addEventListener("keydown", function(evt) {
    //console.log("window.keydown", evt.keyCode);
    if (evt.keyCode === 32) { // space
	if (!document.getElementById("recstart").disabled) 
	    document.getElementById("recstart").click();
	else if (!document.getElementById("recstop").disabled) 
	    document.getElementById("recstop").click();
    }
    if (evt.ctrlKey && evt.keyCode === 39) { // right arrow
	document.getElementById("recsaveandnext").click();
    }
    if (evt.ctrlKey && evt.keyCode === 37) { // left arrow
    	document.getElementById("prevprompt").click();
    }
});
