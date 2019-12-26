package main

import (
	"log"
	"flag"

	"jangled/util"

	"github.com/vharitonsky/iniflags"
	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const VERSION = "0.1.0/v6"

var (
	flgListen = flag.String("listen", "0.0.0.0:8081", "Listen address for API server")
	flgMongoDB = flag.String("mongo", "mongodb://127.0.0.1:3600/?maxIdleTimeMS=0", "MongoDB URI")

	flgAllowReg = flag.Bool("allowRegistration", false, "Allow registration of accounts on this server")
	flgGatewayUrl = flag.String("apiGatewayUrl", "", "Specify round-robin URL for Gateway v6")
	flgStaging = flag.Bool("staging", false, "Add dummy data for testing")
	flgNoPanic = flag.Bool("nopanic", true, "Catch all panics in API handlers")

	flgNode = flag.Int64("node", 1, "Node ID")
)

var flake *snowflake.Node
var stopChan = make(chan error)

func main() {
	iniflags.Parse()
	util.NoPanic = *flgNoPanic
	log.Printf("info: jangle-jangled/%s loading...", VERSION)

	flake, _ = snowflake.NewNode(*flgNode)

	log.Printf("info: initialized snowflake engine (node=%d)", *flgNode)

	InitDB()
	log.Printf("info: initialized mongodb")

	r := router.New()

	r.GET("/version.txt", func (c *fasthttp.RequestCtx) { c.WriteString(VERSION) })
	InitRestUser(r)
	InitRestChannel(r)
	log.Printf("info: initialized rest api routes")

	InitSessionManager()
	InitGateway(r)
	log.Printf("info: initialized gateway routes")

	log.Printf("info: starting http server (addr=%s)", *flgListen)
	log.Fatal(fasthttp.ListenAndServe(*flgListen, r.Handler))
	log.Println("info:", "shutting down...")
}


