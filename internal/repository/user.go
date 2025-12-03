package repository

import (
	"database/sql"

	"short-linker/internal/model"
)

type UserDataRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserDataRepository {
	return &UserDataRepository{
		db: db,
	}
}

func (r *UserDataRepository) Create(user model.User) (model.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.User{}, err
	}
	defer tx.Rollback()
	
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`

	var createdUser model.User
	err = tx.QueryRow(query, user.Name, user.Email, user.Password).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}

func (r *UserDataRepository) GetByEmail(email string) (model.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.User{}, err
	}
	defer tx.Rollback()
	
	const query = `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err = tx.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, err
	}

	return user, nil
}