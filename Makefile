OUT := go-deploy
PKG := github.com/dedalusj/go-deploy
VERSION := $(shell git describe --always --long)
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

.PHONY: run build vet lint