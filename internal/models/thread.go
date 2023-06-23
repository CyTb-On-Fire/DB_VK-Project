package models

import "time"

type Thread struct {
	Id      int       `json:"id"`
	Title   string    `json:"title" binding:"required"`
	Author  string    `json:"author" binding:"required"`
	Forum   string    `json:"forum"`
	Message string    `json:"message" binding:"required"`
	Votes   int       `json:"votes"`
	Slug    string    `json:"slug"`
	Created time.Time `json:"created"`
}
