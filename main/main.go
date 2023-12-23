package main

import (
	_ "github.com/davidkhala/fabric-server-go/docs"
	"github.com/davidkhala/fabric-server-go/gateway"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/restful"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
)

// @title github.com/davidkhala/fabric-server-go
// @version v0.0.0
// @contact.email david-khala@hotmail.com
func main() {

	app := restful.App(true)
	app.StaticFile("/favicon.ico", "./favicon.ico")
	app.GET("/", restful.Ping)

	app.POST("/fabric/ping", gateway.PingFabric)
	app.POST("/fabric/create-proposal", gateway.CreateProposal)
	app.POST("/fabric/transact/process-proposal", gateway.ProcessProposal)
	app.POST("/fabric/transact/commit", gateway.Commit)
	app.POST("/ecosystem/create-token", CreateToken)
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // refers to /swagger/*any

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "8080"
	}
	goutils.PanicError(app.Run(":" + port))
}
