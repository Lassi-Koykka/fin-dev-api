package main

import (
	"fmt"
	"time"

	"github.com/lassi-koykka/fin-dev-api/src/postparser"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"github.com/lassi-koykka/fin-dev-api/src/utils/fileutils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type Keyword struct {
	Name      string     `gorm:"primaryKey"`
	Postings  []*Posting `gorm:"many2many:posting_keywords;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func main() {

	// DB
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	g.Check(err)

	fmt.Println("Running migrations")
	db.AutoMigrate(&Posting{})
	db.AutoMigrate(&Keyword{})
	fmt.Println("Migrated")

	keywords := fileutils.ParseFileLines("keywords/technologies.txt")
	result := postparser.FetchAndProcessPosts(keywords)

	newPostings := []Posting{}

	for i, p := range result.Postings {
		keywords := []*Keyword{}
		for _, kw := range p.KeywordsFound {
			keywords = append(keywords, &Keyword{Name: kw})
		}

		newPostings = append(newPostings, Posting{
			Slug:       p.Slug,
			Heading:    p.Heading,
			DatePosted: p.DatePosted,
			ImageUrl:   p.ImageUrl,
			Descr:      p.Descr,
			Location:   p.Location,
			Company:    p.Company,
			Keywords:   keywords,
		})
		if i == 30 {
			break
		}
	}

	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&newPostings)
	db.Save(&newPostings)
	fmt.Println("Added Postings")

}
