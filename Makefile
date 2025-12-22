
GOROOT := $(shell go env GOROOT)

.PHONY: wasm
wasm:
	env GOOS=js GOARCH=wasm go build -o ants_again.wasm github.com/rafibayer/ants-again
	cp -f $(GOROOT)/lib/wasm/wasm_exec.js .

.PHONY: serve
serve: wasm
	py -m http.server 8080

.PHONY: dotcpu
dotcpu:
	go tool pprof -dot cpu.prof > cpu.dot