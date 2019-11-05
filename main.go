package main

import (
	"log"
	"flag"
	"net/http"

	"github.com/vharitonsky/iniflags"
	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const VERSION = "0.1.0/v6"

var (
	flgListen = flag.String("listen", "0.0.0.0:8081", "Listen address for API server")
	flgRedis = flag.String("redis", "127.0.0.1:16380", "Redis/ardb/EarlDB address")
	flgAllowReg = flag.Bool("allowRegistration", false, "Allow registration of accounts on this server")
	flgGatewayUrl = flag.String("apiGatewayUrl", "", "Specify round-robin URL for Gateway v6")

	flgNode = flag.Int64("node", 1, "Node ID")
)

var flake *snowflake.Node

var stopChan = make(chan error)

func main() {
	iniflags.Parse()
	log.Printf("info: jangle-server/%s loading...", VERSION)

	flake, _ = snowflake.NewNode(*flgNode)

	r := router.New()

	e.GET("/version.txt", func (c echo.Context) error { return c.String(http.StatusOK, VERSION); })

	InitGateway(e)

	log.Fatal(fasthttp.ListenAndServe(*flgListen, r))
	log.Println("info:", "shutting down...")
}


