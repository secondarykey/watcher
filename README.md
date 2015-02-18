*watcher* is file modified time monitoring tools.
```
go get github.com/secondarykey/watcher
cd $GOPATH/src/github.com/secondarykey/watcher/cmd/watcher
go install 

watcher "go version"

run command

Argument
  -target "search path"
  -patrol "write search date"
  -duration "search duration"
  -ignore "ignore path"
  -version "watcher version"
