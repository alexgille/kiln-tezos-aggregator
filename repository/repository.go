package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	cnxPool *pgxpool.Pool
}

type Delegation struct {
	BlockTimestamp time.Time
	OperationID    int64
	Amount         int64
	Level          int32
	Sender         string
	BlockHash      string
}

func NewPostgresRepository(ctx context.Context, cnxString string) (PostgresRepository, error) {
	cnxPool, err := pgxpool.New(ctx, cnxString)
	if err != nil {
		return PostgresRepository{}, err
	}
	if err := cnxPool.Ping(ctx); err != nil {
		return PostgresRepository{}, err
	}
	return PostgresRepository{
		cnxPool: cnxPool,
	}, nil
}

// AddNewDelegations inserts delegations and skips duplicates.
func (p PostgresRepository) AddNewDelegations(ctx context.Context, dlgs []Delegation) error {
	const query = `
		INSERT INTO delegation (block_timestamp, operation_id, amount, level, sender, block_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`
	tx, err := p.cnxPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range dlgs {
		_, err := tx.Exec(ctx, query,
			dlgs[i].BlockTimestamp, dlgs[i].OperationID, dlgs[i].Amount, dlgs[i].Level, dlgs[i].Sender, dlgs[i].BlockHash,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// GetDelegations get all delegations sorted by block timestamp most recent first.
func (p PostgresRepository) GetDelegations(ctx context.Context) ([]Delegation, error) {
	const query = `
		SELECT block_timestamp, operation_id, amount, level, sender, block_hash
		FROM delegation
		ORDER BY block_timestamp DESC
	`

	rows, err := p.cnxPool.Query(ctx, query)
	if err != nil {
		return []Delegation{}, err
	}

	ret := make([]Delegation, 0)
	for rows.Next() {
		var dlg Delegation
		if err := rows.Scan(
			&dlg.BlockTimestamp, &dlg.OperationID, &dlg.Amount, &dlg.Level, &dlg.Sender, &dlg.BlockHash,
		); err != nil {
			return []Delegation{}, err
		}
		ret = append(ret, dlg)
	}

	return ret, nil
}

// GetDelegationsOfYear get all delegations of a given year sorted by block timestamp most recent first.
func (p PostgresRepository) GetDelegationsOfYear(ctx context.Context, year int) ([]Delegation, error) {
	const query = `
		SELECT block_timestamp, operation_id, amount, level, sender, block_hash
		FROM delegation
		WHERE EXTRACT(YEAR FROM block_timestamp) = $1
		ORDER BY block_timestamp DESC
	`

	rows, err := p.cnxPool.Query(ctx, query, year)
	if err != nil {
		return []Delegation{}, err
	}

	ret := make([]Delegation, 0)
	for rows.Next() {
		var dlg Delegation
		if err := rows.Scan(
			&dlg.BlockTimestamp, &dlg.OperationID, &dlg.Amount, &dlg.Level, &dlg.Sender, &dlg.BlockHash,
		); err != nil {
			return []Delegation{}, err
		}
		ret = append(ret, dlg)
	}

	return ret, nil
}

// GetLatestBlockTimestamp gets the most recent delegation's block timestamp.
func (p PostgresRepository) GetLatestBlockTimestamp(ctx context.Context) (time.Time, error) {
	const query = "SELECT block_timestamp FROM delegation ORDER BY block_timestamp DESC LIMIT 1"
	var ts time.Time
	if err := p.cnxPool.QueryRow(ctx, query).Scan(&ts); err != nil {
		return time.Time{}, err
	}
	return ts, nil
}
