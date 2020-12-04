# RPC
[![Test](https://github.com/rpccloud/rpc/workflows/Test/badge.svg)](https://github.com/rpccloud/rpc/actions?query=workflow%3ATest)
[![Lint](https://github.com/rpccloud/rpc/workflows/Lint/badge.svg)](https://github.com/rpccloud/rpc/actions?query=workflow%3ALint)


### 性能测试
```go
import _ "net/http/pprof"


go func() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
install graphviz
$ sudo brew install graphviz

look at a 30-second CPU profile:
$ curl http://localhost:6060/debug/pprof/profile?seconds=10 -o cpu.prof
$ go tool pprof -web cpu.prof
```


### golint
```bash
$ go get -u golang.org/x/lint/golint
$ golint ./...
```

### go fmt
```bash
$ go fmt ./...
```

### calculate lines of code
```bash
$ git ls-files | xargs wc -l
```

### 参考
https://github.com/tslearn/rpc-cluster-go/tree/984c4a17ffd777b268a13c0506aef83d9ba6b15d

### Problems
##### How to avoid session flood
