BIN := bin/kanban

.PHONY: run build test vet fmt tidy clean

run:
	go run ./cmd/kanban

build:
	go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/kanban

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

tidy:
	go mod tidy

clean:
	rm -rf bin
