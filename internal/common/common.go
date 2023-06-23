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
