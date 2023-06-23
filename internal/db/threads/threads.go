package threads

import (
	"DBProject/internal/common"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
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
		&thread.Title,
		&thread.Forum,
		&thread.Slug,
		&thread.Created,
		&thread.Votes,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}

		return nil, err
	}

	log.Println(thread)

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

func (s *ThreadStorage) Update(thread *models.Thread) (*models.Thread, error) {
	err := s.db.QueryRow(
		`UPDATE thread SET title=$1, message=$2
		WHERE id=$3`,
		thread.Title,
		thread.Message,
		thread.Id,
	).Scan()

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	return thread, nil
}

func (s *ThreadStorage) GetPostsWithFlat(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 1)

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
	var exists bool

	if govalidator.IsInt(params.ThreadSlug) {
		threadId = "$1"
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	} else {
		threadId = " (SELECT id FROM thread where slug=$1) "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where slug=$1)`, params.ThreadSlug).Scan(&exists)
	}

	if !exists {
		return nil, utils.ErrNonExist
	}

	rows, err := s.db.Query(
		`SELECT p.id, p.parent_id, u.nickname, p.message, p.edited, p.thread_id, p.created, f.slug FROM post p
            JOIN users u on p.author_id = u.id
            JOIN forum f on f.id = p.forum_id
		WHERE p.thread_id=`+threadId+sinceStmt+`ORDER BY created `+order+`
		LIMIT $2`,
		args...,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		post := &models.Post{}

		err := rows.Scan(
			&post.Id,
			&post.ParentId,
			&post.Author,
			&post.Message,
			&post.Edited,
			&post.ThreadId,
			&post.Created,
			&post.ForumSlug,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) GetPostsWithTree(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 1)

	var threadId string

	var exists bool
	var err error

	if govalidator.IsInt(params.ThreadSlug) {
		threadId = " $1 "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	} else {
		threadId = " (select id from thread where slug=$1) "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where slug=$1)`, params.ThreadSlug).Scan(&exists)
	}

	if !exists {
		return nil, utils.ErrNonExist
	}

	var order string
	var comparison string

	if params.Desc {
		order = "desc"
		comparison = "<"
	} else {
		order = "asc"
		comparison = ">"
	}

	sinceStmt := ""

	args := []interface{}{
		params.ThreadSlug,
		params.Limit,
	}

	if params.Since != 0 {
		sinceStmt = " AND tree.id" + comparison + "$3\n"
		args = append(args, params.Since)
	}

	query := fmt.Sprintf(`WITH RECURSIVE tree(id, parent_id, author_id, message, edited, thread_id, created, forum_id, path, key1, key2) AS(
    (SELECT p1.id, p1.parent_id, p1.author_id, p1.message, p1.edited, p1.thread_id, p1.created, p1.forum_id, p1.path, p1.created as key1, ARRAY[]::timestamptz[] as key2 FROM post p1 WHERE parent_id IS NULL ORDER BY created asc)
    UNION ALL
    SELECT p2.id, p2.parent_id, p2.author_id, p2.message, p2.edited, p2.thread_id, p2.created, p2.forum_id, p2.path, tree.key1 as key1, tree.key2 || p2.created as key2 FROM post p2 INNER JOIN tree on tree.id=p2.parent_id
) SELECT tree.id, parent_id, u.nickname, message, edited, thread_id, created, f.slug, key1, key2
FROM tree
         JOIN users u on u.id = tree.author_id
         JOIN forum f on f.id = tree.forum_id
WHERE thread_id=%s
%s
ORDER BY key1 %s, key2
LIMIT $2`, threadId, sinceStmt, order)

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

		err := rows.Scan(
			&post.Id,
			&post.ParentId,
			&post.Author,
			&post.Message,
			&post.Edited,
			&post.ThreadId,
			&post.Created,
			&post.ForumSlug,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) GetPostsWithParentTree(params *common.FilterParams) ([]*models.Post, error) {
	posts := make([]*models.Post, 1)

	var exists bool
	var err error

	var threadId string

	if govalidator.IsInt(params.ThreadSlug) {
		threadId = " $1 "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, params.ThreadSlug).Scan(&exists)
	} else {
		threadId = " (select id from thread where slug=$1) "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where slug=$1)`, params.ThreadSlug).Scan(&exists)
	}

	if !exists {
		return nil, utils.ErrNonExist
	}

	var order string
	var comparison string

	if params.Desc {
		order = "desc"
		comparison = "<"
	} else {
		order = "asc"
		comparison = ">"
	}

	sinceStmt := ""

	args := []interface{}{
		params.ThreadSlug,
		params.Limit,
	}

	if params.Since != 0 {
		sinceStmt = " AND tree.id" + comparison + "$3\n"
		args = append(args, params.Since)
	}

	query := fmt.Sprintf(`WITH RECURSIVE tree(id, parent_id, author_id, message, edited, thread_id, created, forum_id, path, key1, key2) AS(
    (SELECT p1.id, p1.parent_id, p1.author_id, p1.message, p1.edited, p1.thread_id, p1.created, p1.forum_id, p1.path, p1.created as key1, ARRAY[]::timestamptz[] as key2 FROM post p1 WHERE parent_id IS NULL %s ORDER BY created %s LIMIT $2)
    UNION ALL
    SELECT p2.id, p2.parent_id, p2.author_id, p2.message, p2.edited, p2.thread_id, p2.created, p2.forum_id, p2.path, tree.key1 as key1, tree.key2 || p2.created as key2 FROM post p2 INNER JOIN tree on tree.id=p2.parent_id
) SELECT tree.id, parent_id, u.nickname, message, edited, thread_id, created, f.slug, key1, key2
FROM tree
         JOIN users u on u.id = tree.author_id
         JOIN forum f on f.id = tree.forum_id
