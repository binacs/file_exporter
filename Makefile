BUILD_FLAGS = -ldflags "-X github.com/BinacsLee/file_exporter/version.GitCommit=`git rev-parse HEAD`"

default: clean build

clean:
	rm -rf bin

build:
	go build $(BUILD_FLAGS) -o bin/file_exporter ./cmd