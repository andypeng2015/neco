// This server can run on Google App Engine.
package main

import (
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/neco/gcp"
	"github.com/cybozu-go/neco/gcp/app"
	"google.golang.org/appengine"
	yaml "gopkg.in/yaml.v2"
)

const (
	cfgFile = ".necogcp.yml"
)

func main() {
	// seed math/random
	rand.Seed(time.Now().UnixNano())

	cfg, err := gcp.NewConfig()
	if err != nil {
		log.ErrorExit(err)
	}
	f, err := os.Open(cfgFile)
	if err != nil {
		log.ErrorExit(err)
	}
	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		log.ErrorExit(err)
	}
	f.Close()

	server, err := app.NewServer(cfg)
	if err != nil {
		log.ErrorExit(err)
	}
	http.HandleFunc("/shutdown", server.HandleShutdown)

	appengine.Main()
}
