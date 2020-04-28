var AUDIO = {};

AUDIO.sendBlob = function(audioBlob, project, session, user, text, uttid, timestamp, onLoadEndFunc) {
    //console.log("audio.js : BLOB SIZE: "+ audioBlob.size);
    //console.log("audio.js : BLOB TYPE: "+ audioBlob.type);
    
    // This is a bit backwards, since reader.readAsBinaryString below runs async.
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
	let rez = reader.result //contains the contents of blob as a typed array
	let payload = {
	    project : project,
	    session : session,
	    user: user,
	    audio : { file_type : audioBlob.type, data: btoa(rez)},
	    text : text,
	    uttid : uttid,
	    timestamp: timestamp,
	};
	
	AUDIO.sendJSON(payload, onLoadEndFunc);
    });
    
    reader.readAsBinaryString(audioBlob);
    
    //console.log("audio.js : SENDING BLOB"); 
};

AUDIO.sendJSON = function(payload, onLoadEndFunc) {
    //console.log("PAYLOAD:", payload);

    var xhr = new XMLHttpRequest();
    xhr.open("POST", baseURL + "/save/?verb=true", true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');   
    xhr.onloadend = onLoadEndFunc;    
    xhr.send(JSON.stringify(payload));
};
