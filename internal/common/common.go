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
