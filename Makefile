GO=$(shell which go)
DOCKER=$(shell which docker)
COMMIT_ID=$(shell git rev-parse HEAD)
NOW=$(shell date +%s)

BINARY_NAME=promadapter
SERVER_PACKAGE=github.com/circonus/promadapter/cmd/server/

all: test build
promadapter: build
build:
	CGO_ENABLED=0 GOOS=linux $(GO) build \
		-ldflags=all='-X "main.commitID=$(COMMIT_ID)" -X "main.buildTime=$(NOW)"' \
		-a -installsuffix cgo -o $(BINARY_NAME) $(SERVER_PACKAGE)
test:
	$(GO) test -v ./... -cover
clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
run:
	$(GO) run -v $(SERVER_PACKAGE)
docker: promadapter
	$(DOCKER) build -t promadapter:$(COMMIT_ID) .
