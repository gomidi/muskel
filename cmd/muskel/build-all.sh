#! /bin/bash
#BIN_FILE_NAME_PREFIX=$1
BIN_FILE_NAME_PREFIX=muskel
PROJECT_DIR=$2
#PLATFORMS=$(go tool dist list)
PLATFORMS="dragonfly/amd64 freebsd/amd64 netbsd/amd64 openbsd/amd64 darwin/amd64 linux/amd64 linux/arm linux/arm64 windows/386 windows/amd64"
rm -rf ./artifacts/
for PLATFORM in $PLATFORMS; do
        GOOS=${PLATFORM%/*}
        GOARCH=${PLATFORM#*/}
        #FILEPATH="$PROJECT_DIR/artifacts/${GOOS}-${GOARCH}"
        FILEPATH="./artifacts/${GOOS}-${GOARCH}"
        #echo $FILEPATH
        mkdir -p $FILEPATH
        BIN_FILE_NAME="$FILEPATH/${BIN_FILE_NAME_PREFIX}"
        #echo $BIN_FILE_NAME
        if [[ "${GOOS}" == "windows" ]]; then BIN_FILE_NAME="${BIN_FILE_NAME}.exe"; fi
        CMD="GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN_FILE_NAME}"
        #echo $CMD
        echo "${CMD}"
        eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
done
