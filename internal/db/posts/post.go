package posts

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
	"time"
)

type PostStorage struct {
	db *pgx.ConnPool
}

func NewStorage(pool *pgx.ConnPool) *PostStorage {
	return &PostStorage{db: pool}
}

func (s *PostStorage) Insert(batch []*models.Post) ([]*models.Post, error) {
	tx, err := s.db.Begin()

	if err != nil {
		return nil, err
	}

	beginDate := time.Now()

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, post := range batch {
		post.Created = beginDate

		args := []interface{}{
			post.Author,
			post.Message,
			post.ThreadId,
			post.Created,
		}

		var parentId string
		if post.ParentId != 0 {
			parentId = "$5"
			args = append(args, post.ParentId)
		} else {
			parentId = "null"
		}

		var threadId string

		if govalidator.IsInt(post.ThreadId) {
			threadId = "$3"
		} else {
			threadId = " (SELECT id FROM thread where slug=$3) "
		}

		query := fmt.Sprintf(`INSERT INTO post(parent_id, author_id, message, thread_id, created, forum_id) 
		VALUES (%s, (SELECT id FROM users WHERE nickname=$1), $2, %s, $4, (SELECT f.id FROM forum f JOIN thread t on f.id = t.forum_id WHERE t.id=%s))
		RETURNING id, forum_id`,
			parentId,
			threadId,
			threadId,
		)

		var forumId int

		err = tx.QueryRow(query, args...).Scan(&post.Id, &forumId)

		if err != nil {
			errCode, ok := err.(pgx.PgError)
			if ok {
				switch errCode.Code {
				case pgerrcode.ForeignKeyViolation:
					return nil, utils.ErrConflict
				case pgerrcode.InvalidColumnReference:
					return nil, utils.ErrConflict
				case pgerrcode.NotNullViolation:
					return nil, utils.ErrNonExist
				}
			}
			log.Println("In tx: ", err)
			return nil, err
		}

		var forumSlug string

		err = tx.QueryRow(`SELECT slug FROM forum where id=$1`,
			forumId,
		).Scan(&forumSlug)

		post.ForumSlug = forumSlug

		if err != nil {
			log.Println("error scanning forumslug: ", err)
		}

	}
	log.Println("Exited without error")
	return batch, nil
}

func (s *PostStorage) Details(id int) (*models.Post, error) {
	post := &models.Post{Id: id}

	err := s.db.QueryRow(
		`SELECT p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created, f.slug 
		FROM post p
		JOIN forum f on f.id = p.forum_id
		JOIN users u on u.id = f.author_id
		WHERE id=$1`,
		id,
	).Scan(
		&post.ParentId,
		&post.Author,
		&post.Message,
		&post.Edited,
		&post.ThreadId,
		&post.Created,
		&post.ForumSlug,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	return post, nil
}

func (s *PostStorage) Update(id int, message string) (*models.Post, error) {
	post := &models.Post{Id: id, Message: message}

	err := s.db.QueryRow(`SELECT p.parent_id, u.nickname, p.thread_id, p.created, f.slug FROM post p
                                                                               JOIN forum f on f.id = p.forum_id
                                                                               JOIN users u on u.id = p.author_id
                                                                               WHERE id=$1`).Scan(
		&post.ParentId,
		&post.Author,
		&post.ThreadId,
		&post.Created,
		&post.ForumSlug,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	err = s.db.QueryRow(`UPDATE post SET message=$1, edited=true 
            WHERE id=$2`,
		message,
		id,
	).Scan()

	if err != nil {
		return nil, err
	}

	return post, nil
}
