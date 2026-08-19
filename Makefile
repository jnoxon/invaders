GOOS   := js
GOARCH := wasm
WASM   := main.wasm
GLUE   := wasm_exec.js
PORT   := 8080

.PHONY: all build test vet serve clean

all: build

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(WASM) .
	cp $$(go env GOROOT)/lib/wasm/$(GLUE) $(GLUE)
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
	@echo "serving on http://localhost:$(PORT) (ctrl-c to stop)"
	python3 -m http.server $(PORT)

clean:
	rm -f $(WASM) $(GLUE)
