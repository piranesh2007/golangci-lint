FROM golang:1.12

# don't place it into $GOPATH/bin because Drone mounts $GOPATH as volume
COPY golangci-lint /usr/bin/
CMD ["golangci-lint"]
