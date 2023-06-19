package db

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
)

type ForumStorage struct {
	db *pgx.ConnPool
}

func NewForumStorage(pool *pgx.ConnPool) *ForumStorage {
	return &ForumStorage{
		db: pool,
	}
}

func (f *ForumStorage) InsertForum(forum *models.Forum) (*models.Forum, error) {
	var userId int
	err := f.db.QueryRow(
		`SELECT id FROM users where nickname=$1`,
		forum.UserName,
	).Scan(&userId)

	if err != nil {
		if errCode, ok := err.(pgx.PgError); ok && errCode == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		log.Println("DB InsertForum error: ", err)
		return nil, err
	}

	err = f.db.QueryRow(
		`INSERT INTO forum(title, author_id, slug) values
                                                                 ($1, $2, $3)
                                                                 RETURNING id`,
		forum.Title,
		userId,
		forum.Slug,
	).Scan(&forum.Id)

	if err != nil {
		if errCode, ok := err.(pgx.PgError); ok && errCode.Code == pgerrcode.UniqueViolation {
			return nil, utils.ErrConflict
		}
		log.Println("DB InsertForum error: ", err)
		return nil, err
	}
	return forum, nil
}

func (f *ForumStorage) GetBySlug(slug string) (*models.Forum, error) {
	forum := &models.Forum{}

	err := f.db.QueryRow(
		`SELECT u.nickname, f.slug, f.title, f.post_count, f.thread_count 
			FROM forum f
			JOIN users u on u.id = f.author_id
			WHERE f.slug=$1`,
		slug,
	).Scan(
		&forum.UserName,
		&forum.Slug,
		&forum.Title,
		&forum.PostCount,
		&forum.ThreadCount,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		return nil, err
	}
	return forum, nil
}
