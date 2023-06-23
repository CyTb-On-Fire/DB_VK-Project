package common

import "time"

type ListParams struct {
	Slug  string `form:"-"`
	Limit int    `form:"limit"`
	Since string `form:"since"`
	Desc  bool   `form:"desc"`
}

type ThreadListParams struct {
	Slug  string    `form:"-"`
	Limit int       `form:"limit"`
	Since time.Time `form:"since"`
	Desc  bool      `form:"desc"`
}

type FilterParams struct {
	ThreadSlug string `json:"slug"`
	Limit      int    `json:"limit"`
	Since      int    `json:"since"`
	Sort       string `json:"sort"`
	Desc       bool   `json:"desc"`
}

type Vote struct {
	Nickname   string `json:"nickname" binding:"required"`
	Voice      int    `json:"voice" binding:"required"`
	ThreadSlug string `json:"-"`
}

type DbStatus struct {
	User   int `json:"user"`
	Forum  int `json:"forum"`
	Thread int `json:"thread"`
	Post   int `json:"post"`
}

type PostViewParams struct {
	Id     int      `form:"-"`
	Params []string `form:"params"`
}
