package main

import (
	"log"
	"net/http"

	"multicarrier-email-api/internal/app"
	"multicarrier-email-api/internal/config"
)

const configFilePath = "config/app.yaml"

func newAppServer() *http.Server {
	cfg, err := config.NewFromYaml(configFilePath)
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
