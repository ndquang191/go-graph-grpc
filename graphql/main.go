package main

import (
	"log"
	"net/http"

	// "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL" json:"account_url"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL" json:"catalog_url"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL" json:"order_url"`
}

func main() {
	cfg := AppConfig{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	s, err := NewGraphQLServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/graphql", handler.GraphQL(s.ToExecutableSchema()))
	http.Handle("/graphql", playground.Handler("quang", "/graphql"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
