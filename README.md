*watcher* is file modified time monitoring tools.
```
go get -u github.com/secondarykey/watcher

$GOPATH/bin/watcher "go version"

If the specified path is updated, and then run the specified command .

Argument
  -target "search path:default workdir"
  -ignore "ignore path(; split)"
  -version "watcher version"
