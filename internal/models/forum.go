package models

type Forum struct {
	Id          int    `json:"-"`
	Title       string `json:"title"`
	UserName    string `json:"user"`
	Slug        string `json:"slug"`
	ThreadCount uint64 `json:"threads"`
	PostCount   uint64 `json:"posts"`
}
