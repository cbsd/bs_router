#!/bin/sh
set -e

OD=$( dirname `realpath $0` )
cd ${OD}

# Check go install
if [ "$(which go)" == "" ]; then
	echo "error: Go is not installed. Please install go: pkg install -y lang/go"
	exit 1
fi

# Check go version
GOVERS="$(go version | cut -d " " -f 3)"

export GOPATH="${OD}"
export GOBIN="${OD}"
go get
go build -ldflags "$LDFLAGS -extldflags '-static'" -o "$OD/bs_router"
