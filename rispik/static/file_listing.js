//let baseURL 
window.onload = function() { 
    
    let btn = document.getElementById("file_list_button");
    btn.addEventListener('click', function(evt) {


	console.log("file_list_button clicked");
	
	let wsURL = "ws://" + window.location.host + "/ws/list_audio_files_for_user";
	let fileListWS = new WebSocket(wsURL);
	
	fileListWS.onopen = function() {
	    let request = {
		"user": localStorage.getItem("user"),
		"project": localStorage.getItem("project"),
		"session":localStorage.getItem("session")
	    };
	    
	    console.log("sending", JSON.stringify(request));
	    fileListWS.send(JSON.stringify(request));
	    
	};
	fileListWS.onerror = function () {
            logMessage("client", "error", "Couldn't open file listing socket");
	}
	
	fileListWS.onmessage = function(evt) {
	    let resp = JSON.parse(evt.data);
	    console.log("got file info from server", resp);
	};
    });
};


function logMessage(source, title, text, stacktrace) {
    let ts = timestampHHMMSS(new Date());
    console.log("logMessage", ts, source, title, text);
    if (stacktrace !== undefined) {
        //const stack = new Error().stack;
        console.log("logMessage stack", ts, stacktrace.stack);
    }
    let p = document.createElement("p");
    p.innerText = ts + " " + source + " " + title + ": " + text;
    document.getElementById("system_message_list").appendChild(p);
}
