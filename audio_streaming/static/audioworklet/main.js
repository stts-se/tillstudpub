
let audioContext;
let button = document.getElementById('button');

// https://bitbucket.org/alvaro_maceda/notoono/src/stackoverflow/src/index.js

async function startProcessFunc(context) {
  let mikeStream = await openMike();
  let mikeNode = context.createMediaStreamSource(mikeStream);

  await context.audioWorklet.addModule('processor.js');
  const bypasser = new AudioWorkletNode(context, 'bypass-processor');
  mikeNode.connect(bypasser).connect(context.destination);
}

window.onload = function () {
  if (!audioContext) {
    audioContext = new AudioContext();
  }
  let isFirstClick = true;
  button.addEventListener("click", function (event) {
    if (button.textContent === 'START') {
      if (isFirstClick) {
        startProcessFunc(audioContext);
        isFirstClick = false;
      }
      audioContext.resume();
      button.textContent = 'STOP';
    } else {
      audioContext.suspend();
      button.textContent = 'START';
    }
  });

}

async function openMike() {

  try {
    let stream = await navigator.mediaDevices.getUserMedia({ "audio": true, "video": false });
    return stream;
  } catch (e) {
    alert('getUserMedia threw exception :' + e);
  }
}
