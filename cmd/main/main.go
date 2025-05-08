package main

import (
	_ "embed"
	"log"
	"net/http"

	"multicarrier-email-api/internal/app"
	"multicarrier-email-api/internal/config"
)

//go:embed config/config.yaml
var configYamlContent []byte

func newAppServer() *http.Server {
	cfg, err := config.NewFromYamlContent(configYamlContent)
	if err != nil {
		log.Panicf("error loading config: %v", err)
		return nil
	}

	appInstance := app.NewApp(cfg)
	return appInstance.NewServer(cfg.Server.Port)
}

var newAppServerFn = newAppServer

func main() {
	server := newAppServerFn()
	log.Print(server.ListenAndServe())
}
