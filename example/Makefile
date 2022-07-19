gitCommit:=$(shell git rev-parse --short HEAD)
buildTime:=$(shell date -R)
branch:=$(shell git symbolic-ref --short HEAD)
GOVERSION:=$(shell go version)

BUILD_NAME:=bin/<appname>
PACKAGENAME:=<packagename>
BUILD_VERSION:=0.1
GOOS:=linux
GOARCH:=amd64
CGO_ENABLED:=0
SOURCE:=cmd/<appname>/main.go

BUILDINFO:=-X ${PACKAGENAME}/pkg/buildinfo.Version=${BUILD_VERSION}
BUILDINFO:=${BUILDINFO} -X ${PACKAGENAME}/pkg/buildinfo.GitCommitId=${gitCommit}
BUILDINFO:=${BUILDINFO} -X ${PACKAGENAME}/pkg/buildinfo.ProjectName=moon
BUILDINFO:=${BUILDINFO} -X '${PACKAGENAME}/pkg/buildinfo.BuildTime=${buildTime}'
BUILDINFO:=${BUILDINFO} -X ${PACKAGENAME}/pkg/buildinfo.Branch=${branch}
BUILDINFO:=${BUILDINFO} -X '${PACKAGENAME}/pkg/buildinfo.GoVersion=${GOVERSION}'
BUILDINFO:=${BUILDINFO} -X '${PACKAGENAME}/pkg/buildinfo.OsArch=${GOOS}/${GOARCH}'
BUILDINFO:=${BUILDINFO} -X '${PACKAGENAME}/pkg/buildinfo.Mode=${MODE}'
LDFLAGS:=-ldflags "-s -w ${BUILDINFO}"

all: build cli

cli:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${CLI_NAME} ${CLI_SOURCE}

.PHONY: build

build: clean
	GOOS=${GOOS} GOARCH=${GOARCH} go build -tags=nomsgpack -o ${BUILD_NAME} ${LDFLAGS} ${SOURCE}
tidy:
	go mod tidy

gen:
	go run gen/gen.go

clean:
	rm -f ${BUILD_NAME}
	rm -f ${CLI_NAME}

dev-image:
	docker build -t imageurl --build-arg MODE=dev .

dev-push: dev-image
	docker push imageurl:latest

image:
	docker build -t imageurl:latest -t imageurl:${BUILD_VERSION} --build-arg MODE=prod .

push: image
	docker push imageurl:latest
	docker push imageurl:${BUILD_VERSION}