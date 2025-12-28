
GOROOT := $(shell go env GOROOT)

BENCHDIR := benchmarks
TIMESTAMP := $(shell date -u +"%Y%m%dT%H%M%SZ")

GO_TEST_BENCH := go test -bench=. -benchmem -count=6

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


.PHONY: bench
bench:
	@mkdir -p $(BENCHDIR)
	@echo "Running benchmarks..."
	@$(GO_TEST_BENCH) > $(BENCHDIR)/bench_$(TIMESTAMP).txt
	@echo "Benchmark saved to $(BENCHDIR)/bench_$(TIMESTAMP).txt"

.PHONY: compare
compare:
	@latest=$$(ls -1t $(BENCHDIR)/bench_*.txt | head -n 2 | tail -n 1); \
	newest=$$(ls -1t $(BENCHDIR)/bench_*.txt | head -n 1); \
	if [ -n "$$latest" ] && [ -n "$$newest" ]; then \
		echo "Comparing $$latest -> $$newest"; \
		benchstat $$latest $$newest; \
	else \
		echo "Need at least 2 benchmark files to compare"; \
	fi