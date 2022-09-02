package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/lassi-koykka/fin-dev-api/src/postparser"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	// "gorm.io/gorm/clause"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	DB_NAME = "database.db"
)

func ConnectAndMigrate() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(DB_NAME), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	g.Check(err)

	fmt.Println("Running auto migrations")
	db.AutoMigrate(&Posting{})
	db.AutoMigrate(&Keyword{})
	fmt.Println("Migration has completed")

	return db
}

func UpsertPostingsAndPruneDangling(db *gorm.DB, result postparser.ProcessingResult) {
	_, debug := os.LookupEnv("DEBUG")

	newPostings := []Posting{}
	for _, p := range result.Postings {
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
	}

	var oldPostings []Posting

	db.Find(&oldPostings)

	postsToDelete := []string{}
	for _, p := range oldPostings {
		found := false
		for _, np := range newPostings {
			if p.Slug == np.Slug {
				found = true
				break
			}
		}
		if !found {
			postsToDelete = append(postsToDelete, p.Slug)
		}
	}

	if len(postsToDelete) > 0 {
		fmt.Printf("Deleting %d removed postings \n", len(postsToDelete))
		db.Delete(&Posting{}, postsToDelete)
		if debug {
		}
		fmt.Println("Deleted postings: ",  strings.Join(postsToDelete, ", "))
	}

	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&newPostings)
	db.Save(&newPostings)
	fmt.Printf("Upserted %d postings\n", len(newPostings))
}
