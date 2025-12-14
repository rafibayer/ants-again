
GOROOT := $(shell go env GOROOT)


wasm:
	env GOOS=js GOARCH=wasm go build -o ants_again.wasm github.com/rafibayer/ants-again
	cp -f $(GOROOT)/lib/wasm/wasm_exec.js .

dotcpu:
	go tool pprof -dot cpu.prof > cpu.dot