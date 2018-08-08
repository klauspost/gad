var script = document.createElement('script');
script.src = "./../wasm_exec.js";

script.onload = function () {
    if (!WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    const go = new Go();
    let mod, inst;
    WebAssembly.instantiateStreaming(fetch("./fx.wasm"), go.importObject).then((result) => {
        mod = result.module;
        inst = result.instance;
        run();
    });

    async function run() {
        console.clear();
        await go.run(inst);
        inst = await WebAssembly.instantiate(mod, go.importObject); // reset instance
    }

	var canvas = document.querySelector('#canvas-container');

    canvas.requestFullScreen = canvas.webkitRequestFullScreen || canvas.msRequestFullscreen ||
  		canvas.mozRequestFullScreen || canvas.requestFullScreen;

    canvas.cancelFullScreen = canvas.webkitCancelFullScreen || canvas.msCancelFullscreen ||
        canvas.mozCancelFullScreen || canvas.cancelFullScreen;

};
document.head.appendChild(script);

