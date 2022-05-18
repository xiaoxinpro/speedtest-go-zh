package main

import (
	"flag"
	_ "time/tzdata"

	"github.com/xiaoxinpro/speedtest-go-zh/config"
	"github.com/xiaoxinpro/speedtest-go-zh/database"
	"github.com/xiaoxinpro/speedtest-go-zh/results"
	"github.com/xiaoxinpro/speedtest-go-zh/web"

	_ "github.com/breml/rootcerts"
	log "github.com/sirupsen/logrus"
)

var (
	optConfig = flag.String("c", "", "config file to be used, defaults to settings.toml in the same directory")
)

func main() {
	flag.Parse()
	conf := config.Load(*optConfig)
	web.SetServerLocation(&conf)
	results.Initialize(&conf)
	database.SetDBInfo(&conf)
	log.Fatal(web.ListenAndServe(&conf))
}
