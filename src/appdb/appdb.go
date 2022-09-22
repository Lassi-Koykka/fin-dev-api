package appdb

import (
	"fmt"
	"strings"
	"github.com/lassi-koykka/fin-dev-api/src/postings"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	DB_NAME = "database.db"
)

type AppDB struct {
	Db *gorm.DB
}

type SearchTerms struct {
	Location string
	Company  string
	TechName string
	Exact    bool
}

func Instance() AppDB {
	db, err := gorm.Open(sqlite.Open(DB_NAME), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	g.Check(err)

	fmt.Println("Running auto migrations")
	db.AutoMigrate(&Posting{})
	db.AutoMigrate(&Keyword{})
	fmt.Println("Migration has completed")

	return AppDB{
		Db: db,
	}
}

func (appdb AppDB) UpsertPostingsAndPruneDangling(result []postings.Posting) {
	db := appdb.Db
	newPostingSlugs := []string{}
	newPostings := []Posting{}
	for _, p := range result {
		slug := strings.TrimSpace(strings.ToLower(p.Slug))
		newPostingSlugs = append(newPostingSlugs, slug)
		keywords := []*Keyword{}
		for _, kw := range p.Keywords {
			keywords = append(keywords, &Keyword{Name: kw})
		}

		newPostings = append(newPostings, Posting{
			Slug:       slug,
			Heading:    p.Heading,
			DatePosted: p.DatePosted,
			ImageUrl:   p.ImageUrl,
			Descr:      p.Descr,
			Location:   p.Location,
			Company:    p.Company,
			Keywords:   keywords,
		})
	}

	var dbUpsertedCount int64
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&newPostings).Count(&dbUpsertedCount)
	db.Save(&newPostings)
	fmt.Printf("Upserted %d postings ", len(newPostings))

	var postingsToDelete []Posting
	db.Not(&newPostingSlugs).Find(&postingsToDelete)
	postingsToDeleteSlugs := []string{}
	for _, v := range postingsToDelete {
		postingsToDeleteSlugs = append(postingsToDeleteSlugs, v.Slug)
	}

	if len(postingsToDelete) > 0 {
		var deletedPostings []Posting
		db.Clauses(clause.Returning{}).Delete(&deletedPostings, &postingsToDeleteSlugs)
		db.Select(clause.Associations).Where(&postingsToDeleteSlugs).Delete(&deletedPostings)

		fmt.Println("and Deleted", len(deletedPostings), "postings: ")
		for _, dp := range deletedPostings {
			fmt.Println("\t - ", dp.Slug)
		}
	}

	var dbPostingsCount int64
	db.Find(&[]Posting{}).Count(&dbPostingsCount)
	fmt.Println("Postings in db:", dbPostingsCount)
}

func searchInput(str string, exact bool) string {
	if !exact {
		return "%" + str + "%"
	}
	return str
}

func (appdb AppDB) GetPostings(searchTerms *SearchTerms) []postings.Posting {
	db := appdb.Db
	exact := searchTerms.Exact

	var queryResult []Posting
	arguments := []interface{}{}
	searchStringParts := []string{}
	if len(searchTerms.Location) > 0 {
		arguments = append(arguments, searchInput(searchTerms.Location, exact))
		searchStringParts = append(searchStringParts, "location LIKE ?")
	}
	if len(searchTerms.Company) > 0 {
		arguments = append(arguments, searchInput(searchTerms.Company, exact))
		searchStringParts = append(searchStringParts, "company LIKE ?")
	}
	if len(searchTerms.TechName) > 0 {
		arguments = append(arguments, searchInput(searchTerms.TechName, exact))
		searchStringParts = append(searchStringParts, "slug IN ( SELECT posting_slug FROM posting_keywords WHERE keyword_name LIKE ? )")
	}

	if len(searchStringParts) > 0 {
		searchString := strings.Join(searchStringParts, " AND ")
		db.Preload("Keywords").Where(searchString, arguments...).Find(&queryResult)
	} else {
		db.Preload("Keywords").Where("1 == 1").Find(&queryResult)
	}

	resultPostings := []postings.Posting{}
	for _, p := range queryResult {

		keywords := []string{}
		for _, kw := range p.Keywords {
			keywords = append(keywords, kw.Name)
		}

		resultPostings = append(resultPostings, postings.Posting{
			Slug:       p.Slug,
			Heading:    p.Heading,
			DatePosted: p.DatePosted,
			Url:        "https://duunitori.fi/tyopaikat/tyo/" + p.Slug,
			ImageUrl:   p.ImageUrl,
			Descr:      p.Descr,
			Location:   p.Location,
			Company:    p.Company,
			Keywords:   keywords,
		})
	}

	return resultPostings

}

func (db *AppDB) UpdateData(keywords []string) {
	fmt.Println("\n------------ UPDATING DATA ------------ ", g.TimeStamp(), "\n ")
	result := postings.FetchAndProcessPostings(keywords)
	db.UpsertPostingsAndPruneDangling(result)
	fmt.Println("\n------------ UPDATING DONE ------------\n ")
}
