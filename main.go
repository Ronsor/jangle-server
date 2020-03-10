package main

import (
	"flag"
	"log"
	"syscall"

	"jangled/util"

	"github.com/bwmarrin/snowflake"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/vharitonsky/iniflags"
)

const VERSION = "version=0.1.2 GW-level=v6 API-level=v6,v7(?) \"I can't believe it's not Discord!\" Edition [prerelease]"

var (
	flgListen     = flag.String("listen", "0.0.0.0:8081", "Listen address for API server")
	flgMongoDB    = flag.String("mongo", "mongodb://127.0.0.1:3600/?maxIdleTimeMS=0", "MongoDB URI")
	flgMsgMongoDB = flag.String("mongoMessages", "primary:", "MongoDB URI for messages database")
	flgSmtpServer = flag.String("smtp", "127.0.0.1:25", "SMTP server for sending emails")

	flgAllowReg   = flag.Bool("allowRegistration", true, "Allow registration of accounts on this server")
	flgGatewayUrl = flag.String("apiGatewayUrl", "", "Specify URL for Gateway v6")
	flgStaging    = flag.Bool("staging", false, "Add dummy data for testing")
	flgNoPanic    = flag.Bool("nopanic", true, "Catch all panics in API handlers")

	flgEnableFileServer = flag.Bool("enableFileServer", true, "Enable file server (BogusCDN)")
	flgFileServerPath   = flag.String("fileServerPath", "/tmp/janglefileserver", "File server path")

	flgCDNServeBase = flag.String("cdnServeBase", "https://cdn.jangleapp.com", "CDN base URL")
	flgCDNUploadBase = flag.String("cdnUploadBase", "", "CDN upload base URL")

	flgObjCacheLimit = flag.Int("cachelimit", 4096, "Object cache limit")

	flgNode = flag.Int64("node", 1, "Node ID")
)

var flake *snowflake.Node
var stopChan = make(chan error)

func main() {
	iniflags.Parse()
	if syscall.Getuid() == 1 {
		syscall.Chdir(".")
		syscall.Chroot(".")
		syscall.Setgid(65534)
		syscall.Setuid(65534)
	}

	util.NoPanic = *flgNoPanic
	gCache.Limit(*flgObjCacheLimit)
	if *flgMsgMongoDB == "primary:" {
		flgMsgMongoDB = flgMongoDB
	}

	log.Printf("info: jangle-jangled/%s loading...", VERSION)

	flake, _ = snowflake.NewNode(*flgNode)

	log.Printf("info: initialized snowflake engine (node=%d)", *flgNode)

	InitDB()
	log.Printf("info: initialized mongodb")

	r := router.New()
	if *flgNoPanic {
		r.PanicHandler = func(c *fasthttp.RequestCtx, e interface{}) {
			if err, ok := e.(*APIResponseError); ok {
				util.WriteJSONStatus(c, 500, err)
				return
			}
			log.Printf("Internal error: %v", e)
			util.WriteJSONStatus(c, 500, &APIResponseError{0, "Unknown error"})
		}
	}

	r.GET("/version.txt", func(c *fasthttp.RequestCtx) { c.WriteString(VERSION) })
	InitRestUser(r)
	InitRestChannel(r)
	InitRestGuild(r)
	InitRestAuth(r)
	log.Printf("info: initialized rest api routes")

	InitSessionManager()
	InitGateway(r)
	log.Printf("info: initialized gateway routes")

	if *flgEnableFileServer {
		gFileStore.Init(r)
		log.Printf("info: initialized boguscdn")
	}

	log.Printf("info: starting http server (addr=%s)", *flgListen)

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		r.Handle(method, "/api/v7/*path", func (c *fasthttp.RequestCtx) {
			c.Request.URI().SetPath(c.UserValue("path").(string))
			r.Handler(c)
		})
	}

	log.Fatal(fasthttp.ListenAndServe(*flgListen, MwAccCtl(r.Handler, "*")))
	log.Println("info:", "shutting down...")
}
