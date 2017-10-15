OUT := portainer-endpoints
PKG := github.com/dedalusj/portainer-endpoints
VERSION := $(shell git describe --tags)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: run

build:
	go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION}" ${PKG}

test:
	@go test -short -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

run: build
	./${OUT}

clean:
	-@rm ${OUT}

coverage:
	@go test -short -coverprofile=coverage.out ${PKG_LIST}
	@go tool cover -html coverage.out

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.version=${VERSION}" -o ${OUT} ${PKG}

docker: build-linux
	docker build -t "dedalusj/portainer-endpoints:${VERSION}" .

push: docker
	docker push dedalusj/portainer-endpoints:${VERSION}

.PHONY: run build vet lint
