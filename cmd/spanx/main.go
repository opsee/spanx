package main

import (
	"os"

	log "github.com/opsee/logrus"
	"github.com/opsee/spanx/service"
	"github.com/opsee/spanx/store"
)

func main() {
	var (
		postgresConn = mustEnvString("POSTGRES_CONN")
		listenAddr   = mustEnvString("SPANX_ADDRESS")
		cert         = mustEnvString("SPANX_CERT")
		certkey      = mustEnvString("SPANX_CERT_KEY")
	)

	db, err := store.NewPostgres(postgresConn)
	if err != nil {
		log.Fatalf("Error while initializing postgres: ", err)
	}

	log.Infof("service spanx starting grpc listener at %s", listenAddr)

	svc, err := service.New(db)
	if err != nil {
		log.Fatal("Error initializing service: ", err)
	}

	log.Fatal(svc.Start(listenAddr, cert, certkey))
}

func mustEnvString(envVar string) string {
	out := os.Getenv(envVar)
	if out == "" {
		log.Fatal(envVar, "must be set")
	}
	return out
}
