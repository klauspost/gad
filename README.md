# Go After Dark

This is a new web series where we take a look at Go from a different perspective. In this series we drop everything about Go routers, REST, containers, databases and all the boring stuff of everyday Go. 

Instead we will have fun by recreating demo effects from the 90’s. We will mainly focus on graphical effects, but over time we may dive into other topics. 

This is the associated code for each episode.

To watch the episodes go to https://afterdark.klauspost.com/ for a list of episodes. 


# Installing

Install the code in your GOPATH and fetch dependencies using:

```
go get -u github.com/klauspost/gad/...
```

# Running

Go into the directory you want to run and `go run main.go`. For example

```
cd ep01
go run main.go
```

Data, except music, is embedded into the binaries. To re-generate this run `go generate` in
the folder with the `main.go` file.

# Web Assembly

WASM requires Go 1.11. Your `wasm_exec.js` must match the version you use for building.

In your go installation, you find it in `misc/wasm`. Copy it to the repository root.

In each folder you will find a `wasm.cmd` that contains the commands needed to build the webassembly.

The output will be called `fx.wasm` and will be placed in the same folder as `main.go`.

A local server is needed to execute it. In the repository root execute `go run server/main.go`. 
When the server is running, you can see your effect on localhost, for example `http://localhost:8080/ep01`. See the browser console if errors occur.


# License

This code is released under a standard MIT license. See LICENCE file for more information.
