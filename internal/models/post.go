package models

import "time"

type Post struct {
	Id        int       `json:"id"`
	ParentId  int       `json:"parent" binding:"required"`
	Author    string    `json:"author" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Edited    bool      `json:"isEdited"`
	ForumSlug string    `json:"forum"`
	Created   time.Time `json:"created"`
	ThreadId  string    `json:"thread"`
}

type ProxyPost struct {
	Id        int       `json:"id"`
	ParentId  int       `json:"parent,omitempty" binding:"required"`
	Author    string    `json:"author" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Edited    bool      `json:"isEdited,omitempty"`
	ForumSlug string    `json:"forum"`
	Created   time.Time `json:"created"`
	ThreadId  int       `json:"thread"`
}
