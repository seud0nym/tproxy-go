VERSION ?= `[ -d ".git" ] && git describe --tags --long --dirty || date +%Y.%m.%d-dev`
LDFLAGS=-ldflags "-s -w -X main.appVersion=${VERSION}"
BINARY="tproxy-go"

MAKEFLAGS += --no-print-directory

build: *.go go.*
	go build ${LDFLAGS} -o ${BINARY}
	upx --ultra-brute ${BINARY}

clean:
	rm -f ${BINARY} *.bz2

arm:
	GOOS=linux GOARCH=arm GOARM=5 $(MAKE) build

all:
	GOOS=linux GOARCH=arm GOARM=5 $(MAKE) build
	bzip2 --best --force ${BINARY} && mv ${BINARY}.bz2 ${BINARY}.armv5.bz2
	GOOS=linux GOARCH=arm64 GOARM= $(MAKE) build
	bzip2 --best --force ${BINARY} && mv ${BINARY}.bz2 ${BINARY}.arm64.bz2
	GOOS=linux GOARCH=mips GOARM= $(MAKE) build
	bzip2 --best --force ${BINARY} && mv ${BINARY}.bz2 ${BINARY}.mips.bz2
