*watcher* is file modified time monitoring tools.
```
go get -u github.com/secondarykey/watcher

$GOPATH/bin/watcher "go version"

If the specified path is updated, and then run the specified command .

Argument
  -target "search path:default 'os.Getwd()'"
  -ignore "ignore path(; split):default 'tmp;cache;.swp'"
  -version "watcher version"

Warning:
  Source in the [Qiita](http://qiita.com/secondarykey/items/6fa481cbdee24632e80e) can be found [here](https://github.com/secondarykey/watcher/releases/tag/original)
