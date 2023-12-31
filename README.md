# fabric-server-go
A restful server support offline signing

## Swagger Support
To generate swagger docs
1. get binary `swag` by `go install github.com/swaggo/swag/cmd/swag@latest`
2. `swag init -g main/main.go`
3. access swagger: go to `/swagger/index.html`

## Release
### Docker
```
docker pull ghcr.io/davidkhala/fabric-server-go:latest
```



## Dependencies
- Key required modules
  - `github.com/davidkhala/fabric-common/golang`: Wrapper or alternative of fabric-sdk-go. 
  - `github.com/gin-gonic/gin`: The using golang restful API framework
  - `github.com/hyperledger-twgc/tape`: A traffic generator of Fabric for benchmark test. Here, we appreciate its simplicity design and reuse some slim structure  
  - `github.com/hyperledger/fabric`: fabric itself. Used by importing its package `/protoutil`
  - `github.com/davidkhala/goutils`: generic golang utils. Used for grpc, http and other syntax-reform cases
  - `github.com/swaggo/gin-swagger`: The swagger docs generator for gin framework.
