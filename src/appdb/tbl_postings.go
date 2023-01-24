package appdb

import (
	"fmt"
	"strings"

	models "github.com/lassi-koykka/fin-dev-api/src/models"
	"gorm.io/gorm/clause"
)

// GET
func (appdb *AppDB) GetPostings(searchTerms *SearchTerms) []models.Posting {
	db := appdb.Db
	exact := searchTerms.Exact

	var queryResult []models.Posting
	arguments := []interface{}{}
	searchStringParts := []string{}
	if len(searchTerms.Location) >  0 {
		arguments = append(arguments, searchInput(searchTerms.Location, exact))
		searchStringParts = append(searchStringParts, "location LIKE ?")
	}
	if len(searchTerms.Company) > 0 {
		arguments = append(arguments, searchInput(searchTerms.Company, exact))
		searchStringParts = append(searchStringParts, "company LIKE ?")
	}
	if len(searchTerms.Query) > 0 {
		arguments = append(arguments, searchInput(searchTerms.Query, exact))
		arguments = append(arguments, searchInput(searchTerms.Query, exact))
		searchStringParts = append(searchStringParts, "(heading LIKE ? OR descr LIKE ?)")
	}

	if len(searchStringParts) > 0 {
		searchString := strings.Join(searchStringParts, " AND ")
		db.Preload("Keywords").Where(searchString, arguments...).Find(&queryResult)
	} else {
		db.Preload("Keywords").Where("1 == 1").Find(&queryResult)
	}

	resultPostings := []models.Posting{}
	for _, p := range queryResult {

		keywords := []string{}
		for _, kw := range p.Keywords {
			keywords = append(keywords, kw.Name)
		}

		resultPostings = append(resultPostings, models.Posting{
			Slug:       p.Slug,
			Heading:    p.Heading,
			DatePosted: p.DatePosted,
			Url:        "https://duunitori.fi/tyopaikat/tyo/" + p.Slug,
			ImageUrl:   p.ImageUrl,
			Descr:      p.Descr,
			Location:   p.Location,
			Company:    p.Company,
			Keywords: 	p.Keywords,
			KeywordsFound:   keywords,
		})
	}
	return resultPostings
}

//DELETE 
func (appdb *AppDB) DeletePostings(postingsToDeleteSlugs []string) []models.Posting {
	if len(postingsToDeleteSlugs) < 1 {
		return []models.Posting{}
	}

	db := appdb.Db
	var deletedPostings []models.Posting
	db.Clauses(clause.Returning{}).Delete(&deletedPostings, &postingsToDeleteSlugs)
	db.Select(clause.Associations).Where(&postingsToDeleteSlugs).Delete(&deletedPostings)

	return deletedPostings
}

// UPSERT
func (appdb *AppDB) UpsertPostings(newPostings []models.Posting) []models.Posting {
	db := appdb.Db
	var dbUpsertedCount int64
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&newPostings).Count(&dbUpsertedCount)
	db.Save(&newPostings)
	return newPostings
}

// UPSERT
func (appdb *AppDB) UpsertAndPrunePostings(postings []models.Posting) {
	db := appdb.Db
	newPostingSlugs := []string{}
	newPostings := []models.Posting{}
	for _, p := range postings {
		slug := strings.TrimSpace(strings.ToLower(p.Slug))
		newPostingSlugs = append(newPostingSlugs, slug)

		newPostings = append(newPostings, models.ToPosting(p))
	}

	upsertedPostings := appdb.UpsertPostings(newPostings)
	fmt.Printf("Upserted %d postings ", len(upsertedPostings))

	var postingsToDelete []models.Posting
	db.Not(&newPostingSlugs).Find(&postingsToDelete)
	postingsToDeleteSlugs := []string{}
	for _, v := range postingsToDelete {
		postingsToDeleteSlugs = append(postingsToDeleteSlugs, v.Slug)
	}

	deletedPostings := appdb.DeletePostings(postingsToDeleteSlugs)
	fmt.Println("and Deleted", len(deletedPostings), "postings: ")
	for _, dp := range deletedPostings {
		fmt.Println("\t - ", dp.Slug)
	}

	var dbPostingsCount int64
	db.Find(&[]models.Posting{}).Count(&dbPostingsCount)
	fmt.Println("Postings in db:", dbPostingsCount)
}
