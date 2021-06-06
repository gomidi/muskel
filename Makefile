.PHONY: all test coverage

all: get build install

get:
	go get ./...

build:
	cd cmd/muskel && config build --versiondir='../../'

release:
	config release --versiondir='.' --package='muskel'
	cd cmd/muskel && config build -v --versiondir='../../' && config build --versiondir='../../'

test:
	go test ./... -v -coverprofile .coverage.txt
	go tool cover -func .coverage.txt

coverage: test
	go tool cover -html=.coverage.txt
