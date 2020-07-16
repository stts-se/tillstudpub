"use strict";
const baseURL = window.location.host;
const baseURLWithProtocol = window.location.protocol + '//' + baseURL;

window.onload = function () {

	let btn = document.getElementById("file_list_button");
	btn.addEventListener('click', function (evt) {

		let table = document.getElementById("file_listing_table");
		table.innerHTML = '';

		console.log("file_list_button clicked");

		let wsURL = "ws://" + window.location.host + "/ws/list_audio_files_for_user";
		let fileListWS = new WebSocket(wsURL);

		fileListWS.onopen = function () {
			let request = {
				"user": localStorage.getItem("user"),
				"project": localStorage.getItem("project"),
				"session": localStorage.getItem("session")
			};

			console.log("sending", JSON.stringify(request));
			fileListWS.send(JSON.stringify(request));

		};
		fileListWS.onerror = function () {
			// TODO: this function exists in main.js...
			//logMessage("client", "error", "Couldn't open file listing socket");
			console.log("Couldn't open file listing socket");
		}

		fileListWS.onmessage = function (evt) {
			let resp = JSON.parse(evt.data);
			console.log("got file info from server", resp);

			let tr = document.createElement("tr");
			tr.setAttribute("id", resp.uuid);
			tr.setAttribute("title", "uuid: " + resp.uuid);

			let td1 = document.createElement("td");
			td1.innerHTML = resp.project;
			//td1.classList.add("btn");
			tr.appendChild(td1);

			let td2 = document.createElement("td");
			td2.innerHTML = resp.session;
			//td2.classList.add("btn");
			tr.appendChild(td2);

			let td3 = document.createElement("td");
			td3.innerHTML = resp.timestamp;
			//td2.classList.add("btn");
			tr.appendChild(td3);

			let td0 = document.createElement("td");
			let audioEle = document.createElement("audio");
			td0.appendChild(audioEle);
			let playPause = document.createElement("span");
			playPause.innerHTML = "&#x25b6;";
			playPause.style["font-family"] = "times new roman, times, serif";
			td0.appendChild(playPause);
			tr.appendChild(td0);

			audioEle.onended = function() {
				playPause.innerHTML = "&#x25b6;";
			};

			playPause.addEventListener('click', function (evt) {
				if (audioEle.paused || audioEle.ended) {
					playPause.innerHTML = "&#10074;&#10074;";
					let uuid = tr.getAttribute("id");
					getAudio(uuid, audioEle);
					//audioEle.play();
				}
				else {
					playPause.innerHTML = "&#x25b6;";
					audioEle.pause();
				}
			});


			table.appendChild(tr);

		};


	}


	);
	btn.click();
};


function getAudio(uuid, audioEle) {
	console.log(uuid);

	const baseURLWithProtocol = window.location.protocol + '//' + baseURL;
	fetch(baseURLWithProtocol + "/get_audio_for_uuid/" + uuid)
		.then(response => response.text())
		.then(base64 => {
			//let wav = atob(base64);
			//console.log(wav);

			// https://stackoverflow.com/questions/16245767/creating-a-blob-from-a-base64-string-in-javascript#16245768
			let byteCharacters = atob(base64);

			var byteNumbers = new Array(byteCharacters.length);
			for (var i = 0; i < byteCharacters.length; i++) {
				byteNumbers[i] = byteCharacters.charCodeAt(i);
			}
			var byteArray = new Uint8Array(byteNumbers);

			let blob = new Blob([byteArray], { 'type': 'audio/wav' });
			//let audio = document.getElementById("audio_element");
			audioEle.src = URL.createObjectURL(blob);
			console.log("getAudio onloadend")
			audioEle.play();



		}
		)
		.catch(msg => {
			console.log("ERROR", msg);
		});
}

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
