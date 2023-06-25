package posts

import (
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
	"strconv"
)

func (s *PostStorage) BatchInsert(batch []*models.Post) ([]*models.Post, error) {
	tx, err := s.db.Begin()

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if len(batch) == 0 {
		return batch, nil
	}

	args := make([]interface{}, 0)

	forumIds := make([]int, 0)
	userIds := make([]int, 0)

	query := `INSERT INTO post(parent_id, author_id, message, thread_id, created, forum_id) values
                                                                                  `
	i := 0
	for _, post := range batch {
		var uId int
		var fId int
		err = s.db.QueryRow(`SELECT id from forum where lower(slug)=lower($1)`, post.ForumSlug).Scan(&fId)
		if err != nil {
			return nil, utils.ErrNonExist
		}
		err = s.db.QueryRow(`SELECT id from users where lower(nickname)=lower($1)`, post.Author).Scan(&uId)
		if err != nil {
			return nil, utils.ErrNonExist
		}

		forumIds = append(forumIds, fId)
		userIds = append(userIds, uId)

		inThread := true

		tId, _ := strconv.Atoi(post.ThreadId)

		if post.ParentId != 0 {
			var tempId int
			err = s.db.QueryRow(`SELECT thread_id from post where id=$1`, post.ParentId).Scan(&tempId)
			inThread = tempId == tId
		}

		if !inThread {
			return nil, utils.ErrConflict
		}
		query += fmt.Sprintf("(nullif($%d, 0), $%d, $%d, $%d, $%d, $%d),", i+1, i+2, i+3, i+4, i+5, i+6)
		args = append(args, post.ParentId, uId, post.Message, tId, post.Created.UnixNano(), fId)
		i += 6

	}

	query = query[:len(query)-1] + " RETURNING id"

	log.Println(query)

	rows, err := s.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	i = 0

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		log.Println("ERror ", err)
		batch[i].Id = id
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
		i++
	}

	err = s.updateForumUsers(forumIds, userIds)

	log.Println("First id: ", batch[0].Id)

	return batch, err

}

func (s *PostStorage) updateForumUsers(forums, users []int) error {
	query := `INSERT INTO forumusers(user_id, forum_id) VALUES `

	paramCounter := 0

	args := make([]interface{}, 0)

	for i := 0; i < len(forums); i++ {
		query += fmt.Sprintf("($%d, $%d),", paramCounter+1, paramCounter+2)
		args = append(args, users[i], forums[i])
		paramCounter += 2
	}
	query = query[:len(query)-1]
	query += "on conflict do nothing"

	_, err := s.db.Exec(query, args...)

	if err != nil {
		return err
	}

	return nil
}
