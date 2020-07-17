"use strict";

const baseURL = window.location.host;
const baseURLWithProtocol = window.location.protocol + '//' + baseURL;// + window.location.pathname.replace(/\/$/g,"");

function timestampYYYYMMDDHHMMSS(date) {
    let yyyy = date.getFullYear();
    let mo = date.getMonth()+1;
    if (mo < 10) mo = "0" + mo;
    let da = date.getDate();
    if (da < 10) da = "0" + da;
    let hh = date.getHours();
    if (hh < 10) hh = "0" + hh;
    let mi = date.getMinutes();
    if (mi < 10) mi = "0" + mi;
    let ss = date.getSeconds();
    if (ss < 10) ss = "0" + ss;
    return yyyy + "-" + mo + "-" + da + " " + hh + ":" + mi + ":" + ss;
}

let messageList = document.getElementById("message_list");

function logMessage(source, title, text, timestamp) {
    if (!timestamp) {
        timestamp = timestampYYYYMMDDHHMMSS(new Date());
    }
    console.log("logMessage", timestamp, source, title, text);
    let p = document.createElement("p");
    p.innerText = timestamp + " " + source + " " + title + ": " + text;
    messageList.appendChild(p);

    let maxCount = document.getElementById("max_count").value;
    let messages = messageList.children;
    if (maxCount && maxCount !== "" && messages.length > maxCount) {
        for (let i=0;i<messages.length-maxCount;i++) {
            messageList.removeChild(messages[i]);
        }
    }
}

document.getElementById("clear").addEventListener("click", function() {
    messageList.innerHTML = '';
});

window.onload = function () {
    if (!typeof (Storage)) {
        alert("Your browser does not support localStorage.");
        return;
    }

    let wsURL = "ws://" + baseURL + "/ws/admin";
    console.log(wsURL);
    let adminWS = new WebSocket(wsURL);
    
    adminWS.onopen = function () {
        logMessage("client", "info", "Opened admin socket");
    }
    adminWS.onerror = function () {
        logMessage("client", "error", "Couldn't open admin socket");
    }
    adminWS.onmessage = function (evt) { 
        let resp = JSON.parse(evt.data);
        logMessage("server", resp.level, resp.message, resp.timestamp);
    }
}
