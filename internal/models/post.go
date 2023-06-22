package models

import "time"

type Post struct {
	Id        int       `json:"id"`
	ParentId  int       `json:"parent"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Edited    bool      `json:"isEdited"`
	ForumSlug string    `json:"forum"`
	Created   time.Time `json:"created"`
	ThreadId  string    `json:"thread"`
}
