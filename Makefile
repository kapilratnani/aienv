BINARY   := aienv
GO       := go
GOFLAGS  :=
LDFLAGS  :=

.PHONY: all build clean test test-verbose test-race coverage coverage-html lint fmt vet install

all: build

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build
	$(GO) install $(GOFLAGS) .

clean:
	$(GO) clean
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

test:
	$(GO) test $(GOFLAGS) ./...

test-verbose:
	$(GO) test $(GOFLAGS) -v ./...

test-race:
	$(GO) test $(GOFLAGS) -race ./...

coverage:
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

coverage-html: coverage
	$(GO) tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...
