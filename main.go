package main

import (
	"log"
	"flag"
	"net/http"

	"github.com/vharitonsky/iniflags"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/bwmarrin/snowflake"
)

const VERSION = "0.1.0/v6"

var (
	flgListen = flag.String("listen", "0.0.0.0:8081", "Listen address for API server")
	flgRedis = flag.String("redis", "127.0.0.1:16380", "Redis/ardb/EarlDB address")
	flgAllowReg = flag.Bool("allowRegistration", false, "Allow registration of accounts on this server")
	flgGatewayUrl = flag.String("apiGatewayUrl", "", "Specify round-robin URL for Gateway v6")

	flgNode = flag.Int64("node", 1, "Node ID")

	flgEnableEchoBanner = flag.Bool("debugHideEchoBanner", true, "Hide banner for echo/v4 framework")
)

var flake *snowflake.Node

var stopChan = make(chan error)

func main() {
	iniflags.Parse()
	log.Printf("info: jangle-server/%s loading...", VERSION)

	flake, _ = snowflake.NewNode(*flgNode)

	e := echo.New()
	e.HideBanner = *flgEnableEchoBanner

	e.Use(middleware.Recover())

	e.GET("/version.txt", func (c echo.Context) error { return c.String(http.StatusOK, VERSION); })

	InitGateway(e)

	e.Logger.Fatal(e.Start(*flgListen))
	log.Println("info:", "shutting down...")
}


