package appdb

import (
	"time"
)

type Posting struct {
	Slug       string `gorm:"primaryKey"`
	Heading    string
	DatePosted string
	ImageUrl   string
	Descr      string
	Location   string
	Company    string
	Keywords   []*Keyword `gorm:"many2many:posting_keywords;"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Keyword struct {
	Name      string     `gorm:"primaryKey"`
	Postings  []*Posting `gorm:"many2many:posting_keywords;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
