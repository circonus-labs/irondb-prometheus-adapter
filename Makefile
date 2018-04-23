GO=$(shell which go)
DOCKER=$(shell which docker)
COMMIT_ID=$(shell git rev-parse HEAD)

BINARY_NAME=promadapter
SERVER_PACKAGE=github.com/circonus/promadapter/cmd/server/

all: test build
promadapter: build
build:
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o $(BINARY_NAME) -v $(SERVER_PACKAGE)
test:
	$(GO) test -v ./... -cover
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
run:
	$(GO) run -v $(SERVER_PACKAGE)
docker: promadapter
	$(DOCKER) build -t promadapter:$(COMMIT_ID) .
