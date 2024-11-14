package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/ndquang191/go-graph-grpc/account"
	"github.com/tinrab/retry"
	"log"
	"time"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		r, err = account.NewPostgresRepository(config.DatabaseURL)
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	})

	defer r.Close()

	log.Println('s', "Starting server")
	s := account.NewService(r)

	log.Fatal(account.ListenGRPC(s, 8080))
}
