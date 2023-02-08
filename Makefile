APP = ctc
BUILD_DIR = build

export CTC_TORRENT_FILES_DIR :=./testdata;
export CTC_DOWNLOAD_FILES_DIR :=./testdata;

.PHONY: build all test clean exec

clean:
	rm -rf ./${BUILD_DIR}/${APP}

test:
	go test ./...

build: clean
	go build -ldflags '-s -w -extldflags "-static"' -o ./${BUILD_DIR}/${APP} ./cmd/main.go

exec: build
	exec  ${BUILD_DIR}/${APP}
