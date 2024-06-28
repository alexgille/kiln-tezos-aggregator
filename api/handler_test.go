package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"kiln-tezos-delegation/api"
	"kiln-tezos-delegation/repository"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type controllerMock struct {
	GetDelegationHandlerRet   []repository.Delegation
	GetDelegationHandlerErr   error
	GetDelegationHandlerIn    int
	GetDelegationHandlerCount int
}

func (m *controllerMock) GetDelegations(_ context.Context, year int) ([]repository.Delegation, error) {
	m.GetDelegationHandlerIn = year
	return m.GetDelegationHandlerRet, m.GetDelegationHandlerErr
}

func TestGetDelegationHandler(t *testing.T) {
	t.Run("successfully return data", func(t *testing.T) {
		mock := controllerMock{
			GetDelegationHandlerRet: []repository.Delegation{
				{
					BlockTimestamp: time.Date(2024, 06, 26, 10, 02, 33, 0, time.UTC),
					Sender:         "addr1",
					Level:          142,
					Amount:         242,
				},
			},
		}
		req := httptest.NewRequest("GET", "/xtz/delegations", http.NoBody)
		resp := httptest.NewRecorder()

		hdl := api.GetDelegationHandler(&mock)
		hdl.ServeHTTP(resp, req)

		pld := make(map[string][]map[string]string)
		err := json.NewDecoder(resp.Body).Decode(&pld)
		require.NoError(t, err)

		assert.Equal(t, resp.Code, http.StatusOK)
		assert.Len(t, pld["data"], 1)
		assert.Equal(t, "2024-06-26T10:02:33Z", pld["data"][0]["timestamp"])
		assert.Equal(t, "addr1", pld["data"][0]["delegator"])
		assert.Equal(t, "142", pld["data"][0]["level"])
		assert.Equal(t, "242", pld["data"][0]["amount"])
		assert.Equal(t, api.YearNotSpecified, mock.GetDelegationHandlerIn)
	})

	t.Run("handles year query parameter", func(t *testing.T) {
		mock := controllerMock{
			GetDelegationHandlerRet: []repository.Delegation{
				{
					BlockTimestamp: time.Date(2024, 06, 26, 10, 02, 33, 0, time.UTC),
					Sender:         "addr1",
					Level:          142,
					Amount:         242,
				},
			},
		}
		req := httptest.NewRequest("GET", "/xtz/delegations?year=2024", http.NoBody)
		resp := httptest.NewRecorder()

		hdl := api.GetDelegationHandler(&mock)
		hdl.ServeHTTP(resp, req)

		pld := make(map[string][]map[string]string)
		err := json.NewDecoder(resp.Body).Decode(&pld)
		require.NoError(t, err)

		assert.Equal(t, resp.Code, http.StatusOK)
		assert.Len(t, pld["data"], 1)
		assert.Equal(t, 2024, mock.GetDelegationHandlerIn)
	})

	t.Run("keeps delegations ordered", func(t *testing.T) {
		mock := controllerMock{
			GetDelegationHandlerRet: []repository.Delegation{
				{Sender: "addr1"}, {Sender: "addr2"},
			},
		}
		req := httptest.NewRequest("GET", "/xtz/delegations", http.NoBody)
		resp := httptest.NewRecorder()

		hdl := api.GetDelegationHandler(&mock)
		hdl.ServeHTTP(resp, req)

		pld := make(map[string][]map[string]string)
		err := json.NewDecoder(resp.Body).Decode(&pld)
		require.NoError(t, err)

		assert.Equal(t, resp.Code, http.StatusOK)
		assert.Len(t, pld["data"], 2)
		assert.Equal(t, "addr1", pld["data"][0]["delegator"])
		assert.Equal(t, "addr2", pld["data"][1]["delegator"])
	})

	t.Run("status code on bad query parameter", func(t *testing.T) {
		testCases := []string{"1", "abc", "12345"}
		for _, val := range testCases {
			mock := controllerMock{ /* unused */ }
			req := httptest.NewRequest("GET", "/xtz/delegations?year="+val, http.NoBody)
			resp := httptest.NewRecorder()

			hdl := api.GetDelegationHandler(&mock)
			hdl.ServeHTTP(resp, req)

			assert.Equal(t, resp.Code, http.StatusBadRequest)
			assert.Equal(t, 0, resp.Body.Len())
			assert.Equal(t, 0, mock.GetDelegationHandlerCount)
		}
	})

	t.Run("status code on controller error", func(t *testing.T) {
		mock := controllerMock{
			GetDelegationHandlerErr: errors.New("fake controller error"),
		}
		req := httptest.NewRequest("GET", "/xtz/delegations", http.NoBody)
		resp := httptest.NewRecorder()

		hdl := api.GetDelegationHandler(&mock)
		hdl.ServeHTTP(resp, req)

		assert.Equal(t, resp.Code, http.StatusInternalServerError)
		assert.Equal(t, 0, resp.Body.Len())
		assert.Equal(t, 0, mock.GetDelegationHandlerCount)
	})

	t.Run("status code on bad method", func(t *testing.T) {
		mock := controllerMock{}
		req := httptest.NewRequest("POST", "/xtz/delegations", http.NoBody)
		resp := httptest.NewRecorder()

		hdl := api.GetDelegationHandler(&mock)
		hdl.ServeHTTP(resp, req)

		assert.Equal(t, resp.Code, http.StatusMethodNotAllowed)
		assert.Equal(t, 0, resp.Body.Len())
		assert.Equal(t, 0, mock.GetDelegationHandlerCount)
	})
}
