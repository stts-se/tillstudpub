// The code in the main global scope.
class MyWorkletNode extends AudioWorkletNode {
    constructor(context) {
      super(context, 'my-worklet-processor');
    }
  }
  
  let context = new AudioContext();
  
  context.audioWorklet.addModule('processors.js').then(() => {
    let node = new MyWorkletNode(context);
  });
  