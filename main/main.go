package main

//go:generate swag init -g main/main.go -o ../docs -d ../,../vendor/github.com/davidkhala/goutils/restful
import (
	_ "github.com/davidkhala/fabric-server-go/docs"
	"github.com/davidkhala/fabric-server-go/gateway"
	"github.com/davidkhala/goutils/restful"
)

// @title github.com/davidkhala/fabric-server-go
// @version v0.0.0
// @contact.email david-khala@hotmail.com
func main() {

	app, run := restful.SampleApp(8080)

	app.POST("/fabric/ping", gateway.PingFabric)
	app.POST("/fabric/create-proposal", gateway.CreateProposal)
	app.POST("/fabric/transact/process-proposal", gateway.ProcessProposal)
	app.POST("/fabric/transact/commit", gateway.Commit)
	app.POST("/ecosystem/create-token", CreateToken)

	run()
}
