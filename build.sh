go build -ldflags "-s -w -X main.version=`git rev-parse --short HEAD`" --tags "libsqlite3 linux"
