package user

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
)

type UserStorage struct {
	db *pgx.ConnPool
}

func NewUserStorage(pool *pgx.ConnPool) *UserStorage {
	return &UserStorage{
		db: pool,
	}
}

func (u *UserStorage) InsertUser(user *models.User) (*models.User, error) {
	err := u.db.QueryRow(
		`INSERT INTO users(nickname, fullname, about, email)  values 
                                                         ($1, $2, $3, $4)
                                                         RETURNING id`,
		user.Nickname,
		user.Fullname,
		user.About,
		user.Email,
	).Scan(&user.Id)

	if err != nil {
		if errCode, ok := err.(pgx.PgError); ok && errCode.Code == pgerrcode.UniqueViolation {
			return nil, utils.ErrConflict
		}
		log.Println("DB InsertUser error: ", err)
		return nil, err
	}
	return user, nil
}

func (u *UserStorage) GetByNickname(nickname string) (*models.User, error) {
	user := &models.User{}

	err := u.db.QueryRow(
		`SELECT id, nickname, fullname, about, email FROM users
		WHERE nickname=$1`,
		nickname,
	).Scan(
		&user.Id,
		&user.Nickname,
		&user.Fullname,
		&user.About,
		&user.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		return nil, err
	}

	return user, nil
}

func (u *UserStorage) UpdateUser(user *models.User) (*models.User, error) {
	err := u.db.QueryRow(
		`UPDATE users SET fullname=$1, about=$2, email=$3
		WHERE nickname=$4
		RETURNING id`,
		user.Fullname,
		user.About,
		user.Email,
		user.Nickname,
	).Scan(&user.Id)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) GetByEmail(email string) (*models.User, error) {
	user := &models.User{Email: email}

	err := u.db.QueryRow(
		`SELECT id, nickname, fullname, about FROM users WHERE email=$1`,
		email,
	).Scan(
		&user.Id,
		&user.Nickname,
		&user.Fullname,
		&user.About,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	return user, nil
}
