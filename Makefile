.PHONY: all test coverage

all: get build install

get:
	go get ./...

build:
	config build -v --workingdir='./cmd/muskel' --versiondir='.'
	config build --workingdir='./cmd/muskel' --versiondir='.'

release:
	config release --versiondir='.' --package='muskel'
	config build -v --workingdir='./cmd/muskel' --versiondir='.'
	config build --workingdir='./cmd/muskel' --versiondir='.'

test:
	go test ./... -v -coverprofile .coverage.txt
	go tool cover -func .coverage.txt

coverage: test
	go tool cover -html=.coverage.txt