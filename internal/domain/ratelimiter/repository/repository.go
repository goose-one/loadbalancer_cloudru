package repository

import (
	"context"
	"errors"
	"loadbalancer/internal/domain/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLReposytory struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *SQLReposytory {
	return &SQLReposytory{
		pool: pool,
	}
}

func (r *SQLReposytory) AddNewClient(ctx context.Context, client models.Client) (int64, error) {
	var (
		id  int64
		err error
	)
	err = pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		id, err = getClinetByIP(ctx, tx, client.IP)
		if err != nil {
			return err
		}
		//Если клиент уже есть, то меняем ему настройки
		if id != 0 {
			client.ID = id
			err = updateClientByID(ctx, tx, client)
			if err != nil {
				return err
			}
		} else {
			id, err = addClient(ctx, tx, client)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return id, nil
}

func getClinetByIP(ctx context.Context, tx pgx.Tx, ip string) (int64, error) {
	const (
		query = "SELECT id FROM clients WHERE ip = $1;"
	)
	var id int64
	err := tx.QueryRow(ctx, query, ip).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

func updateClientByID(ctx context.Context, tx pgx.Tx, client models.Client) error {
	const (
		query = "UPDATE clients SET capacity = $1, rate_per_sec = $2 WHERE id=$3;"
	)

	_, err := tx.Exec(ctx, query, client.Capacity, client.RatePerSec, client.ID)

	return err
}

func addClient(ctx context.Context, tx pgx.Tx, client models.Client) (int64, error) {
	const (
		query = "INSERT INTO clients(ip, capacity, rate_per_sec) VALUES ($1, $2, $3) RETURNING id;"
	)

	var id int64
	err := tx.QueryRow(ctx, query, client.IP, client.Capacity, client.RatePerSec).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *SQLReposytory) GetClinet(ctx context.Context, ip string) (*models.Client, error) {
	const (
		query = "SELECT id, ip, capacity, rate_per_sec FROM clients WHERE ip = $1;"
	)

	client := &models.Client{}
	err := r.pool.QueryRow(ctx, query, ip).Scan(&client.ID, &client.IP, &client.Capacity, &client.RatePerSec)
	if err != nil {
		return nil, err
	}

	return client, nil
}
