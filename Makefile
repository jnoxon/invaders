GOOS   := js
GOARCH := wasm
WASM   := main.wasm
GLUE   := wasm_exec.js
PORT   := 9090

.PHONY: all build test vet serve clean

all: build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w" -o $(WASM) .
	cp -f $$(go env GOROOT)/lib/wasm/$(GLUE) $(GLUE)
	chmod u+w $(GLUE)
	@echo "built $(WASM) + $(GLUE)"

test:
	@pkgs=$$(go list ./... 2>/dev/null); \
	if [ -z "$$pkgs" ]; then echo "no pure-Go packages to test yet"; \
	else go test $$pkgs; fi

vet:
	@pkgs=$$(go list ./... 2>/dev/null); \
	if [ -z "$$pkgs" ]; then echo "no pure-Go packages to vet yet"; \
	else go vet $$pkgs; fi

serve:
	@echo "serving on http://0.0.0.0:$(PORT) (ctrl-c to stop)"
	python3 -m http.server $(PORT) --bind 0.0.0.0

clean:
	rm -f $(WASM) $(GLUE)
