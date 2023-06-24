package posts

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"database/sql"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
	"strconv"
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
			post.Created.UnixNano(),
		}

		var trueId bool
		if govalidator.IsInt(post.ThreadId) {
			log.Println("Entered validator scope")
			tId, _ := strconv.Atoi(post.ThreadId)
			err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where id=$1)`, tId).Scan(&trueId)
		} else {
			log.Println("Entered validator second scope")
			err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where lower(slug) = lower($1))`, post.ThreadId).Scan(&trueId)
		}
		if !trueId {
			err = utils.ErrNonExist
			return nil, err
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
			threadId = " (SELECT id FROM thread where lower(slug)=lower($3)) "
		}

		query := fmt.Sprintf(`INSERT INTO post(parent_id, author_id, message, thread_id, created, forum_id) 
		VALUES (%s, (SELECT id FROM users WHERE lower(nickname)=lower($1)), $2, %s, $4, (SELECT f.id FROM forum f JOIN thread t on f.id = t.forum_id WHERE t.id=%s))
		RETURNING id, forum_id`,
			parentId,
			threadId,
			threadId,
		)

		var forumId int

		err = tx.QueryRow(query, args...).Scan(&post.Id, &forumId)
		log.Println(err)
		if err != nil {
			errCode, ok := err.(pgx.PgError)
			if ok {
				switch errCode.Code {
				case pgerrcode.ForeignKeyViolation:
					return nil, utils.ErrConflict
				case pgerrcode.InvalidColumnReference:
					return nil, err
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

		if !govalidator.IsInt(post.ThreadId) {
			log.Println("In non-int branch")

			var id int
			err = tx.QueryRow(`SELECT id FROM thread where lower(slug)=lower($1)`, post.ThreadId).Scan(&id)
			var inThread bool
			log.Println("THread Id -- ", threadId, post.ThreadId, post.ParentId)

			if cond := post.ParentId; cond != 0 {
				err = tx.QueryRow(`SELECT EXISTS(SELECT from post WHERE (post.id=$1 AND post.thread_id=(SELECT id from thread where slug=$2)))`, post.ParentId, post.ThreadId).Scan(&inThread)
				log.Println(err)

				log.Printf("%+v", inThread)
			} else {
				inThread = true
			}

			if !inThread {
				err = utils.ErrConflict
				return nil, err
			}
			post.ThreadId = strconv.Itoa(id)
		} else {
			var inThread bool
			log.Println("THread Id -- ", threadId, post.ThreadId, post.ParentId)
			tID, _ := strconv.Atoi(post.ThreadId)
			if cond := post.ParentId; cond != 0 {
				err = tx.QueryRow(`SELECT EXISTS(SELECT from post WHERE (post.id=$1 AND post.thread_id=$2))`, post.ParentId, tID).Scan(&inThread)
				log.Println(err)

				log.Printf("%+v", inThread)
			} else {
				inThread = true
			}

			if !inThread {
				err = utils.ErrConflict
				return nil, err
			}

			log.Println("In int branch")
		}
		if err != nil {
			log.Println("error scanning THreadId: ", err)
		}

		log.Printf("%+v", post)

	}
	log.Println("Exited without error")
	return batch, nil
}

func (s *PostStorage) Details(id int) (*models.Post, error) {
	post := &models.Post{Id: id}

	nullableInt := sql.NullInt64{}
	var tempTime int64
	var tempId int
	err := s.db.QueryRow(
		`SELECT p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created, f.slug 
		FROM post p
		JOIN forum f on f.id = p.forum_id
		JOIN users u on u.id = p.author_id
		WHERE p.id=$1`,
		id,
	).Scan(
		&nullableInt,
		&post.Author,
		&post.Message,
		&post.Edited,
		&tempId,
		&tempTime,
		&post.ForumSlug,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}
	post.ThreadId = strconv.Itoa(tempId)
	post.Created = time.Unix(0, tempTime)
	post.ParentId = int(nullableInt.Int64)

	return post, nil
}

func (s *PostStorage) Update(id int, message string) (*models.Post, error) {
	post := &models.Post{Id: id, Message: message}

	var tempId int
	var tempTime int64
	var oldMsg string
	nullableInt := sql.NullInt64{}
	err := s.db.QueryRow(`SELECT p.message, p.parent_id, u.nickname, p.thread_id, p.created, f.slug FROM post p
                                                                               JOIN forum f on f.id = p.forum_id
                                                                               JOIN users u on u.id = p.author_id
                                                                               WHERE p.id=$1`, id).Scan(
		&oldMsg,
		&nullableInt,
		&post.Author,
		&tempId,
		&tempTime,
		&post.ForumSlug,
	)
	log.Println(err)
	post.Message = oldMsg
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}
	post.ParentId = int(nullableInt.Int64)
	post.ThreadId = strconv.Itoa(tempId)
	post.Created = time.Unix(0, tempTime)
	if oldMsg != message && message != "" {

		_, err = s.db.Exec(`UPDATE post SET message=$1, edited=true 
            WHERE id=$2`,
			message,
			id,
		)
		log.Println(err)
		if err != nil {
			return nil, err
		}
		post.Edited = true
		post.Message = message

	}
	return post, nil
}
