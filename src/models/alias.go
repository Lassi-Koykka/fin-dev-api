package models

import "time"

type Alias struct {
	Name         string `gorm:"primaryKey"`
	KeywordName  string
	Keyword 	 Keyword
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
