package main

import (
	"github.com/opsee/spanx/service"
	"github.com/opsee/spanx/store"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	postgresConn := mustEnvString("POSTGRES_CONN")
	db, err := store.NewPostgres(postgresConn)
	if err != nil {
		log.Fatalf("Error while initializing postgres: ", err)
	}

	listenAddr := mustEnvString("SPANX_ADDRESS")
	log.Infof("service spanx starting http listener at %s", listenAddr)

	svc := service.New(db)
	svc.StartHTTP(listenAddr)
}

func mustEnvString(envVar string) string {
	out := os.Getenv(envVar)
	if out == "" {
		log.Fatal(envVar, "must be set")
	}
	return out
}
