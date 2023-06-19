package db

import (
	"DBProject/internal/models"
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
		log.Println("DB InsertUser error: ", err)
		return nil, err
	}
	return user, nil
}
