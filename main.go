package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const version = "0.3"

var printVersion bool
var listen string

func init() {
	flag.BoolVar(&printVersion, "v", false, "Print version and exit")
	flag.StringVar(&listen, "bind", "0.0.0.0:3030", "Bind the service on this port")
	flag.Parse()
}

func buildMux() *gin.Engine {
	var authorized *gin.RouterGroup
	m := gin.Default()

	userAndPass := coalesceEnv("AUTH", "user:pass")
	if userAndPass != "" {
		up := strings.Split(userAndPass, ":")
		authorized = m.Group("/", gin.BasicAuth(gin.Accounts{
			up[0]: up[1],
		}))
	}

	// Service Instance
	authorized.POST("/resources", Add)
	authorized.GET("/resources/:name", GetServiceInstance)
	authorized.PUT("/resources/:name", UpdateServiceInstance)
	authorized.DELETE("/resources/:name", Remove)
	authorized.GET("/resources/:name/status", Status)

	// Binding
	authorized.POST("/resources/:name/bind", BindUnit)
	authorized.DELETE("/resources/:name/bind", UnbindUnit)
	authorized.POST("/resources/:name/bind-app", BindApp)
	authorized.DELETE("/resources/:name/bind-app", UnbindApp)

	//

	return m
}

func main() {
	if printVersion {
		fmt.Printf("eaas version %s", version)
		return
	}
	log.Fatal(http.ListenAndServe(listen, buildMux()))
}
