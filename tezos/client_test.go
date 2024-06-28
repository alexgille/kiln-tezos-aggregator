package tezos_test

import (
	"context"
	"kiln-tezos-delegation/tezos"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Run("calls current protocol endpoint correctly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/protocols/current", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"constants":{"timeBetweenBlocks": 10}}`))
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		var gotDur time.Duration
		gotDur, err = cli.GetCurrentProtocolTimeBetweenBlocks(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, gotDur)
	})

	t.Run("returns error on current protocol fetch bad status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = cli.GetCurrentProtocolTimeBetweenBlocks(context.Background())

		assert.Error(t, err)
	})

	t.Run("returns error on current protocol response not json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/protocols/current", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NOT JSON"))
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = cli.GetCurrentProtocolTimeBetweenBlocks(context.Background())

		assert.Error(t, err)
	})

	t.Run("calls delegation operations endpoint correctly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/operations/delegations", r.URL.Path)
			assert.Equal(t, 3, len(r.URL.Query())) // only 3 query parameters
			assert.Equal(t, "id,sender,amount,level,timestamp,block", r.URL.Query().Get("select"))
			assert.Equal(t, "id", r.URL.Query().Get("sort.asc"))
			assert.Equal(t, "1991-03-01T10:25:07Z", r.URL.Query().Get("timestamp.ge"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"id":42,"timestamp":"2024-06-25T10:02:33Z","block":"hash1","sender":{"address":"addr1"},"level":242,"amount":342},
				{"id":43,"timestamp":"2024-06-25T14:02:33Z","block":"hash2","sender":{"address":"addr2"},"level":243,"amount":343}
			]`))
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		var gotDlgs []tezos.Delegation
		gotDlgs, err = cli.GetDelegationsSince(context.Background(), time.Date(1991, 03, 01, 10, 25, 07, 0, time.UTC))

		assert.NoError(t, err)
		assert.Len(t, gotDlgs, 2)
		assert.Equal(t, int64(42), gotDlgs[0].ID)
		assert.Equal(t, time.Date(2024, 06, 25, 10, 02, 33, 0, time.UTC), gotDlgs[0].Timestamp)
		assert.Equal(t, "hash1", gotDlgs[0].Block)
		assert.Equal(t, "addr1", gotDlgs[0].Sender.Address)
		assert.Equal(t, int32(242), gotDlgs[0].Level)
		assert.Equal(t, int64(342), gotDlgs[0].Amount)
	})

	t.Run("returns error delegation operations fetch bad status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = cli.GetDelegationsSince(context.Background(), time.Time{})

		assert.Error(t, err)
	})

	t.Run("returns error delegation operations response not json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NOT JSON"))
		}))
		defer server.Close()

		cli, err := tezos.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = cli.GetDelegationsSince(context.Background(), time.Time{})

		assert.Error(t, err)
	})
}
