package tezos

import (
	"context"
	"errors"
	"kiln-tezos-delegation/repository"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type TezosClient interface {
	GetCurrentProtocolTimeBetweenBlocks(context.Context) (time.Duration, error)
	GetDelegationsSince(context.Context, time.Time) ([]Delegation, error)
}

type TezosRepository interface {
	AddNewDelegations(context.Context, []repository.Delegation) error
	GetLatestBlockTimestamp(context.Context) (time.Time, error)
}

type Scraper struct {
	client TezosClient
	repo   TezosRepository
}

func NewScraper(client TezosClient, repo TezosRepository) *Scraper {
	return &Scraper{
		client: client,
		repo:   repo,
	}
}

// Run fetches the execution interval from the TzKT client then starts
// scraping periodically until context is cancelled, or any non-recoverable
// error occurs. An error is returned in that later case.
// If a time is passed as parameter, it will be used as an override of
// the starting time of scraping. It is mostly for testing purposes.
func (s *Scraper) Run(ctx context.Context, beginning time.Time) error {
	interval, err := s.client.GetCurrentProtocolTimeBetweenBlocks(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if beginning.IsZero() {
		beginning, err = s.getStartingTime(ctx)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Default().Println("scraping since", beginning)
			beginning, err = s.scrapDelegations(ctx, beginning)
			if err != nil {
				// do not return, instead log and try again
				log.Default().Println("error during scraping cycle:", err)
			}
		}
	}
}

// scrapDelegations gets the delegation operations from TzKT API starting from
// the time passed as parameter then stores them in storage.
// Returns the time suitable for the next cycle to start with, or an error.
// The returned time is guaranteed to be equal to the one passed as parameter
// whenever the fetched delegations could not be put in storage.
func (s *Scraper) scrapDelegations(ctx context.Context, beginning time.Time) (time.Time, error) {
	// get delegations since the desired beginning from TzKT API
	dlgs, err := s.client.GetDelegationsSince(ctx, beginning)
	if err != nil {
		return beginning, err
	}

	log.Default().Println("fetched", len(dlgs), "delegation(s) from TzKT API")

	if len(dlgs) == 0 {
		// no new delegation, lets start next cycle from now
		return time.Now().UTC(), nil
	}

	// convert BOMs and find oldest timestamp from new delegations
	oldest := beginning
	rdlgs := make([]repository.Delegation, len(dlgs))
	for i := range dlgs {
		rdlgs[i].Amount = dlgs[i].Amount
		rdlgs[i].BlockHash = dlgs[i].Block
		rdlgs[i].OperationID = dlgs[i].ID
		rdlgs[i].BlockTimestamp = dlgs[i].Timestamp
		rdlgs[i].Level = dlgs[i].Level
		rdlgs[i].Sender = dlgs[i].Sender.Address
		if rdlgs[i].BlockTimestamp.After(oldest) {
			oldest = rdlgs[i].BlockTimestamp
		}
	}

	// save in repository
	if err := s.repo.AddNewDelegations(ctx, rdlgs); err != nil {
		return beginning, err
	}

	return oldest.Add(time.Second), nil
}

// getStartingTime fetches the timestamp of the delegation operation most
// recent first, from storage, and returns it with one second added. This
// is to avoid scraping delegation operations having their timestamp equal
// to the most recent timestamp being in storage twice at each cycle.
// If storage contains no operation, the current time is used.
func (s *Scraper) getStartingTime(ctx context.Context) (time.Time, error) {
	latest, err := s.repo.GetLatestBlockTimestamp(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, err
	}
	if latest.IsZero() {
		return time.Now().UTC(), nil
	}
	return latest.Add(time.Second).UTC(), nil
}
