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
		WHERE lower(nickname)=lower($1)`,
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

	tempUsr := &models.User{}

	err := u.db.QueryRow(`SELECT id, nickname, fullname, about, email FROM users where lower(nickname)=lower($1)`, user.Nickname).Scan(
		&tempUsr.Id,
		&tempUsr.Nickname,
		&tempUsr.Fullname,
		&tempUsr.About,
		&tempUsr.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	if user.Fullname == "" {
		user.Fullname = tempUsr.Fullname
	}

	if user.About == "" {
		user.About = tempUsr.About
	}

	if user.Email == "" {
		user.Email = tempUsr.Email
	}

	err = u.db.QueryRow(`UPDATE users SET fullname=$1, about=$2, email=$3 WHERE lower(nickname)=lower($4) RETURNING id`,
		user.Fullname,
		user.About,
		user.Email,
		user.Nickname,
	).Scan(&user.Id)

	//query := `UPDATE users SET `
	//
	//args := []interface{}{
	//	user.Nickname,
	//}

	//argCounter := 2

	//log.Printf("%+v", user)

	//	if user.Fullname != "" {
	//		query += fmt.Sprintf("fullname=$%d,", argCounter)
	//		args = append(args, user.Fullname)
	//		argCounter++
	//	}
	//
	//	if user.About != "" {
	//		query += fmt.Sprintf(" about=$%d,", argCounter)
	//		args = append(args, user.About)
	//		argCounter++
	//	}
	//
	//	if user.Email != "" {
	//		query += fmt.Sprintf(" email=$%d,", argCounter)
	//		args = append(args, user.Email)
	//		argCounter++
	//	}
	//
	//	if argCounter == 2 {
	//		err := u.db.QueryRow(`SELECT nickname, fullname, about, email FROM users WHERE nickname=$1`, user.Nickname).Scan(
	//			&user.Nickname,
	//			&user.Fullname,
	//			&user.About,
	//			&user.Email,
	//		)
	//		if err != nil {
	//			if err == pgx.ErrNoRows {
	//				return nil, utils.ErrNonExist
	//			}
	//			return nil, err
	//		}
	//		return user, nil
	//	}
	//
	//	query = strings.TrimRight(query, ",")
	//
	//	query += `
	//WHERE lower(nickname)=lower($1)
	//		RETURNING id`
	//
	//	err := u.db.QueryRow(
	//		query,
	//		args...,
	//	).Scan(&user.Id)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		if errCode, ok := err.(pgx.PgError); ok && errCode.Code == pgerrcode.UniqueViolation {
			return nil, utils.ErrConflict
		}
		return nil, err
	}

	return user, nil
}

func (u *UserStorage) GetByEmail(email string) (*models.User, error) {
	user := &models.User{Email: email}

	err := u.db.QueryRow(
		`SELECT id, nickname, fullname, about, email FROM users WHERE lower(email)=lower($1)`,
		email,
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
