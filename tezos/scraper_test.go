package tezos_test

import (
	"context"
	"errors"
	"kiln-tezos-delegation/repository"
	"kiln-tezos-delegation/tezos"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

type clientMock struct {
	GetCurrentProtocolTimeBetweenBlocksRet   time.Duration
	GetCurrentProtocolTimeBetweenBlocksErr   error
	GetCurrentProtocolTimeBetweenBlocksCount int
	GetDelegationsSinceRet                   []tezos.Delegation
	GetDelegationsSinceErr                   error
	GetDelegationsSinceCount                 int
	GetDelegationsSinceIn                    time.Time
}

func (m *clientMock) GetCurrentProtocolTimeBetweenBlocks(context.Context) (time.Duration, error) {
	m.GetCurrentProtocolTimeBetweenBlocksCount++
	return m.GetCurrentProtocolTimeBetweenBlocksRet, m.GetCurrentProtocolTimeBetweenBlocksErr
}

func (m *clientMock) GetDelegationsSince(_ context.Context, since time.Time) ([]tezos.Delegation, error) {
	m.GetDelegationsSinceIn = since
	m.GetDelegationsSinceCount++
	return m.GetDelegationsSinceRet, m.GetDelegationsSinceErr
}

type repoMock struct {
	AddNewDelegationsErr         error
	AddNewDelegationsCount       int
	AddNewDelegationsIn          []repository.Delegation
	GetLatestBlockTimestampRet   time.Time
	GetLatestBlockTimestampErr   error
	GetLatestBlockTimestampCount int
}

func (m *repoMock) AddNewDelegations(_ context.Context, tezosDlgs []repository.Delegation) error {
	m.AddNewDelegationsIn = tezosDlgs
	m.AddNewDelegationsCount++
	return m.AddNewDelegationsErr
}

func (m *repoMock) GetLatestBlockTimestamp(context.Context) (time.Time, error) {
	m.GetLatestBlockTimestampCount++
	return m.GetLatestBlockTimestampRet, m.GetLatestBlockTimestampErr
}

func TestScraper(t *testing.T) {
	scrapInterval := 200 * time.Millisecond
	waitTime := 300 * time.Millisecond

	tezosDlgs := []tezos.Delegation{
		{
			ID:        42,
			Timestamp: time.Date(2024, 06, 25, 10, 02, 33, 0, time.UTC),
			Block:     "hash1",
			Sender: struct {
				Address string `json:"address"`
			}{Address: "addr1"},
			Level:  242,
			Amount: 342,
		},
		{
			ID:        43,
			Timestamp: time.Date(2024, 06, 25, 14, 02, 33, 0, time.UTC),
			Block:     "hash2",
			Sender: struct {
				Address string `json:"address"`
			}{Address: "addr2"},
			Level:  243,
			Amount: 343,
		},
	}

	t.Run("happy case with initial data", func(t *testing.T) {
		lastBlockTs := time.Date(2024, 06, 26, 10, 02, 33, 0, time.UTC)
		begin := time.Time{}
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: scrapInterval,
			GetDelegationsSinceRet:                 tezosDlgs,
		}
		repoMock := repoMock{
			GetLatestBlockTimestampRet: lastBlockTs,
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go scraper.Run(ctx, begin)

		time.Sleep(waitTime)
		cancel()

		// tezos client & repo have been called
		assert.Equal(t, 1, cliMock.GetDelegationsSinceCount)
		assert.Equal(t, 1, repoMock.GetLatestBlockTimestampCount)
		assert.Equal(t, 1, repoMock.AddNewDelegationsCount)
		// all data has been passed around to repository layer
		assert.Len(t, repoMock.AddNewDelegationsIn, 2)
		assert.Equal(t, cliMock.GetDelegationsSinceRet[0].ID, repoMock.AddNewDelegationsIn[0].OperationID)
		assert.Equal(t, cliMock.GetDelegationsSinceRet[1].ID, repoMock.AddNewDelegationsIn[1].OperationID)
		// tezos & repository BOMs are equivalent in any aspect
		expBOM := repository.Delegation{
			OperationID:    tezosDlgs[0].ID,
			BlockTimestamp: tezosDlgs[0].Timestamp,
			BlockHash:      tezosDlgs[0].Block,
			Sender:         tezosDlgs[0].Sender.Address,
			Level:          tezosDlgs[0].Level,
			Amount:         tezosDlgs[0].Amount,
		}
		assert.Equal(t, repoMock.AddNewDelegationsIn[0], expBOM)
		// scraper started from (latest block timestamp + 1 second)(since TzKT precision is one second)
		assert.Equal(t, lastBlockTs, cliMock.GetDelegationsSinceIn.Add(-time.Second))
	})

	t.Run("happy case without initial data", func(t *testing.T) {
		begin := time.Time{}
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: scrapInterval,
			GetDelegationsSinceRet:                 tezosDlgs,
		}
		repoMock := repoMock{
			GetLatestBlockTimestampErr: pgx.ErrNoRows, // no initial data in db
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go scraper.Run(ctx, begin)

		time.Sleep(waitTime)
		cancel()

		// tezos client & repo have been called
		assert.Equal(t, 1, cliMock.GetDelegationsSinceCount)
		assert.Equal(t, 1, repoMock.GetLatestBlockTimestampCount)
		assert.Equal(t, 1, repoMock.AddNewDelegationsCount)
		// data has been passed to repository layer as expected
		assert.Len(t, repoMock.AddNewDelegationsIn, 2)
		assert.Equal(t, cliMock.GetDelegationsSinceRet[0].ID, repoMock.AddNewDelegationsIn[0].OperationID)
		// scraper started from since less than one second from now
		assert.WithinDuration(t, time.Now(), cliMock.GetDelegationsSinceIn, time.Second)
	})

	t.Run("starts from expected forced time", func(t *testing.T) {
		begin := time.Date(1991, 01, 03, 10, 02, 33, 0, time.UTC) // desired date-time (force)
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: scrapInterval,
			GetDelegationsSinceRet:                 tezosDlgs,
		}
		repoMock := repoMock{
			GetLatestBlockTimestampErr: pgx.ErrNoRows, // no initial data in db
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go scraper.Run(ctx, begin)

		time.Sleep(waitTime)
		cancel()

		// latest block timestamp has not been fetched from database
		assert.Equal(t, 0, repoMock.GetLatestBlockTimestampCount)
		// scraper started from the expected (forced) time
		assert.Equal(t, begin, cliMock.GetDelegationsSinceIn)
	})

	t.Run("return error on protocol request failure", func(t *testing.T) {
		begin := time.Date(1991, 01, 03, 10, 02, 33, 0, time.UTC) // desired date-time (force)
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksErr: errors.New("fake http error"),
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock{})

		gotErr := scraper.Run(context.Background(), begin)

		// no ticker created in this case; no need to wait

		assert.Error(t, gotErr)
	})

	t.Run("return error on timestamp fetch failure", func(t *testing.T) {
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: time.Millisecond,
		}
		repoMock := repoMock{
			GetLatestBlockTimestampErr: errors.New("fake database error"),
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock)

		err := scraper.Run(context.Background(), time.Time{})

		// ticker not blocking; no need to wait

		assert.Error(t, err)
	})

	t.Run("starts from same timestamp on client error", func(t *testing.T) {
		begin := time.Date(2024, 06, 26, 10, 02, 33, 0, time.UTC)
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: scrapInterval,
			GetDelegationsSinceErr:                 errors.New("fake TzKT error"),
		}
		scraper := tezos.NewScraper(&cliMock, &repoMock{})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go scraper.Run(ctx, begin)

		// wait a bit more than two cycles
		time.Sleep(2*scrapInterval + scrapInterval/4)
		cancel()

		// delegations were fetched twice ...
		assert.Equal(t, 2, cliMock.GetDelegationsSinceCount)
		// ... and at the second call the same begin time was used
		assert.Equal(t, begin, cliMock.GetDelegationsSinceIn)
	})

	t.Run("starts from now when no new delegations", func(t *testing.T) {
		begin := time.Date(2020, 06, 26, 10, 02, 33, 0, time.UTC)
		cliMock := clientMock{
			GetCurrentProtocolTimeBetweenBlocksRet: scrapInterval,
			GetDelegationsSinceRet:                 []tezos.Delegation{}, // nothing new
		}
		repoMock := repoMock{}
		scraper := tezos.NewScraper(&cliMock, &repoMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go scraper.Run(ctx, begin)

		// wait a bit more than two cycles
		time.Sleep(2*scrapInterval + scrapInterval/4)
		cancel()

		// delegations were fetched twice ...
		assert.Equal(t, 2, cliMock.GetDelegationsSinceCount)
		// ... and at the second call the begin time was within a second to now
		assert.WithinDuration(t, time.Now(), cliMock.GetDelegationsSinceIn, time.Second)
	})
}
