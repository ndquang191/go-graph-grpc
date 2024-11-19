package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/ndquang191/go-graph-grpc/catalog"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config

	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)

	}

	var r catalog.Repository

	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = catalog.NewElasticRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	defer r.Close()
	log.Println("Connected to Elastic")

	s := catalog.NewService(r)
	log.Fatal(catalog.ListenGRPC(s, 8080))
}
