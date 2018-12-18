SET GOOS=
SET GOARCH=

go generate

SET GOOS=js
SET GOARCH=wasm
go build -o fx.wasm main.go
