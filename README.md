# RPC
[![Test](https://github.com/rpccloud/rpc/workflows/Test/badge.svg)](https://github.com/rpccloud/rpc/actions?query=workflow%3ATest)

### 性能测试
```bash
$ go get -u github.com/google/pprof
$ pprof -web cpu.prof
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
