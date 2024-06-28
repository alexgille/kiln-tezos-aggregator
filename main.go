package main

import (
	"context"
	"fmt"
	"kiln-tezos-delegation/api"
	"kiln-tezos-delegation/repository"
	"kiln-tezos-delegation/tezos"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type config struct {
	apiAddr    string
	dbHost     string
	dbDatabase string
	dbUser     string
	dbPassword string
	tzktHost   string
	since      time.Time
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal("finished with error: ", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	conf := confFromEnv()
	repo := initRepository(ctx, conf)
	client := initTezosClient(conf)
	svr := initAPI(conf, repo)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer cancel()
		defer wg.Done()
		if err := svr.Start(cctx); err != nil {
			errChan <- fmt.Errorf("api server error: %w", err)
		}
	}()

	go func() {
		defer cancel()
		defer wg.Done()
		if err := tezos.NewScraper(client, repo).Run(cctx, conf.since); err != nil {
			errChan <- fmt.Errorf("scraper error: %w", err)
		}
	}()

	wg.Wait()

	return <-errChan
}

func initRepository(ctx context.Context, conf config) repository.PostgresRepository {
	url := "postgres://" + conf.dbUser + ":" + conf.dbPassword + "@" + conf.dbHost + "/" + conf.dbDatabase
	repo, err := repository.NewPostgresRepository(ctx, url)
	if err != nil {
		log.Fatal("repository: ", err)
	}
	return repo
}

func initTezosClient(conf config) tezos.TezosClient {
	client, err := tezos.NewClient(conf.tzktHost)
	if err != nil {
		log.Fatal("tezos client: ", err)
	}
	return client
}

func initAPI(conf config, repo repository.PostgresRepository) *api.Server {
	return api.NewServer(conf.apiAddr, "/xtz/delegations", api.GetDelegationHandler(api.NewController(repo)))
}

func confFromEnv() config {
	conf := config{
		apiAddr:    os.Getenv("API_ADDR"),
		dbHost:     os.Getenv("DB_HOST"),
		dbDatabase: os.Getenv("DB_DATABASE"),
		dbUser:     os.Getenv("DB_USER"),
		dbPassword: os.Getenv("DB_PASSWORD"),
		tzktHost:   os.Getenv("TZKT_BASE_URL"),
	}

	var err error
	if since := os.Getenv("SCRAP_SINCE"); since != "" {
		conf.since, err = time.Parse(time.RFC3339, since)
		if err != nil {
			log.Fatal("env: SCRAP_SINCE: ", err)
		}
	}

	return conf
}
