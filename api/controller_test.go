package api_test

import (
	"context"
	"errors"
	"kiln-tezos-delegation/api"
	"kiln-tezos-delegation/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

type repoMock struct {
	GetDelegationsRet         []repository.Delegation
	GetDelegationsErr         error
	GetDelegationsCount       int
	GetDelegationsOfYearRet   []repository.Delegation
	GetDelegationsOfYearErr   error
	GetDelegationsOfYearCount int
	GetDelegationsOfYearIn    int
}

func (m *repoMock) GetDelegations(context.Context) ([]repository.Delegation, error) {
	m.GetDelegationsCount++
	return m.GetDelegationsRet, m.GetDelegationsErr
}

func (m *repoMock) GetDelegationsOfYear(_ context.Context, year int) ([]repository.Delegation, error) {
	m.GetDelegationsOfYearIn = year
	m.GetDelegationsOfYearCount++
	return m.GetDelegationsOfYearRet, m.GetDelegationsOfYearErr
}

func TestGetDelegations(t *testing.T) {
	t.Run("no year filter", func(t *testing.T) {
		mock := repoMock{
			GetDelegationsRet: []repository.Delegation{
				{OperationID: 42}, {OperationID: 43},
			},
		}
		ctl := api.NewController(&mock)
		dlgs, err := ctl.GetDelegations(context.Background(), api.YearNotSpecified)
		assert.NoError(t, err)
		assert.Len(t, dlgs, 2)
		assert.Equal(t, int64(42), dlgs[0].OperationID)
		assert.Equal(t, int64(43), dlgs[1].OperationID)
		assert.Equal(t, 1, mock.GetDelegationsCount)
		assert.Equal(t, 0, mock.GetDelegationsOfYearCount)
	})

	t.Run("with year filter", func(t *testing.T) {
		mock := repoMock{
			GetDelegationsOfYearRet: []repository.Delegation{
				{OperationID: 42}, {OperationID: 43},
			},
		}
		ctl := api.NewController(&mock)
		dlgs, err := ctl.GetDelegations(context.Background(), 2024)
		assert.NoError(t, err)
		assert.Len(t, dlgs, 2)
		assert.Equal(t, int64(42), dlgs[0].OperationID)
		assert.Equal(t, int64(43), dlgs[1].OperationID)
		assert.Equal(t, 0, mock.GetDelegationsCount)
		assert.Equal(t, 1, mock.GetDelegationsOfYearCount)
		assert.Equal(t, 2024, mock.GetDelegationsOfYearIn)
	})

	t.Run("handler error without year", func(t *testing.T) {
		mock := repoMock{
			GetDelegationsErr: errors.New("fake database error"),
		}
		ctl := api.NewController(&mock)
		_, err := ctl.GetDelegations(context.Background(), api.YearNotSpecified)
		assert.Error(t, err)
	})

	t.Run("handler error with year", func(t *testing.T) {
		mock := repoMock{
			GetDelegationsOfYearErr: errors.New("fake database error"),
		}
		ctl := api.NewController(&mock)
		_, err := ctl.GetDelegations(context.Background(), 2024)
		assert.Error(t, err)
	})
}
