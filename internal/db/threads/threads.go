package threads

import (
	"DBProject/internal/common"
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

type ThreadStorage struct {
	db *pgx.ConnPool
}

func New(pool *pgx.ConnPool) *ThreadStorage {
	return &ThreadStorage{
		db: pool,
	}
}

func (s *ThreadStorage) Insert(thread *models.Thread, forum *models.Forum) (*models.Thread, error) {

	query := `INSERT INTO thread(author_id, message, title, forum_id, created`

	if thread.Slug != "" {
		query += `, slug`
	}

	query += `) 
			values(
			       (SELECT id FROM users where lower(nickname)=lower($1)), 
			       $2, $3, $4, $5`

	if thread.Slug != "" {
		query += `, $6`
	}

	log.Println("forum id: ", forum.Id)

	query += `)
			RETURNING id`

	var err error

	log.Println(query)

	if thread.Slug != "" {
		err = s.db.QueryRow(
			query,
			thread.Author,
			thread.Message,
			thread.Title,
			forum.Id,
			thread.Created.UnixNano(),
			thread.Slug,
		).Scan(&thread.Id)
	} else {
		err = s.db.QueryRow(
			query,
			thread.Author,
			thread.Message,
			thread.Title,
			forum.Id,
			thread.Created.UnixNano(),
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

	thread.Forum = forum.Slug
	//err = s.db.QueryRow(`SELECT slug from forum WHERE lower(slug)=lower($1)`, thread.Forum).Scan(&thread.Forum)

	return thread, nil
}

func (s *ThreadStorage) GetById(id int) (*models.Thread, error) {
	thread := &models.Thread{Id: id}

	var tempTime int64
	var nullableStr sql.NullString
	err := s.db.QueryRow(
		`SELECT u.nickname, message, thread.title, f.slug, thread.slug, created, vote_count FROM thread
		JOIN users u on u.id = thread.author_id
		JOIN forum f on thread.forum_id = f.id
        WHERE thread.id=$1`,
		id,
	).Scan(
		&thread.Author,
		&thread.Message,
		&thread.Title,
		&thread.Forum,
		&nullableStr,
		&tempTime,
		&thread.Votes,
	)

	thread.Slug = nullableStr.String

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		return nil, err
	}

	thread.Created = time.Unix(0, tempTime)

	log.Println(thread)

	return thread, nil
}

func (s *ThreadStorage) GetBySlug(slug string) (*models.Thread, error) {
	thread := &models.Thread{Slug: slug}

	log.Println(slug)

	var tempTime int64

	err := s.db.QueryRow(
		`SELECT t.id, u.nickname, t.message, t.title, f.slug, t.created, t.vote_count, t.slug FROM thread t 
		JOIN users u on u.id = t.author_id
		JOIN forum f on t.forum_id = f.id
        WHERE lower(t.slug)=lower($1)`,
		slug,
	).Scan(
		&thread.Id,
		&thread.Author,
		&thread.Message,
		&thread.Title,
		&thread.Forum,
		&tempTime,
		&thread.Votes,
		&thread.Slug,
	)

	thread.Created = time.Unix(0, tempTime)

	if err != nil {
		log.Println(err)
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
	}

	log.Println(thread)

	return thread, nil
}

func (s *ThreadStorage) Update(thread *models.Thread) (*models.Thread, error) {
	_, err := s.db.Exec(
		`UPDATE thread SET title=$1, message=$2
		WHERE id=$3`,
		thread.Title,
		thread.Message,
		thread.Id,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	return thread, nil
}

func (s *ThreadStorage) GetPostsWithFlat(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 0)

	sinceStmt := ""

	var order string
	var comparison string

	if params.Desc {
		order = "desc"
		comparison = "<"
	} else {
		order = "asc"
		comparison = ">"
	}

	args := []interface{}{
		params.ThreadSlug,
		params.Limit,
	}

	if params.Since != 0 {
		sinceStmt = `
		AND p.id` + comparison + `$3
`
		args = append(args, params.Since)
	}

	var threadId string

	var err error
	//var exists bool

	if govalidator.IsInt(params.ThreadSlug) {
		threadId = "$1"
		//err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	} else {
		threadId = " (SELECT id FROM thread where lower(slug)=lower($1)) "
		//err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where lower(slug)=lower($1))`, params.ThreadSlug).Scan(&exists)
	}

	//if !exists {
	//	return nil, utils.ErrNonExist
	//}

	log.Println(`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created, f.slug FROM post p
            JOIN users u on p.author_id = u.id
            JOIN forum f on f.id = p.forum_id
		WHERE p.thread_id=` + threadId + sinceStmt + `ORDER BY created ` + order + `, p.id ` + order + `
		LIMIT $2;`)

	rows, err := s.db.Query(
		`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created FROM post p
            JOIN users u on p.author_id = u.id
		WHERE p.thread_id=`+threadId+sinceStmt+` ORDER BY created `+order+`, p.id `+order+`
		LIMIT $2;`,
		args...,
	)

	if err != nil {
		log.Println("AFter complex query", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		nullableInt := sql.NullInt64{}

		var tempId int

		var tempTime int64

		err := rows.Scan(
			&post.Id,
			&nullableInt,
			&post.Author,
			&post.Message,
			&post.Edited,
			&tempId,
			&tempTime,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		post.Created = time.Unix(0, tempTime)
		post.ThreadId = strconv.Itoa(tempId)
		post.ParentId = int(nullableInt.Int64)
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) GetPostsWithTree(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 0)

	//var threadId string

	//var exists bool
	var err error

	//if govalidator.IsInt(params.ThreadSlug) {
	//	//threadId = " $1 "
	//	err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	//} else {
	//	//threadId = " (select id from thread where lower(slug)=lower($1)) "
	//	err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where lower(slug)=lower($1))`, params.ThreadSlug).Scan(&exists)
	//}
	//
	//if !exists {
	//	return nil, utils.ErrNonExist
	//}

	var order string
	var comparison string

	var tId int
	if !govalidator.IsInt(params.ThreadSlug) {
		err = s.db.QueryRow(`SELECT id from thread where lower(slug)=lower($1)`, params.ThreadSlug).Scan(&tId)
	} else {
		tId, _ = strconv.Atoi(params.ThreadSlug)
	}
	if params.Desc {
		order = "desc"
		comparison = "<"
	} else {
		order = "asc"
		comparison = ">"
	}

	sinceStmt := ""

	args := []interface{}{
		tId,
		params.Limit,
	}

	if params.Since != 0 {
		sinceStmt = " AND p.path" + comparison + "(select path from post where id=$3)\n"
		args = append(args, params.Since)
	}

	query := fmt.Sprintf(`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created FROM post p
			JOIN users u on p.author_id = u.id
-- 			JOIN forum f on f.id = p.forum_id
			WHERE p.thread_id=$1
			%s
			ORDER BY p.path %s
			LIMIT $2`, sinceStmt, order)

	log.Println(query)

	//	oldQuery := fmt.Sprintf(`WITH RECURSIVE tree(id, parent_id, author_id, message, edited, thread_id, created, forum_id, path, key1, key2) AS(
	//    (SELECT p1.id, p1.parent_id, p1.author_id, p1.message, p1.edited, p1.thread_id, p1.created, p1.forum_id, p1.path, p1.created as key1, ARRAY[]::bigint[] as key2 FROM post p1 WHERE parent_id IS NULL ORDER BY created asc)
	//    UNION ALL
	//    SELECT p2.id, p2.parent_id, p2.author_id, p2.message, p2.edited, p2.thread_id, p2.created, p2.forum_id, p2.path, tree.key1 as key1, tree.key2 || p2.created as key2 FROM post p2 INNER JOIN tree on tree.id=p2.parent_id
	//) SELECT tree.id, parent_id, u.nickname, message, edited, thread_id, created, f.slug
	//FROM tree
	//         JOIN users u on u.id = tree.author_id
	//         JOIN forum f on f.id = tree.forum_id
	//WHERE thread_id=%s
	//%s
	//ORDER BY key1 %s, key2, tree.id
	//LIMIT $2`, threadId, sinceStmt, order)

	//_ = oldQuery

	rows, err := s.db.Query(
		query,
		args...,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		nullableInt := sql.NullInt64{}
		var tempTime int64
		var tempId int
		err := rows.Scan(
			&post.Id,
			&nullableInt,
			&post.Author,
			&post.Message,
			&post.Edited,
			&tempId,
			&tempTime,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		post.Created = time.Unix(0, tempTime)
		post.ThreadId = strconv.Itoa(tempId)
		post.ParentId = int(nullableInt.Int64)
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) GetPostsWithParentTree(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 0)

	//var exists bool
	var err error
	var tId int

	if govalidator.IsInt(params.ThreadSlug) {
		tId, _ = strconv.Atoi(params.ThreadSlug)
		//err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	} else {
		err = s.db.QueryRow(`SELECT id from thread where lower(slug)=lower($1)`, params.ThreadSlug).Scan(&tId)
		if err != nil {
			return nil, err
		}
		//err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where slug=$1)`, params.ThreadSlug).Scan(&exists)
	}

	//if !exists {
	//	return nil, utils.ErrNonExist
	//}

	var comparison string
	var order string
	var outerOrder string

	args := []interface{}{
		tId,
		params.Limit,
	}

	if params.Desc {
		order = "desc"
		comparison = "<"
		outerOrder = "ORDER BY path[1] desc, path"
	} else {
		order = "asc"
		comparison = ">"
		outerOrder = "ORDER BY path"
	}

	var sinceStmt string

	if params.Since > 0 {
		sinceStmt = `AND path[1]` + comparison + `(SELECT path[1] FROM post WHERE id = $3)`
		args = append(args, params.Since)
	}
	baseQuery := fmt.Sprintf(`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created FROM post p
--                                                                                   JOIN forum f on f.id = p.forum_id
                                                                                  JOIN users u on u.id = p.author_id
					WHERE p.path[1] IN (SELECT id FROM post WHERE thread_id = $1 AND parent_id IS NULL 
					        %s
					        ORDER BY id %s LIMIT $2)
					        %s`,
		sinceStmt,
		order,
		outerOrder,
	)

	//query := fmt.Sprintf(`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, f.slug, p.thread_id, p.created FROM post p
	//                                                                              JOIN forum f on f.id = p.forum_id
	//                                                                              JOIN users u on u.id = f.author_id
	//				WHERE p.path[1] IN
	//					(SELECT id FROM post p1 WHERE p1.thread_id = $1
	//					                          %s
	//					                          AND p1.parent_id IS NULL ORDER BY id %s LIMIT $2)
	//				ORDER BY path[1] %s, path ASC, id ASC`, sinceStmt, order, order)

	rows, err := s.db.Query(
		baseQuery,
		args...,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		nullableInt := sql.NullInt64{}
		var tempTime int64
		var tempId int
		err := rows.Scan(
			&post.Id,
			&nullableInt,
			&post.Author,
			&post.Message,
			&post.Edited,
			&tempId,
			&tempTime,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		post.Created = time.Unix(0, tempTime)
		post.ThreadId = strconv.Itoa(tempId)
		post.ParentId = int(nullableInt.Int64)
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) NewVote(vote *common.Vote) (*models.Thread, error) {
	thread := &models.Thread{Slug: vote.ThreadSlug}

	var exists bool
	var err error
	var threadId int
	var slugisInt bool

	if govalidator.IsInt(vote.ThreadSlug) {
		slugisInt = true
		threadId, _ = strconv.Atoi(vote.ThreadSlug)
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, vote.ThreadSlug).Scan(&exists)
	} else {
		err = s.db.QueryRow(`SELECT id FROM thread where lower(slug)=lower($1)`, thread.Slug).Scan(&threadId)
		log.Println(threadId)
		if err != nil {
			log.Println("err trying get threadId: ", err)
		}
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where lower(slug)=lower($1))`, vote.ThreadSlug).Scan(&exists)
	}

	var value bool

	if vote.Voice == 1 {
		value = true
	} else {
		value = false
	}

	err = s.db.QueryRow(`INSERT INTO vote(user_id, thread_id, positive_voice)
values ((SELECT id from users where lower(nickname)=lower($1)), $2, $3) RETURNING thread_id`,
		vote.Nickname,
		threadId,
		value,
	).Scan(&thread.Id)

	log.Println(err)

	log.Println(thread.Id)

	if err != nil {

		log.Println(err)
		errCode, ok := err.(pgx.PgError)
		if ok {
			switch errCode.Code {
			case pgerrcode.UniqueViolation:

				var positive bool

				err = s.db.QueryRow(`SELECT positive_voice from vote WHERE user_id=(SELECT id from users where lower(nickname)=lower($1)) and thread_id=$2`, vote.Nickname, threadId).Scan(&positive)

				log.Println("err before old voice is: ", err)

				log.Println("old voice is:", positive)

				if value != positive {
					_, err = s.db.Exec(`UPDATE vote set positive_voice=$1 where user_id=(SELECT id from users where lower(nickname)=lower($2)) and thread_id=$3`, value, vote.Nickname, threadId)

					if err != nil {
						log.Println(err)
					}
				}

			case pgerrcode.NotNullViolation:
				log.Println(err)
				return nil, utils.ErrNonExist
			}
		}
	}

	var tempTime int64

	if !slugisInt {
		err = s.db.QueryRow(
			`SELECT t.id, u.nickname, t.message, t.title, f.slug, t.created, t.vote_count, t.slug FROM thread t
			JOIN users u on t.author_id = u.id
			JOIN forum f on f.id = t.forum_id
			WHERE lower(t.slug)=lower($1)`, thread.Slug).
			Scan(
				&thread.Id,
				&thread.Author,
				&thread.Message,
				&thread.Title,
				&thread.Forum,
				&tempTime,
				&thread.Votes,
				&thread.Slug,
			)
	} else {
		err = s.db.QueryRow(
			`SELECT t.id, u.nickname, t.message, t.title, f.slug, t.created, t.vote_count, t.slug FROM thread t
			JOIN users u on t.author_id = u.id
			JOIN forum f on f.id = t.forum_id
			WHERE t.id=$1`, threadId).
			Scan(
				&thread.Id,
				&thread.Author,
				&thread.Message,
				&thread.Title,
				&thread.Forum,
				&tempTime,
				&thread.Votes,
				&thread.Slug,
			)
	}

	thread.Created = time.Unix(0, tempTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		thread.Slug = strconv.Itoa(thread.Id)
	}
	return thread, nil
}
