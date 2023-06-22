package threads

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
)

type ThreadStorage struct {
	db *pgx.ConnPool
}

func New(pool *pgx.ConnPool) *ThreadStorage {
	return &ThreadStorage{
		db: pool,
	}
}

func (s *ThreadStorage) Insert(thread *models.Thread) (*models.Thread, error) {

	query := `INSERT INTO thread(author_id, message, title, forum_id, created`

	if thread.Slug != "" {
		query += `, slug`
	}

	query += `) 
			values(
			       (SELECT id FROM users where nickname=$1), 
			       $2, $3, (SELECT id FROM forum WHERE slug=$4), $5`

	if thread.Slug != "" {
		query += `, $6`
	}

	query += `)
			RETURNING id`

	var err error

	if thread.Slug != "" {
		err = s.db.QueryRow(
			query,
			thread.Author,
			thread.Message,
			thread.Title,
			thread.Forum,
			thread.Created,
			thread.Slug,
		).Scan(&thread.Id)
	} else {
		err = s.db.QueryRow(
			query,
			thread.Author,
			thread.Message,
			thread.Title,
			thread.Forum,
			thread.Created,
		).Scan(&thread.Id)
	}

	if err != nil {
		log.Println(err)
		errCode, ok := err.(pgx.PgError)
		if ok {
			switch errCode.Code {
			case pgerrcode.UniqueViolation:
				return nil, utils.ErrConflict
			case pgerrcode.NotNullViolation:
				return nil, utils.ErrNonExist
			}
		}
		return nil, err
	}

	return thread, nil
}

func (s *ThreadStorage) GetById(id int) (*models.Thread, error) {
	thread := &models.Thread{Id: id}

	err := s.db.QueryRow(
		`SELECT u.nickname, message, thread.title, f.title, thread.slug, created, vote_count FROM thread
		JOIN users u on u.id = thread.author_id
		JOIN forum f on thread.forum_id = f.id
        WHERE thread.id=$1`,
		id,
	).Scan(
		&thread.Author,
		&thread.Message,
		&thread.Forum,
		&thread.Slug,
		&thread.Created,
		&thread.Votes,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
	}

	return thread, nil
}

func (s *ThreadStorage) GetBySlug(slug string) (*models.Thread, error) {
	thread := &models.Thread{Slug: slug}

	err := s.db.QueryRow(
		`SELECT thread.id, u.nickname, message, thread.title, f.title, created, vote_count FROM thread
		JOIN users u on u.id = thread.author_id
		JOIN forum f on thread.forum_id = f.id
        WHERE thread.slug=$1`,
		slug,
	).Scan(
		&thread.Id,
		&thread.Author,
		&thread.Message,
		&thread.Forum,
		&thread.Created,
		&thread.Votes,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
	}

	return thread, nil
}
