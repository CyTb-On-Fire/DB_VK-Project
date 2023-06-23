package forum

import (
	"DBProject/internal/common"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
	"time"
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
		`SELECT id, nickname FROM users where lower(nickname)=lower($1)`,
		forum.UserName,
	).Scan(&userId, &forum.UserName)

	if err != nil {
		if err == pgx.ErrNoRows {
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
			WHERE lower(f.slug)=lower($1)`,
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

func (f *ForumStorage) GetUsers(params *common.ListParams) ([]*models.User, error) {
	users := make([]*models.User, 0)

	if params.Limit == 0 {
		params.Limit = 100
	}

	var order string
	var comparsion string
	if params.Desc {
		comparsion = "<"
		order = "desc"
	} else {
		order = "asc"
		comparsion = ">"
	}

	log.Println(`SELECT DISTINCT u.id, u.nickname collate "C", u.fullname, u.about, u.email FROM users u
		RIGHT JOIN (SELECT id, author_id FROM thread WHERE (SELECT id from forum f where lower(f.slug)=lower($1)) = forum_id) t on t.author_id = u.id
		RIGHT JOIN (SELECT id, author_id FROM post WHERE (SELECT id from forum f where lower(f.slug)=lower($1)) = forum_id) p on u.id = p.author_id
	   WHERE nickname collate "C" ` + comparsion + ` $3 collate "C"
	   ORDER BY nickname collate "C"
	   ` + order +
		`
		LIMIT $2`)

	var forumId int

	err := f.db.QueryRow(`SELECT id from forum where lower(slug)=lower($1)`, params.Slug).Scan(&forumId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		return nil, err
	}

	args := []interface{}{
		forumId,
		params.Limit,
	}

	sinceFilter := ""

	if params.Since != "" {
		sinceFilter = ` AND nickname collate nickname_case_insensitive ` + comparsion + ` $3 collate nickname_case_insensitive
`
		args = append(args, params.Since)
	}

	rows, err := f.db.Query(
		`SELECT u.id, u.nickname, u.fullname, u.about, u.email FROM users u
		LEFT JOIN (SELECT id, author_id FROM thread WHERE $1 = forum_id) t on t.author_id = u.id
		LEFT JOIN (SELECT id, author_id FROM post WHERE $1 = forum_id) p on u.id = p.author_id
        WHERE (t.id is not null or p.id is not null)
		`+sinceFilter+`
        ORDER BY lower(nickname) 
        `+order+
			`
		LIMIT $2`,
		args...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		err = rows.Scan(
			&user.Id,
			&user.Nickname,
			&user.Fullname,
			&user.About,
			&user.Email,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil

}

func (f *ForumStorage) GetThreads(params *common.ThreadListParams) ([]*models.Thread, error) {
	threads := make([]*models.Thread, 0)

	var forumId int

	err := f.db.QueryRow(`SELECT id FROM forum where lower(slug)=lower($1)`, params.Slug).Scan(&forumId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, utils.ErrNonExist
		}
		log.Println(err)
		return nil, err
	}

	var order string
	var comparsion string

	if params.Desc {

		order = "desc"
		comparsion = "<="
	} else {

		order = "asc"
		comparsion = ">="
	}

	sinceStmt := ""

	args := []interface{}{
		forumId,
		params.Limit,
	}

	if !params.Since.Equal(time.Time{}) {
		sinceStmt =
			` AND t.created` + comparsion + ` $3`
		args = append(args, params.Since.UnixNano())
	}

	log.Println(forumId,
		params.Since,
		params.Limit,
	)

	log.Println(`SELECT t.id, u.nickname, t.message, t.title, t.created, t.vote_count, t.slug FROM thread t
	JOIN users u on u.id = t.author_id
    WHERE t.forum_id=$1` + sinceStmt + `
    ORDER BY t.created ` + order + `
	LIMIT $2`)

	rows, err := f.db.Query(`SELECT t.id, u.nickname, t.message, t.title, t.created, t.vote_count, t.slug, f.slug FROM thread t
	JOIN users u on u.id = t.author_id
                                                                             JOIN forum f on f.id = t.forum_id
    WHERE t.forum_id=$1`+sinceStmt+`
    ORDER BY t.created `+order+`
	LIMIT $2`,
		args...,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Scan()

	var tempInt int64

	for rows.Next() {
		thread := &models.Thread{Forum: params.Slug}
		err = rows.Scan(
			&thread.Id,
			&thread.Author,
			&thread.Message,
			&thread.Title,
			&tempInt,
			&thread.Votes,
			&thread.Slug,
			&thread.Forum,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		thread.Created = time.Unix(0, tempInt)

		threads = append(threads, thread)
	}

	return threads, nil
}
