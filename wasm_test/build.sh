GOOS=js GOARCH=wasm go build -o bbolt.wasm main.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
python3 -m http.server
