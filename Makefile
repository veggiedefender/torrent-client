APP = ctc
BUILD_DIR = build
BUILD_TIME = $(shell date +%s)
BUILD_LCOMMIT =$(shell git log --pretty=format:"%s"  | head -n 1)

export CTC_TORRENT_FILES_DIR :=./testdata;
export CTC_DOWNLOAD_FILES_DIR :=./testdata;

.PHONY: build all test clean exec

clean:
	rm -rf ./${BUILD_DIR}/${APP}

test:
	go test ./...

build: clean
	go build -ldflags "-s -w -extldflags -static -X 'github.com/edelars/console-torrent-client/version.BuildTime=${BUILD_TIME}' -X 'github.com/edelars/console-torrent-client/version.Commit=${BUILD_LCOMMIT}'" -o ./${BUILD_DIR}/${APP} ./cmd/main.go

exec: build
	exec  ${BUILD_DIR}/${APP}
