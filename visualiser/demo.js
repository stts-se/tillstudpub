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
    VISUALISER.init();
    let mediaAccess = navigator.mediaDevices.getUserMedia({ 'audio': true, 'video': false });
    mediaAccess.then(function (stream) {
        console.log("navigator.mediaDevices.getUserMedia was called")
	let audioCtx = window.AudioContext || window.webkitAudioContext;
	let context = new audioCtx();
        VISUALISER.visualise(context, stream, shouldVisualise);
    });


}
