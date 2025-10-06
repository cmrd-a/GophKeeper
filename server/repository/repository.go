package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cmrd-a/GophKeeper/server/models"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	r := &Repository{pool: pool}
	return r, nil
}

func (r Repository) InsertUser(login string) error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var id string
	err = conn.QueryRow(context.Background(), "SELECT login FROM \"user\" WHERE login=$1", login).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(id)
	return nil
}

func (r Repository) InsertLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	_, err := r.pool.Exec(
		ctx,
		"INSERT INTO login_password (login, password, user_id) VALUES ($1, $2, $3)",
		lp.Login,
		lp.Password,
		lp.UserID,
	)
	return err
}

func (r Repository) UpdateLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	_, err := r.pool.Exec(
		ctx,
		"UPDATE login_password SET login=$1, password=$2 WHERE id=$3",
		lp.Login,
		lp.Password,
		lp.ID,
	)
	return err
}
