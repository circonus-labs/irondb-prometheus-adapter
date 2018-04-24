GO=$(shell which go)
DOCKER=$(shell which docker)
COMMIT_ID=$(shell git rev-parse HEAD)
NOW=$(shell date +%s)

BINARY_NAME=irondb-prometheus-adapter
SERVER_PACKAGE=github.com/circonus-labs/irondb-prometheus-adapter/cmd/server/

all: test build
irondb-prometheus-adapter: build
build:
	$(GO) build -o $(BINARY_NAME) \
		-ldflags=all='-X "main.commitID=development" -X "main.buildTime=$(NOW)"' \
		$(SERVER_PACKAGE)
test:
	$(GO) test -v ./... -cover
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
run:
	$(GO) run -v $(SERVER_PACKAGE)
docker:
	$(GO) clean
	rm -f $(BINARY_NAME)
	CGO_ENABLED=0 GOOS=linux $(GO) build \
		-ldflags=all='-X "main.commitID=$(COMMIT_ID)" -X "main.buildTime=$(NOW)"' \
		-a -installsuffix cgo -o $(BINARY_NAME) $(SERVER_PACKAGE)
	$(DOCKER) build -t irondb-prometheus-adapter:$(COMMIT_ID) .
