package posts

import (
	"DBProject/internal/models"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jackc/pgx"
	"time"
)

type PostStorage struct {
	db *pgx.ConnPool
}

func NewStorage(pool *pgx.ConnPool) *PostStorage {
	return &PostStorage{db: pool}
}

func (s *PostStorage) Insert(batch []*models.Post) (*models.Post, error) {
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

		err := tx.QueryRow(query, args...).Scan(&post.Id, &forumId)
	}
}