WHERE thread_id=%s
ORDER BY key1 %s, key2
`, sinceStmt, order, threadId, order)

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

		err := rows.Scan(
			&post.Id,
			&post.ParentId,
			&post.Author,
			&post.Message,
			&post.Edited,
			&post.ThreadId,
			&post.Created,
			&post.ForumSlug,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		posts = append(posts, post)

	}

	return posts, nil
}

func (s *ThreadStorage) NewVote(vote *common.Vote) (*models.Thread, error) {
	thread := &models.Thread{Slug: vote.ThreadSlug}

	var exists bool
	var err error
	var threadId string

	if govalidator.IsInt(vote.ThreadSlug) {
		threadId = " $2 "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread WHERE id=$1)`, vote.ThreadSlug).Scan(&exists)
	} else {
		threadId = " (select id from thread where slug=$2) "
		err = s.db.QueryRow(`SELECT EXISTS(SELECT from thread where slug=$1)`, vote.ThreadSlug).Scan(&exists)
	}

	var value bool

	if vote.Voice == 1 {
		value = true
	} else {
		value = false
	}

	err = s.db.QueryRow(`INSERT INTO vote(user_id, thread_id, positive_voice)
values ((SELECT id from users where nickname=$1), `+threadId+`, $3)`,
		vote.Nickname,
		vote.ThreadSlug,
		value,
	).Scan()

	if err != nil {
		errCode, ok := err.(pgx.PgError)
		if ok {
			switch errCode.Code {
			case pgerrcode.UniqueViolation:

				var positive bool

				err = s.db.QueryRow(`SELECT positive_voice from vote WHERE user_id=$1 and thread_id=` + threadId).Scan(&positive)

				if value == positive {
					return nil, utils.ErrConflict
				}
				err = s.db.QueryRow(`UPDATE vote set positive_voice=$1 where user_id=$2 and thread_id=` + threadId).Scan()

			case pgerrcode.NotNullViolation:
				return nil, utils.ErrNonExist
			}
		}
	}
	err = s.db.QueryRow(
		`SELECT t.id, u.nickname, t.message, t.title, f.slug, t.slug, t.created, t.vote_count FROM thread t
			JOIN users u on t.author_id = u.id
			JOIN forum f on f.id = t.forum_id`).
		Scan(
			&thread.Id,
			&thread.Author,
			&thread.Message,
			&thread.Title,
			&thread.Forum,
			&thread.Slug,
			&thread.Created,
			&thread.Votes,
		)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}
	return thread, nil
}
