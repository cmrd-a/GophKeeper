package repository

import (
	"context"
	"time"

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

// User methods

// InsertUser inserts a new user with hashed password and returns the generated id.
func (r *Repository) InsertUser(ctx context.Context, login string, password []byte) (string, error) {
	var id string
	// password is stored as bytea in DB
	err := r.pool.QueryRow(ctx, `INSERT INTO "user" (login, password) VALUES ($1, $2) RETURNING id`, login, password).
		Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetUserByLogin returns id and hashed password for a given login.
func (r *Repository) GetUserByLogin(ctx context.Context, login string) (string, []byte, error) {
	var id string
	var pw []byte
	err := r.pool.QueryRow(ctx, `SELECT id, password FROM "user" WHERE login=$1`, login).Scan(&id, &pw)
	if err != nil {
		return "", nil, err
	}
	return id, pw, nil
}

// LoginPassword methods

func (r *Repository) GetLoginPasswords(ctx context.Context, userID string) ([]models.LoginPassword, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, login, password, created_at, updated_at 
         FROM login_password WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.LoginPassword
	for rows.Next() {
		var lp models.LoginPassword
		err := rows.Scan(&lp.ID, &lp.UserID, &lp.Login, &lp.Password, &lp.CreatedAt, &lp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, lp)
	}
	return result, rows.Err()
}

func (r *Repository) InsertLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO login_password (id, login, password, user_id, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6)`,
		lp.ID, lp.Login, lp.Password, lp.UserID, now, now)
	return err
}

func (r *Repository) UpdateLoginPassword(ctx context.Context, lp models.LoginPassword) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE login_password 
         SET login=$1, password=$2, updated_at=$3 
         WHERE id=$4 AND user_id=$5`,
		lp.Login, lp.Password, time.Now(), lp.ID, lp.UserID)
	return err
}

func (r *Repository) DeleteLoginPassword(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM login_password WHERE id=$1 AND user_id=$2",
		id, userID)
	return err
}

// TextData methods

func (r *Repository) GetTextData(ctx context.Context, userID string) ([]models.TextData, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, text, created_at, updated_at 
         FROM text_data WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.TextData
	for rows.Next() {
		var td models.TextData
		err := rows.Scan(&td.ID, &td.UserID, &td.Text, &td.CreatedAt, &td.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, td)
	}
	return result, rows.Err()
}

func (r *Repository) InsertTextData(ctx context.Context, td models.TextData) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO text_data (id, user_id, text, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5)`,
		td.ID, td.UserID, td.Text, now, now)
	return err
}

func (r *Repository) DeleteTextData(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM text_data WHERE id=$1 AND user_id=$2",
		id, userID)
	return err
}

// BinaryData methods

func (r *Repository) GetBinaryData(ctx context.Context, userID string) ([]models.BinaryData, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, data, created_at, updated_at 
         FROM binary_data WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.BinaryData
	for rows.Next() {
		var bd models.BinaryData
		err := rows.Scan(&bd.ID, &bd.UserID, &bd.Data, &bd.CreatedAt, &bd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, bd)
	}
	return result, rows.Err()
}

func (r *Repository) InsertBinaryData(ctx context.Context, bd models.BinaryData) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO binary_data (id, user_id, data, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5)`,
		bd.ID, bd.UserID, bd.Data, now, now)
	return err
}

func (r *Repository) DeleteBinaryData(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM binary_data WHERE id=$1 AND user_id=$2",
		id, userID)
	return err
}

// CardData methods

func (r *Repository) GetCardData(ctx context.Context, userID string) ([]models.CardData, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, number, cvv, holder, expires, created_at, updated_at 
         FROM card_data WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.CardData
	for rows.Next() {
		var cd models.CardData
		err := rows.Scan(&cd.ID, &cd.UserID, &cd.Number, &cd.CVV, &cd.Holder, &cd.Expires,
			&cd.CreatedAt, &cd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, cd)
	}
	return result, rows.Err()
}

func (r *Repository) InsertCardData(ctx context.Context, cd models.CardData) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO card_data (id, user_id, number, cvv, holder, expires, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		cd.ID, cd.UserID, cd.Number, cd.CVV, cd.Holder, cd.Expires, now, now)
	return err
}

func (r *Repository) DeleteCardData(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM card_data WHERE id=$1 AND user_id=$2",
		id, userID)
	return err
}

// Meta methods

func (r *Repository) GetMetaForItem(ctx context.Context, relationID string) ([]models.Meta, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, relation, name, data, created_at, updated_at 
         FROM meta WHERE relation = $1`, relationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Meta
	for rows.Next() {
		var m models.Meta
		err := rows.Scan(&m.ID, &m.Relation, &m.Name, &m.Data, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repository) InsertMeta(ctx context.Context, m models.Meta) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO meta (id, relation, name, data, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.Relation, m.Name, m.Data, now, now)
	return err
}

func (r *Repository) DeleteMeta(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM meta WHERE id=$1",
		id)
	return err
}
